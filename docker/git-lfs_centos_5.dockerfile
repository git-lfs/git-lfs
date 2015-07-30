FROM centos:5
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

SOURCE bootstrap_centos_5.dockerfile

RUN rm -rf /tmp/docker_setup

#Add the simple build repo script
COPY rpm_sign.exp git-lfs_centos_5.key centos_script.bsh /tmp/

CMD /tmp/centos_script.bsh
