FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-checkout-dir:/src -v repo_dir:/repo" 

COPY test_lfs.bsh /tmp/test_lfs.bsh

#TODO: Needs to be replaced by an apt repo
RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y wget
ENV REPO_HOSTNAME="[{REPO_HOSTNAME}]"
RUN echo "deb http://${REPO_HOSTNAME:-git-lfs.github.com}/debian/8 jessie main" > /etc/apt/sources.list.d/git-lfs-main.list
RUN wget -P /tmp/ http://${REPO_HOSTNAME:-git-lfs.github.com}/debian/8/GPG-KEY-GITLFS || :
RUN [ ! -s /tmp/GPG-KEY-GITLFS ] || \
    gpg --dearmour -o /etc/apt/trusted.gpg.d/git-lfs.gpg /tmp/GPG-KEY-GITLFS
RUN [ ! -s /etc/apt/trusted.gpg.d/git-lfs.gpg ] || \
    apt-key add /etc/apt/trusted.gpg.d/git-lfs.gpg
#I wasn't supposed to need to do this, but I don't know how not to.

#These SHOULD be throw away commands, and not stored as Docker commits
CMD DEBIAN_FRONTEND=noninteractive apt-get -y update && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y git-lfs && \
    git lfs && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y golang curl && \
    /tmp/test_lfs.bsh
