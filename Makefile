GOC ?= gccgo
AR ?= ar

SRCDIR := $(dir $(lastword $(MAKEFILE_LIST)))

LIBDIR := out/github.com/git-lfs/git-lfs
GOFLAGS := -Iout

ifeq ($(MAKEFILE_GEN),)

MAKEFILE_GEN := out/Makefile.gen

all: $(MAKEFILE_GEN)
	@$(MAKE) -f $(lastword $(MAKEFILE_LIST)) $(MAKEFLAGS) MAKEFILE_GEN=$(MAKEFILE_GEN) $@

$(MAKEFILE_GEN) : out/genmakefile $(SRCDIR)commands/mancontent_gen.go
	@mkdir -p $(dir $@)
	$< "$(SRCDIR)" github.com/git-lfs/git-lfs/ > $@

else

all : bin/git-lfs

include $(MAKEFILE_GEN)

$(LIBDIR)/git-lfs.o : $(SRC_main) $(DEPS_main)
	@mkdir -p $(dir $@)
	$(GOC) $(GOFLAGS) -c -o $@ $(SRC_main)

bin/git-lfs : $(LIBDIR)/git-lfs.o $(DEPS_main)
	@mkdir -p $(dir $@)
	$(GOC) $(GOFLAGS) -o $@ $^

%.a : %.o
	$(AR) rc $@ $<

endif

$(SRCDIR)commands/mancontent_gen.go : out/mangen
	cd $(SRCDIR)commands && $(CURDIR)/out/mangen

out/mangen : $(SRCDIR)docs/man/mangen.go
	@mkdir -p $(dir $@)
	$(GOC) -o $@ $<

out/genmakefile : $(SRCDIR)script/genmakefile/genmakefile.go
	@mkdir -p $(dir $@)
	$(GOC) -o $@ $<

clean :
	rm -rf out bin
	rm -f $(SRCDIR)commands/mancontent_gen.go
