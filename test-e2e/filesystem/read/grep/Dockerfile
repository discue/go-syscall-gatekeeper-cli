FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

CMD ["grep", "debian", "/etc/os-release"]