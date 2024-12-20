package commands

import (
	"encoding/json"
	"os"
	"sync"

	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	fetchUrlsJson = false
)

type fetchUrlsObject struct {
	Name   string            `json:"name"`
	Oid    string            `json:"oid"`
	Size   int64             `json:"size"`
	Href   string            `json:"href"`
	Header map[string]string `json:"headers,omitempty"`
}

func getPointer(gitfilter *lfs.GitFilter, file string) (*lfs.Pointer, error) {
	var pointer, err = lfs.DecodePointerFromFile(file)
	if err == nil {
		return pointer, nil
	}
	openedFile, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer openedFile.Close()
	stat, err := openedFile.Stat()
	if err != nil {
		return nil, err
	}
	cleaned, err := gitfilter.Clean(openedFile, file, stat.Size(), nil)
	if err != nil {
		return nil, err
	}
	defer cleaned.Teardown()
	return cleaned.Pointer, nil
}

func initData(args []string) ([]*fetchUrlsObject, []*lfs.Pointer) {
	var items = make([]*fetchUrlsObject, 0, len(args))
	var pointers = make([]*lfs.Pointer, 0, len(args))
	gitfilter := lfs.NewGitFilter(cfg)
	for _, file := range args {
		pointer, err := getPointer(gitfilter, file)
		if err != nil {
			ExitWithError(err)
		}
		items = append(items, &fetchUrlsObject{Name: file, Oid: pointer.Oid, Size: pointer.Size})
		pointers = append(pointers, pointer)
	}
	return items, pointers
}

func groupByOid(items []*fetchUrlsObject) map[string][]*fetchUrlsObject {
	var argMap = make(map[string][]*fetchUrlsObject)
	for _, item := range items {
		items, exists := argMap[item.Oid]
		if !exists {
			items = make([]*fetchUrlsObject, 0, 1)
		}
		argMap[item.Oid] = append(items, item)
	}
	return argMap
}

func fillData(wait *sync.WaitGroup, data []*fetchUrlsObject, transfers <-chan *tq.Transfer) {
	byOid := groupByOid(data)
	for t := range transfers {
		a := t.Actions["download"]
		items, exists := byOid[t.Oid]
		if !exists {
			continue
		}
		for _, item := range items {
			item.Href = a.Href
			item.Header = a.Header
		}
		delete(byOid, t.Oid)
	}
	// any remaining items are missing, mark by removing the name
	for _, items := range byOid {
		for _, item := range items {
			item.Name = ""
		}
	}
	wait.Done()
}

func checkForErrors(queue *tq.TransferQueue) bool {
	var ok = true
	for _, e := range queue.Errors() {
		ok = false
		FullError(e)
	}
	return ok
}

func populateData(args []string) ([]*fetchUrlsObject, bool) {
	data, pointers := initData(args)
	var queue = newDownloadCheckQueue(getTransferManifestOperationRemote("download", cfg.Remote()), cfg.Remote())
	transfers := queue.WatchTransfers()
	var wait sync.WaitGroup
	wait.Add(1)
	go fillData(&wait, data, transfers)
	for _, pointer := range pointers {
		// only use first item, as other share the OID
		queue.Add(downloadTransfer(&lfs.WrappedPointer{Pointer: pointer}))
	}
	queue.Wait()
	wait.Wait()
	return data, checkForErrors(queue)
}

func removeUnnamed(data []*fetchUrlsObject) []*fetchUrlsObject {
	var result = make([]*fetchUrlsObject, 0, len(data))
	for _, item := range data {
		if item.Name != "" {
			result = append(result, item)
		}
	}
	return result
}

func dumpJson(data []*fetchUrlsObject) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", " ")
	output := struct {
		Files []*fetchUrlsObject `json:"files"`
	}{Files: data}
	if err := encoder.Encode(output); err != nil {
		ExitWithError(err)
	}
}

func dumpText(data []*fetchUrlsObject) {
	for _, item := range data {
		Print("%s %s", item.Name, item.Href)
		if item.Header != nil {
			for key, value := range item.Header {
				Print("  %s: %s", key, value)
			}
		}
	}
}

func fetchUrlsCommand(cmd *cobra.Command, args []string) {
	setupRepository()
	data, ok := populateData(args)
	data = removeUnnamed(data)
	if fetchUrlsJson {
		dumpJson(data)
	} else {
		dumpText(data)
	}
	if !ok {
		c := getAPIClient()
		e := c.Endpoints.Endpoint("download", cfg.Remote())
		Exit(tr.Tr.Get("error: failed to fetch some objects from '%s'", e.Url))
	}
}

func init() {
	RegisterCommand("fetch-urls", fetchUrlsCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&fetchUrlsJson, "json", "", false, "print output in JSON")
	})
}
