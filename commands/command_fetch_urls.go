package commands

import (
	"encoding/json"
	"os"

	"github.com/git-lfs/git-lfs/v3/errors"
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
	Url    string            `json:"url"`
	Header map[string]string `json:"headers,omitempty"`
}

func fetchUrlsCommand(cmd *cobra.Command, args []string) {
	setupRepository()
	var items = make([]*fetchUrlsObject, 0, len(args))
	var transfers = make([]*tq.Transfer, 0, len(args))
	for _, file := range args {
		pointer, err := lfs.DecodePointerFromFile(file)
		if err != nil {
			ExitWithError(err)
		}
		items = append(items, &fetchUrlsObject{Name: file, Oid: pointer.Oid, Size: pointer.Size})
		transfers = append(transfers, &tq.Transfer{Oid: pointer.Oid, Size: pointer.Size})
	}
	bRes, err := tq.Batch(getTransferManifestOperationRemote("download", cfg.Remote()), tq.Download, cfg.Remote(), nil, transfers)
	if err != nil {
		ExitWithError(err)
	}
	for i := 0; i < len(args); i++ {
		o := bRes.Objects[i]
		item := items[i]
		if o.Error != nil {
			ExitWithError(o.Error)
		}
		if o.Oid != item.Oid {
			ExitWithError(errors.Errorf(tr.Tr.Get("oid mismatch: expected %s, got %s", item.Oid, o.Oid)))
		}
		a := o.Actions["download"]
		item.Url = a.Href
		item.Header = a.Header
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
