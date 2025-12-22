FROM debian

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN apt-get update && apt-get install -y curl python3 && rm -rf /var/lib/apt/lists/*

CMD ["./entrypoint.sh"]