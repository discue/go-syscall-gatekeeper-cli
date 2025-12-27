FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update \
    && apt-get install -y --no-install-recommends socat ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Use IPv4 address to avoid DNS; timeout to fail fast on denied connects
CMD ["bash", "-lc", "printf 'GET / HTTP/1.0\r\n\r\n' | socat - TCP:1.1.1.1:80,connect-timeout=3 | sed -n '1p'"]
