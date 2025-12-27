FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

# Install HTTPie via pip to ensure availability across Debian variants
RUN apt-get update \
    && apt-get install -y --no-install-recommends python3-pip ca-certificates \
    && update-ca-certificates \
    && pip3 install --no-cache-dir --break-system-packages httpie \
    && rm -rf /var/lib/apt/lists/*

# Simple HTTPS request to exercise TLS + sockets
CMD ["http", "-v", "https://www.google.com"]
