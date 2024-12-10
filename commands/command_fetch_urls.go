package commands

import (
	"encoding/json"
	"os"

	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/spf13/cobra"
)

var (
	fetchUrlsJson = false
)

type fetchUrlsObject struct {
	Name   string            `json:"name"`
	Oid    string            `json:"oid"`
	Size   int64             `json:"size"`
	Url    string            `json:"url"`
	Header map[string]string `json:"headers,omitempty"`
}

type transferData struct {
	Transfer tq.Transfer
	Objects  []*fetchUrlsObject
}

func fetchUrlsCommand(cmd *cobra.Command, args []string) {
	setupRepository()
	// items is the list of objects to fetch, initially populated from CLI args, then expanded by the batch API
	var items = make([]*fetchUrlsObject, 0, len(args))
	// transferMap is a map from OIDs to transfer objects + a list of fetchUrlsObject instances to fill
	var transferMap = make(map[string]*transferData)
	// transfers is the list of transfer objects taken from the above maps to be passed to `Batch`
	var transfers = make([]*tq.Transfer, 0, len(args))
	for _, file := range args {
		pointer, err := lfs.DecodePointerFromFile(file)
		if err != nil {
			ExitWithError(err)
		}
		data, exists := transferMap[pointer.Oid]
		item := &fetchUrlsObject{Name: file, Oid: pointer.Oid, Size: pointer.Size}
		items = append(items, item)
		if !exists {
			data = &transferData{Transfer: tq.Transfer{Oid: pointer.Oid, Size: pointer.Size}, Objects: make([]*fetchUrlsObject, 0, 1)}
			transferMap[pointer.Oid] = data
			transfers = append(transfers, &data.Transfer)
		}
		data.Objects = append(data.Objects, item)
	}
	bRes, err := tq.Batch(getTransferManifestOperationRemote("download", cfg.Remote()), tq.Download, cfg.Remote(), nil, transfers)
	if err != nil {
		ExitWithError(err)
	}
	for _, o := range bRes.Objects {
		data, exists := transferMap[o.Oid]
		if !exists {
			continue
		}
		if o.Error != nil {
			ExitWithError(o.Error)
		} // TODO: retry
		a := o.Actions["download"]
		for _, item := range data.Objects {
			item.Url = a.Href
			item.Header = a.Header
		}
	}
	if fetchUrlsJson {
		data := struct {
			Urls []*fetchUrlsObject `json:"files"`
		}{Urls: items}
		encoder := json.NewEncoder(os.Stdout)
		encoder.SetIndent("", " ")
		if err := encoder.Encode(data); err != nil {
			ExitWithError(err)
		}
	} else {
		for _, item := range items {
			Print("%s %s", item.Name, item.Url)
		}
	}

}

func init() {
	RegisterCommand("fetch-urls", fetchUrlsCommand, func(cmd *cobra.Command) {
		cmd.Flags().BoolVarP(&fetchUrlsJson, "json", "", false, "print output in JSON")
	})
}
