FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update && apt-get install -y curl && rm -rf /var/lib/apt/lists/*

CMD ["curl", "-v", "https://www.google.com"]