FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

# -v /tmp
LABEL RUN="docker run -v /tmp/gpg-agent" 
LABEL STOP="docker exec pkill gpgp-agent"

RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y gnupg-agent gnupg2

VOLUME /tmp/gpg-agent

COPY .start_gpg-agent.bsh signing.key /tmp/

CMD /tmp/.start_gpg-agent.bsh