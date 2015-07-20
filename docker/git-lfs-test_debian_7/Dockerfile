FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-checkout-dir:/src" 

ADD test_lfs.bsh /tmp/test_lfs.bsh

#TODO: Needs to be replaced by an apt repo
ADD public.key /etc/apt/trusted.gpg.d/git-lfs.gpg
ADD git-lfs-main_7.list /etc/apt/sources.list.d/git-lfs-main.list

#These SHOULD be throw away commands, and not stored as Docker commits
CMD DEBIAN_FRONTEND=noninteractive \
    apt-get -y update && \
    apt-get install -y git-lfs && \
    git lfs && \
    apt-get install -y golang curl && \
    /tmp/test_lfs.bsh