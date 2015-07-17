FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src
LABEL POST="docker cp containerid:/tmp/docker_run/repo*"

RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y golang git dpkg-dev dh-golang ruby-ronn

ADD debian_script.bsh /tmp/docker_run/

CMD /tmp/docker_run/debian_script.bsh