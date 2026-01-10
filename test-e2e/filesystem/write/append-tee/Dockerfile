FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Prepare target file to append to
RUN sh -lc "echo base > /tmp/append-target"

ENTRYPOINT ["bash"]
CMD ["/entrypoint.sh"]
