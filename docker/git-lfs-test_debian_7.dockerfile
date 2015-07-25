FROM debian:wheezy
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-checkout-dir:/src -v repo_dir:/repo" 

COPY test_lfs.bsh /tmp/test_lfs.bsh

#TODO: Needs to be replaced by an apt repo
COPY public.key /etc/apt/trusted.gpg.d/git-lfs.gpg
COPY git-lfs-main_7.list /etc/apt/sources.list.d/git-lfs-main.list
RUN gpg --dearmour -o /etc/apt/trusted.gpg.d/stupid-debian.gpg /etc/apt/trusted.gpg.d/git-lfs.gpg
RUN rm /etc/apt/trusted.gpg.d/git-lfs.gpg
RUN apt-key add /etc/apt/trusted.gpg.d/stupid-debian.gpg 
#ssgelm said I didn't need to do this, but I don't know how not to.

#These SHOULD be throw away commands, and not stored as Docker commits
CMD echo 'deb http://http.debian.net/debian wheezy-backports main' > /etc/apt/sources.list.d/wheezy-backports-main.list && \
    DEBIAN_FRONTEND=noninteractive apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y -t wheezy-backports git-lfs && \
    git lfs && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y -t wheezy-backports golang curl && \
    /tmp/test_lfs.bsh
