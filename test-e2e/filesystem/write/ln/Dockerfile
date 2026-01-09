FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN mkdir -p /tmp

CMD ["ln", "/etc/os-release", "/tmp/os-release.link"]