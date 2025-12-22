FROM node:lts

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY server.cjs /server.cjs
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

CMD ["/entrypoint.sh"]