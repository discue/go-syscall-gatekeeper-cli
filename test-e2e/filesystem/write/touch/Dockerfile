FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

CMD ["mv", "/etc/os-release", "/etc/os-release.copy"]