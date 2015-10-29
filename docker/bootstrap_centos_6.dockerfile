FROM centos:6
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src -v repo_dir:/repo"

RUN yum install -y epel-release && \
    yum install -y createrepo rsync tar gnupg2 expect

ENV GOLANG_VERSION=[{GOLANG_VERSION}]

ENV GOROOT=/usr/local/go

RUN cd /usr/local && \
    curl -L -O https://storage.googleapis.com/golang/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar zxf go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    ln -s /usr/local/go/bin/go /usr/bin/go && \
    ln -s /usr/local/go/bin/gofmt /usr/bin/gofmt

#The purpose of this is to build and install everything needed to build git-lfs
#Next time. So that the LONG build/installed in centos are only done once, and
#stored in the image.

#Set to master if you want the lastest, but IF there is a failure,
#the docker will not build, so I decided to make a stable version the default
ENV DOCKER_LFS_BUILD_VERSION=[{DOCKER_LFS_BUILD_VERSION}]

ADD https://github.com/github/git-lfs/archive/${DOCKER_LFS_BUILD_VERSION}.tar.gz /tmp/docker_setup/
RUN cd /tmp/docker_setup/; \
    tar zxf ${DOCKER_LFS_BUILD_VERSION}.tar.gz; \
    cd /tmp/docker_setup/git-lfs-*/rpm; \
    touch build.log; \
    tail -f build.log & ./build_rpms.bsh; \
    pkill tail; \
    rm -rvf /tmp/docker_setup/git-lfs-*/rpm/BUILD*
