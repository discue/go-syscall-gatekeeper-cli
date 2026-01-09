FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN mkdir -p /tmp

CMD ["ln","-s","/etc/os-release","/tmp/os-release.symlink"]
