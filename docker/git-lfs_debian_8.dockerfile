FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-checkout-dir:/src -v repo_dir:/repo" 



RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y git dpkg-dev dh-golang ruby-ronn reprepro curl

ENV GOLANG_VERSION=[{GOLANG_VERSION}]

ENV GOROOT=/usr/local/go

RUN cd /usr/local && \
    curl -L -O https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar zxf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt

COPY git-lfs_debian_8.key debian_script.bsh distributions dpkg-package-gpg.bsh /tmp/

CMD /tmp/debian_script.bsh