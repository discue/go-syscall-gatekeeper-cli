FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update \
    && apt-get install -y --no-install-recommends inetutils-telnet ca-certificates \
    && rm -rf /var/lib/apt/lists/*

# Connect to an IPv4 address to avoid DNS/FS-read dependency
# Send a minimal HTTP/1.0 request and print the first response line
CMD ["bash", "-lc", "printf 'GET / HTTP/1.0\r\n\r\n' | telnet 1.1.1.1 80"]
