FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update \
    && apt-get install -y --no-install-recommends socat ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Use IPv4 address to avoid DNS; timeout to fail fast on denied connects
CMD ["socat", "-", "TCP4:www.example.com:80,keepidle=10,connect-timeout=3"]
