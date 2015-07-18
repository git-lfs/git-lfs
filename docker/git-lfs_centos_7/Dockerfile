FROM centos:7
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src"
LABEL POST="docker cp containerid:/repo*"


RUN yum install -y createrepo rsync git ruby ruby-devel golang

#The purpose of this is to build and install everything needed to build git-lfs
#Next time. So that the LONG build/installed in centos are only done once, and
#stored in the image.

ENV DOCKER_LFS_BUILD_VERSION=master
#ENV DOCKER_LFS_BUILD_VERSION=v0.5.3
#Set to master if you want the lastest, but IF there is a failure,
#the docker will not build, so I decided to make a stable version the default
ADD https://github.com/github/git-lfs/archive/${DOCKER_LFS_BUILD_VERSION}.tar.gz /tmp/docker_setup/
RUN cd /tmp/docker_setup/; \
    tar zxf ${DOCKER_LFS_BUILD_VERSION}.tar.gz
RUN cd /tmp/docker_setup/git-lfs-*/rpm; \
    touch build.log; \
    tail -f build.log & ./build_rpms.bsh; \
    pkill tail
RUN rm -rf /tmp/docker_setup

#Add the simple build repo script
ADD centos_script.bsh /tmp/docker_run/

CMD /tmp/docker_run/centos_script.bsh