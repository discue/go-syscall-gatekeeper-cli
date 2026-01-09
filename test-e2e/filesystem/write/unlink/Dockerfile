FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN mkdir -p /tmp/ \
 && touch /tmp/my-file

CMD ["unlink", "/tmp/my-file"]