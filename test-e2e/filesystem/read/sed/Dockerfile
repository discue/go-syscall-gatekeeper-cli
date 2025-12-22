FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

CMD ["sed", "-n", "10,20p", "/etc/os-release"]