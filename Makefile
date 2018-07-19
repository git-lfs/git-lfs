GIT_LFS_SHA ?= $(shell git rev-parse --short HEAD)
VERSION ?= $(shell git describe HEAD)

GO ?= go

GO_TEST_EXTRA_ARGS =

BUILTIN_LD_FLAGS =
BUILTIN_LD_FLAGS += -X github.com/git-lfs/git-lfs/config.GitCommit=$(GIT_LFS_SHA)
ifneq ("$(DWARF)","YesPlease")
BUILTIN_LD_FLAGS += -s
BUILTIN_LD_FLAGS += -w
endif
EXTRA_LD_FLAGS =
LD_FLAGS = $(BUILTIN_LD_FLAGS) $(EXTRA_LD_FLAGS)

BUILTIN_GC_FLAGS =
EXTRA_GC_FLAGS =
GC_FLAGS = $(BUILTIN_GC_FLAGS) $(EXTRA_GC_FLAGS)

GLIDE ?= glide

RONN ?= ronn
RONN_EXTRA_ARGS ?=

GREP ?= grep
RM ?= rm -f
XARGS ?= xargs

GOIMPORTS ?= goimports
GOIMPORTS_EXTRA_OPTS ?= -w -l

SOURCES = $(shell find . -type f -name '*.go')
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

ifeq ($(OS),Windows_NT)
X ?= .exe
else
X ?=
endif
.DEFAULT_GOAL := bin/git-lfs$(X)

BUILD = GOOS=$(1) GOARCH=$(2) \
	$(GO) build \
	-ldflags="$(LD_FLAGS)" \
	-gcflags="$(GC_FLAGS)" \
	-o ./bin/git-lfs$(3) ./git-lfs.go

BUILD_TARGETS = bin/git-lfs-darwin-amd64 bin/git-lfs-darwin-386 \
	bin/git-lfs-linux-amd64 bin/git-lfs-linux-386 \
	bin/git-lfs-freebsd-amd64 bin/git-lfs-freebsd-386 \
	bin/git-lfs-windows-amd64.exe bin/git-lfs-windows-386.exe

.PHONY : all build
all build : $(BUILD_TARGETS)

bin/git-lfs-darwin-amd64 : fmt
	$(call BUILD,darwin,amd64,-darwin-amd64)
bin/git-lfs-darwin-386 : fmt
	$(call BUILD,darwin,386,-darwin-386)
bin/git-lfs-linux-amd64 : fmt
	$(call BUILD,linux,amd64,-linux-amd64)
bin/git-lfs-linux-386 : fmt
	$(call BUILD,linux,386,-linux-386)
bin/git-lfs-freebsd-amd64 : fmt
	$(call BUILD,freebsd,amd64,-freebsd-amd64)
bin/git-lfs-freebsd-386 : fmt
	$(call BUILD,freebsd,386,-freebsd-386)
bin/git-lfs-windows-amd64.exe : version-info fmt
	$(call BUILD,windows,amd64,-windows-amd64.exe)
bin/git-lfs-windows-386.exe : version-info fmt
	$(call BUILD,windows,386,-windows-386.exe)

bin/git-lfs : $(SOURCES)
	$(call BUILD,$(GOOS),$(GOARCH),)

bin/git-lfs.exe : $(SOURCES) version-info
	$(call BUILD,$(GOOS),$(GOARCH),.exe)

.PHONY : version-info
version-info:
	go get github.com/josephspurrier/goversioninfo/cmd/goversioninfo
	PATH=$$PATH:$$GOPATH/bin/windows_386 $(GO) generate

RELEASE_TARGETS = bin/releases/git-lfs-darwin-amd64-$(VERSION).tar.gz \
	bin/releases/git-lfs-darwin-386-$(VERSION).tar.gz \
	bin/releases/git-lfs-linux-amd64-$(VERSION).tar.gz \
	bin/releases/git-lfs-linux-386-$(VERSION).tar.gz \
	bin/releases/git-lfs-freebsd-amd64-$(VERSION).tar.gz \
	bin/releases/git-lfs-freebsd-386-$(VERSION).tar.gz \
	bin/releases/git-lfs-windows-amd64-$(VERSION).zip \
	bin/releases/git-lfs-windows-386-$(VERSION).zip

RELEASE_INCLUDES = README.md CHANGELOG.md script/install.sh

.PHONY : release
release : $(RELEASE_TARGETS)
	shasum -a 256 $(RELEASE_TARGETS)

bin/releases/git-lfs-%-$(VERSION).tar.gz : $(RELEASE_INCLUDES) bin/git-lfs-%
	@mkdir -p bin/releases
	tar -s '!bin/git-lfs-.*!git-lfs!' -s '!script/!!' -czf $@ $^

