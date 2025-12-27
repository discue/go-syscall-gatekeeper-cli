FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper
# Prepare target file to append to
RUN sh -lc "echo base > /tmp/append-target"

# Append via tee -a; exit 0 on success
CMD ["/bin/sh", "-lc", "echo more | tee -a /tmp/append-target > /dev/null"]
