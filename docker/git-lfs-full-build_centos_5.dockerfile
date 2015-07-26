FROM centos:5
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

SOURCE bootstrap_centos_5.dockerfile

CMD rm -rf /tmp/docker_setup/*/rpm/SRPMS/git-lfs* /tmp/docker_setup/*/rpm/RPMS/*/git-lfs* && \
    rsync -ra /tmp/docker_setup/*/rpm/{RPMS,SRPMS} /repo && \
    createrepo /repo/SRPMS && \
    createrepo /repo/RPMS
