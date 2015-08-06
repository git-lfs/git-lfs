FROM debian:jessie
MAINTAINER Andy Neff <andyneff@users.noreply.github.com>

# -v /tmp
LABEL RUN="docker run -v /tmp/gpg-agent" 
LABEL STOP="docker exec pkill gpgp-agent"

RUN DEBIAN_FRONTEND=noninteractive apt-get -y update && \
apt-get install -y gnupg-agent gnupg2

ENV GNUPGHOME=/tmp/gpg-agent

VOLUME /tmp/gpg-agent

COPY *.key /tmp/

CMD chmod 700 /tmp/gpg-agent; \
eval $(gpg-agent --write-env-file /tmp/gpg-agent/gpg_agent_info \
                 --use-standard-socket --daemon \
                 --default-cache-ttl=${GPG_DEFAULT_CACHE:-31536000} \
                 --max-cache-ttl=${GPG_MAX_CACHE:-31536000} ); \
while gpg-connect-agent /bye; do \
  sleep 2; \
done