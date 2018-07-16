GLIDE ?= glide

RM ?= rm -f

glide.lock : glide.yaml
	$(GLIDE) update

vendor : glide.lock
	$(GLIDE) install
	$(RM) -r vendor/github.com/ThomsonReutersEikon/go-ntlm/utils
	$(RM) -r vendor/github.com/davecgh/go-spew
	$(RM) -r vendor/github.com/pmezard/go-difflib
