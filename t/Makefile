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
TEST_CMDS += ../bin/lfstest-standalonecustomadapter$X
TEST_CMDS += ../bin/lfstest-testutils$X

all : $(DEFAULT_TEST_TARGET)

test : $(TEST_CMDS)
	$(RM) -r remote test_count{,.lock}
	( . ./testenv.sh && setup )
	prove $(PROVE_EXTRA_ARGS) ./t-*.sh
	( . ./testenv.sh && shutdown )

./t-%.sh : $(TEST_CMDS)
	$(RM) -r remote test_count{,.lock}
	prove -v $(PROVE_EXTRA_ARGS) $@

.PHONY : clean
clean :
	$(RM) -r remote
	$(RM) $(TEST_CMDS)

../bin/%$X : cmd/%.go
	go build -o $@ $^