bin/releases/git-lfs-%-$(VERSION).zip : $(RELEASE_INCLUDES) bin/git-lfs-%.exe
	@mkdir -p bin/releases
	zip -j -l $@ $^

TEST_TARGETS := test-bench test-verbose test-race
.PHONY : $(TEST_TARGETS) test
$(TEST_TARGETS) : test

test-bench : GO_TEST_EXTRA_ARGS=-run=__nothing__ -bench=.
test-verbose : GO_TEST_EXTRA_ARGS=-v
test-race : GO_TEST_EXTRA_ARGS=-race

test : fmt
	$(GO) test $(GO_TEST_EXTRA_ARGS) $(addprefix ./,$(PKGS))

glide.lock : glide.yaml
	$(GLIDE) update

vendor : glide.lock
	$(GLIDE) install
	$(RM) -r vendor/github.com/ThomsonReutersEikon/go-ntlm/utils
	$(RM) -r vendor/github.com/davecgh/go-spew
	$(RM) -r vendor/github.com/pmezard/go-difflib

.PHONY : fmt
fmt : $(SOURCES) | lint
	$(GOIMPORTS) $(GOIMPORTS_EXTRA_OPTS) $?

.PHONY : lint
lint : $(SOURCES)
	$(GO) list -f '{{ join .Deps "\n" }}' . \
	| $(XARGS) $(GO) list -f '{{ if not .Standard }}{{ .ImportPath }}{{ end }}' \
	| $(GREP) -v "github.com/git-lfs/git-lfs" || exit 0

MAN_ROFF_TARGETS = man/git-lfs-checkout.1 \
  man/git-lfs-clean.1 \
  man/git-lfs-clone.1 \
  man/git-lfs-config.5 \
  man/git-lfs-env.1 \
  man/git-lfs-ext.1 \
  man/git-lfs-fetch.1 \
  man/git-lfs-filter-process.1 \
  man/git-lfs-fsck.1 \
  man/git-lfs-install.1 \
  man/git-lfs-lock.1 \
  man/git-lfs-locks.1 \
  man/git-lfs-logs.1 \
  man/git-lfs-ls-files.1 \
  man/git-lfs-migrate.1 \
  man/git-lfs-pointer.1 \
  man/git-lfs-post-checkout.1 \
  man/git-lfs-post-merge.1 \
  man/git-lfs-pre-push.1 \
  man/git-lfs-prune.1 \
  man/git-lfs-pull.1 \
  man/git-lfs-push.1 \
  man/git-lfs-smudge.1 \
  man/git-lfs-status.1 \
  man/git-lfs-track.1 \
  man/git-lfs-uninstall.1 \
  man/git-lfs-unlock.1 \
  man/git-lfs-untrack.1 \
  man/git-lfs-update.1 \
  man/git-lfs.1

MAN_HTML_TARGETS = man/git-lfs-checkout.1.html \
  man/git-lfs-clean.1.html \
  man/git-lfs-clone.1.html \
  man/git-lfs-config.5.html \
  man/git-lfs-env.1.html \
  man/git-lfs-ext.1.html \
  man/git-lfs-fetch.1.html \
  man/git-lfs-filter-process.1.html \
  man/git-lfs-fsck.1.html \
  man/git-lfs-install.1.html \
  man/git-lfs-lock.1.html \
  man/git-lfs-locks.1.html \
  man/git-lfs-logs.1.html \
  man/git-lfs-ls-files.1.html \
  man/git-lfs-migrate.1.html \
  man/git-lfs-pointer.1.html \
  man/git-lfs-post-checkout.1.html \
  man/git-lfs-post-merge.1.html \
  man/git-lfs-pre-push.1.html \
  man/git-lfs-prune.1.html \
  man/git-lfs-pull.1.html \
  man/git-lfs-push.1.html \
  man/git-lfs-smudge.1.html \
  man/git-lfs-status.1.html \
  man/git-lfs-track.1.html \
  man/git-lfs-uninstall.1.html \
  man/git-lfs-unlock.1.html \
  man/git-lfs-untrack.1.html \
  man/git-lfs-update.1.html \
  man/git-lfs.1.html

.PHONY : man
man : $(MAN_ROFF_TARGETS) $(MAN_HTML_TARGETS)

man/% : docs/man/%.ronn
	@mkdir -p man
	$(RONN) $(RONN_EXTRA_ARGS) -r --pipe < $^ > $@

man/%.html : docs/man/%.ronn
	@mkdir -p man
	$(RONN) $(RONN_EXTRA_ARGS) -5 --pipe < $^ > $@
