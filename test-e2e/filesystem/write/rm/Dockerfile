FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN mkdir -p /tmp/new_directory

CMD ["rm", "-rf", "/tmp/new_directory"]