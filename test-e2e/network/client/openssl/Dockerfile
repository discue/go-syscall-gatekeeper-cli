FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update && apt-get install -y openssl coreutils ca-certificates && rm -rf /var/lib/apt/lists/*

# Timeout to avoid hanging on stdin; require cert verification to enforce FS-read dependency
CMD ["openssl", "s_client", "-connect", "example.com:443", "-servername", "example.com"]
