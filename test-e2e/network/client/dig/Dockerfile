FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update && apt-get install -y dnsutils && rm -rf /var/lib/apt/lists/*

CMD ["dig", "+short", "example.com"]
