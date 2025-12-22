FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update && apt-get install -y wget && rm -rf /var/lib/apt/lists/*

CMD ["wget", "-q", "-O", "-", "https://www.google.com"]