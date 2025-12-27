FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper
# Prepare file with content, then truncate to size 0
RUN sh -lc "dd if=/dev/zero of=/tmp/trunc-target bs=1 count=16 >/dev/null 2>&1 || true"

CMD ["truncate", "-s", "0", "/tmp/trunc-target"]
