FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

CMD ["cp", "/etc/os-release", "/etc/os-release.copy"]