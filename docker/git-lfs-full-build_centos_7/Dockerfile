FROM centos:7
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src"
LABEL POST="docker cp containerid:/tmp/docker_run/rpms/RPMS/*"

RUN yum install -y createrepo

#Add the simple build repo script
ADD centos_script.bsh /tmp/docker_run

CMD /tmp/docker_run/centos_script