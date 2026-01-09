FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN mkdir -p /tmp/new_directory

CMD ["rmdir", "--ignore-fail-on-non-empty", "/tmp/new_directory"]