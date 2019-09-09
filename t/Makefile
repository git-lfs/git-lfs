RM ?= rm -f
PROVE ?= prove
PROVE_EXTRA_ARGS =
DEFAULT_TEST_TARGET ?= test

GO ?= go
X =

TEST_CMDS =

TEST_CMDS += ../bin/git-credential-lfsnoop$X
TEST_CMDS += ../bin/git-credential-lfstest$X
TEST_CMDS += ../bin/lfs-askpass$X
TEST_CMDS += ../bin/lfs-ssh-echo$X
TEST_CMDS += ../bin/lfs-ssh-proxy-test$X
TEST_CMDS += ../bin/lfstest-count-tests$X
TEST_CMDS += ../bin/lfstest-customadapter$X
TEST_CMDS += ../bin/lfstest-gitserver$X
TEST_CMDS += ../bin/lfstest-realpath$X
TEST_CMDS += ../bin/lfstest-standalonecustomadapter$X
TEST_CMDS += ../bin/lfstest-testutils$X

# Not used for the integration tests, but build it here anyway to ensure it
# continues to work.
TEST_CMDS += ../bin/git-lfs-test-server-api$X

TEST_SRCS     = $(wildcard t-*.sh)
TEST_API_SRCS = $(wildcard git-lfs-test-server-api/*.go)

all : $(DEFAULT_TEST_TARGET)

test : $(TEST_CMDS)
	$(RM) -r remote test_count{,.lock}
	@GIT_LFS_NO_TEST_COUNT= bash -c '. ./testenv.sh && setup'
	$(PROVE) $(PROVE_EXTRA_ARGS) ./t-*.sh
	@GIT_LFS_NO_TEST_COUNT= bash -c '. ./testenv.sh && shutdown'

.PHONY : $(TEST_SRCS)
$(TEST_SRCS) : $(TEST_CMDS)
	$(RM) -r remote test_count{,.lock}
	$(PROVE) -v $(PROVE_EXTRA_ARGS) $@

.PHONY : clean
clean :
	$(RM) -r remote
	$(RM) $(TEST_CMDS)

../bin/%$X : cmd/%.go
	go build -o $@ $^

../bin/git-lfs-test-server-api$X : $(TEST_API_SRCS)
	go build -o $@ $^
