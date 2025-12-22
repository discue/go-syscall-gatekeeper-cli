FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

RUN apt-get update && apt-get install -y netcat-traditional && rm -rf /var/lib/apt/lists/*

CMD ["nc", "-w", "1", "example.com", "80"]