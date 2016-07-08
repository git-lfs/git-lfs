package transfer

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/github/git-lfs/config"
)

// Adapter for custom transfer via external process
type customAdapter struct {
	*adapterBase
	path       string
	args       string
	concurrent bool
}

func (a *customAdapter) ClearTempStorage() error {
	// no action requred
	return nil
}

func (a *customAdapter) DoTransfer(t *Transfer, cb TransferProgressCallback, authOkFunc func()) error {
	// TODO
	return nil
}

func newCustomAdapter(name string, dir Direction, path, args string, concurrent bool) *customAdapter {
	c := &customAdapter{newAdapterBase(name, dir, nil), path, args, concurrent}
	// self implements impl
	c.transferImpl = c
	return c
}

// Initialise custom adapters based on current config
func ConfigureCustomAdapters() {
	pathRegex := regexp.MustCompile(`lfs.customtransfer.([^.]+).path`)
	for k, v := range config.Config.AllGitConfig() {
		if match := pathRegex.FindStringSubmatch(k); match != nil {
			name := match[1]
			path := v
			var args string
			var concurrent bool
			var direction string
			// retrieve other values
			args, _ = config.Config.GitConfig(fmt.Sprintf("lfs.customtransfer.%s.args", name))
			concurrent = config.Config.GitConfigBool(fmt.Sprintf("lfs.customtransfer.%s.concurrent", name), true)
			direction, _ = config.Config.GitConfig(fmt.Sprintf("lfs.customtransfer.%s.direction", name))
			if len(direction) == 0 {
				direction = "both"
			} else {
				direction = strings.ToLower(direction)
			}

			// Separate closure for each since we need to capture vars above
			newfunc := func(name string, dir Direction) TransferAdapter {
				return newCustomAdapter(name, dir, path, args, concurrent)
			}

			if direction == "download" || direction == "both" {
				RegisterNewTransferAdapterFunc(name, Download, newfunc)
			}
			if direction == "upload" || direction == "both" {
				RegisterNewTransferAdapterFunc(name, Upload, newfunc)
			}

		}
	}

}

func init() {
	ConfigureCustomAdapters()
}
