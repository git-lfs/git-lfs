FROM debian:jessie

MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/tmp/git-lfs"

RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y golang git dpkg-dev dh-golang ruby-ronn

WORKDIR /tmp/git-lfs/
CMD ["dpkg-buildpackage", "-us", "-uc"]