FROM debian:wheezy
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-repo-dir:/src -v repo_dir:/tmp/docker_run/repo" 
LABEL POST="docker cp containerid:/repo*"

RUN echo 'deb http://http.debian.net/debian wheezy-backports main' > /etc/apt/sources.list.d/wheezy-backports-main.list

RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y -t wheezy-backports golang git dpkg-dev dh-golang ruby-ronn reprepro

ADD debian_script.bsh /tmp/docker_run/
ADD distributions /tmp/

CMD /tmp/docker_run/debian_script.bsh