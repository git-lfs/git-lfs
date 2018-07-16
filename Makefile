GO ?= go

GLIDE ?= glide

GREP ?= grep
RM ?= rm -f
XARGS ?= xargs

GOIMPORTS ?= goimports
GOIMPORTS_EXTRA_OPTS ?= -w -l

ifndef PKGS
PKGS =
PKGS += commands
PKGS += config
PKGS += errors
PKGS += filepathfilter
PKGS += fs
PKGS += git
PKGS += git/gitattr
PKGS += git/githistory
PKGS += git
PKGS += lfs
PKGS += lfsapi
PKGS += locking
PKGS += subprocess
PKGS += tasklog
PKGS += tools
PKGS += tools/humanize
PKGS += tools/kv
PKGS += tq
endif

glide.lock : glide.yaml
	$(GLIDE) update

vendor : glide.lock
	$(GLIDE) install
	$(RM) -r vendor/github.com/ThomsonReutersEikon/go-ntlm/utils
	$(RM) -r vendor/github.com/davecgh/go-spew
	$(RM) -r vendor/github.com/pmezard/go-difflib

fmt : $(PKGS) | lint
	$(GOIMPORTS) $(GOIMPORTS_EXTRA_OPTS) $?

lint : $(PKGS)
	$(GO) list -f '{{ join .Deps "\n" }}' . \
	| $(XARGS) $(GO) list -f '{{ if not .Standard }}{{ .ImportPath }}{{ end }}' \
	| $(GREP) -v "github.com/git-lfs/git-lfs" || exit 0
