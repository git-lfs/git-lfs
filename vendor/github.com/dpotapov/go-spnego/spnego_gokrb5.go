// +build !windows

package spnego

import (
	"net/http"
	"os"
	"os/user"
	"strings"

	"gopkg.in/jcmturner/gokrb5.v5/client"
	"gopkg.in/jcmturner/gokrb5.v5/config"
	"gopkg.in/jcmturner/gokrb5.v5/credentials"
)

type krb5 struct {
	cfg *config.Config
	cl  client.Client
}

// New constructs OS specific implementation of spnego.Provider interface
func New() Provider {
	return &krb5{}
}

func (k *krb5) makeCfg() error {
	if k.cfg != nil {
		return nil
	}

	cfgPath := os.Getenv("KRB5_CONFIG")
	if _, err := os.Stat(cfgPath); os.IsNotExist(err) {
		cfgPath = "/etc/krb5.conf" // ToDo: Macs and Windows have different path, also some Unix may have /etc/krb5/krb5.conf
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		return err
	}

	k.cfg = cfg
	return nil
}

func (k *krb5) makeClient() error {
	u, err := user.Current()
	if err != nil {
		return err
	}

	ccpath := "/tmp/krb5cc_" + u.Uid

	ccname := os.Getenv("KRB5CCNAME")
	if strings.HasPrefix(ccname, "FILE:") {
		ccpath = strings.SplitN(ccname, ":", 2)[1]
	}

	ccache, err := credentials.LoadCCache(ccpath)
	if err != nil {
		return err
	}

	cl, err := client.NewClientFromCCache(ccache)
	if err != nil {
		return err
	}

	cl.GoKrb5Conf.DisablePAFXFast = true
	cl.WithConfig(k.cfg)
	k.cl = cl
	return nil
}

func (k *krb5) SetSPNEGOHeader(req *http.Request) error {
	h, err := canonicalizeHostname(req.URL.Hostname())
	if err != nil {
		return err
	}

	if err := k.makeCfg(); err != nil {
		return err
	}

	if err := k.makeClient(); err != nil {
		return err
	}

	return k.cl.SetSPNEGOHeader(req, "HTTP/"+h)
}
