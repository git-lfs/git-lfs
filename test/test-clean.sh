#!/bin/sh

. "test/testlib.sh"


begin_test "clean"
(
  set -e

  mkdir repo
  cd repo
  git init

  # clean a simple file
  echo "whatever" | git lfs clean | tee clean.log
  [ "$(pointer cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411 9)" = "$(cat clean.log)" ]

  # clean a git lfs pointer
  pointer cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411 9 | git lfs clean | tee clean.log
  [ "$(pointer cd293be6cea034bd45a0352775a219ef5dc7825ce55d1f7dae9762d80ce64411 9)" = "$(cat clean.log)" ]

  # clean a pseudo pointer with extra data
  echo "version https://git-lfs.github.com/spec/v1
oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5
size 1024

This is my test pointer.  There are many like it, but this one is mine." | git lfs clean | tee clean.log
  [ "$(pointer f492acbebb5faa22da4c1501c022af035469f624f426631f31936575873fefe1 202)" = "$(cat clean.log)" ]

  # clean a pseudo pointer with extra data separated by enough white space to
  # fill the 'git lfs clean' buffer
  echo "version https://git-lfs.github.com/spec/v1
oid sha256:7cd8be1d2cd0dd22cd9d229bb6b5785009a05e8b39d405615d882caac56562b5
size 1024
\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n\n
This is my test pointer.  There are many like it, but this one is mine." | git lfs clean | tee clean.log
  [ "$(pointer c2f909f6961bf85a92e2942ef3ed80c938a3d0ebaee6e72940692581052333be 586)" = "$(cat clean.log)" ]
)
end_test
