GO ?= go

GLIDE ?= glide

GREP ?= grep
RM ?= rm -f
XARGS ?= xargs

GOIMPORTS ?= goimports
GOIMPORTS_EXTRA_OPTS ?= -w -l

LFS_PACKAGES =
LFS_PACKAGES += commands
LFS_PACKAGES += config
LFS_PACKAGES += errors
LFS_PACKAGES += filepathfilter
LFS_PACKAGES += fs
LFS_PACKAGES += git
LFS_PACKAGES += git/gitattr
LFS_PACKAGES += git/githistory
LFS_PACKAGES += git
LFS_PACKAGES += lfs
LFS_PACKAGES += lfsapi
LFS_PACKAGES += locking
LFS_PACKAGES += subprocess
LFS_PACKAGES += tasklog
LFS_PACKAGES += tools
LFS_PACKAGES += tools/humanize
LFS_PACKAGES += tools/kv
LFS_PACKAGES += tq

glide.lock : glide.yaml
	$(GLIDE) update

vendor : glide.lock
	$(GLIDE) install
	$(RM) -r vendor/github.com/ThomsonReutersEikon/go-ntlm/utils
	$(RM) -r vendor/github.com/davecgh/go-spew
	$(RM) -r vendor/github.com/pmezard/go-difflib

fmt : $(LFS_PACKAGES) | lint
	$(GOIMPORTS) $(GOIMPORTS_EXTRA_OPTS) $?

lint : $(LFS_PACKAGES)
	$(GO) list -f '{{ join .Deps "\n" }}' . \
	| $(XARGS) $(GO) list -f '{{ if not .Standard }}{{ .ImportPath }}{{ end }}' \
	| $(GREP) -v "github.com/git-lfs/git-lfs" || exit 0
