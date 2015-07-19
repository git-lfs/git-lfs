FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

#Docker RUN example, pass in the git-lfs checkout copy you are working with
LABEL RUN="docker run -v git-lfs-checkout-dir:/src -v repo_dir:/repo" 



RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y golang git dpkg-dev dh-golang ruby-ronn reprepro

ADD dpkg-package-gpg.bsh /tmp/
ADD distributions /tmp/
ADD debian_script.bsh /tmp/docker_run/
ADD signing.key /tmp/

CMD /tmp/docker_run/debian_script.bsh