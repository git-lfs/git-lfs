FROM centos:7
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/tmp/git-lfs"

RUN yum install -y createrepo rsync git ruby ruby-devel golang

#The purpose of this is to build and install everything needed to build git-lfs
#Next time. So that the LONG build/installed in centos are only done once, and
#Store in the image

ENV DOCKER_LFS_BUILD_VERSION=master
#ENV DOCKER_LFS_BUILD_VERSION=v0.5.3
#Set to master if you want the lastest, but IF there is a failure,
#the docker will not build, so I decided to make a stable version the default

#ADD https://github.com/github/git-lfs/archive/${DOCKER_LFS_BUILD_VERSION}.tar.gz /tmp/docker_setup/
#TODO: Change these two lines back after next merge
ENV DOCKER_LFS_BUILD_VERSION=git_for_centos
ADD https://github.com/andyneff/git-lfs/archive/${DOCKER_LFS_BUILD_VERSION}.tar.gz /tmp/docker_setup/
WORKDIR /tmp/docker_setup/
RUN tar zxf ${DOCKER_LFS_BUILD_VERSION}.tar.gz
RUN cd /tmp/docker_setup/git-lfs-*/rpm; touch build.log; tail -f build.log & ./build_rpms.bsh; pkill tail
RUN rm -rf /tmp/docker_setup

WORKDIR /tmp
CMD /tmp/git-lfs/rpm/build_rpms.bsh