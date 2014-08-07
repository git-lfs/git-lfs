package commands

import (
	"fmt"
	"github.com/github/git-media/gitmedia"
	"github.com/github/git-media/gitmediaclient"
	"github.com/spf13/cobra"
	"strings"
)

var (
	pushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push files to the media endpoint",
		Run:   pushCommand,
	}
)

func pushCommand(cmd *cobra.Command, args []string) {
	q, err := gitmedia.UploadQueue()
	if err != nil {
		Panic(err, "Error setting up the queue")
	}

	count, err := q.Count()
	i := 1

	q.Walk(func(id string, body []byte) error {
		fileInfo := string(body)
		parts := strings.SplitN(fileInfo, ":", 2)

		var oid, filename string
		oid = parts[0]
		if len(parts) > 1 {
			filename = parts[1]
		}

		if wErr := pushAsset(oid, filename, i, count); wErr != nil {
			Panic(wErr.Inner, wErr.Message)
		}
		i += 1

		fmt.Printf("\n")

		if err := q.Del(id); err != nil {
			Panic(err, "error removing %s from queue", oid)
		}

		return nil
	})
}

func pushAsset(oid, filename string, index, totalFiles int) *wrappedError {
	path, err := gitmedia.LocalMediaPath(oid)
	if err == nil {
		err = gitmediaclient.Options(path)
	}

	if err == nil {
		cb, file, cbErr := gitmedia.CopyCallbackFile("push", filename, index, totalFiles)
		if cbErr != nil {
			Error(cbErr.Error())
		}

		err = gitmediaclient.Put(path, filename, cb)
		if file != nil {
			file.Close()
		}
	}

	if err != nil {
		return &wrappedError{
			Message: fmt.Sprintf("error uploading file %s/%s", oid, filename),
			Inner:   err,
		}
	}

	return nil
}

type wrappedError struct {
	Message string
	Inner   error
}

func init() {
	RootCmd.AddCommand(pushCmd)
}
