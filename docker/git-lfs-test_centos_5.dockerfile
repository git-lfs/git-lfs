FROM centos:5
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src -v repo_dir:/repo"

COPY test_lfs.bsh /tmp/test_lfs.bsh

CMD yum install -y curl.x86_64 && \
    curl -L -O http://${REPO_HOSTNAME:-git-lfs.github.com}/centos/5/RPMS/noarch/git-lfs-repo-release-1-1.noarch.rpm && \
    yum install -y --nogpgcheck git-lfs-repo-release-1-1.noarch.rpm &&\
    yum install -y epel-release &&\
    yum install -y --nogpgcheck git-lfs && \
    git-lfs && \
    yum install -y --nogpgcheck perl-Digest-SHA golang && \
    /tmp/test_lfs.bsh
