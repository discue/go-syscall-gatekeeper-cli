FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper
RUN ln -s /etc/passwd /tmp/passwd.link

CMD ["cat", "/tmp/passwd.link"]
