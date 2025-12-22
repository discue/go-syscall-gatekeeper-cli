FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY 1.sh /1.sh
RUN chmod +x /1.sh

COPY 2.sh /2.sh
RUN chmod +x /2.sh

COPY 3.sh /3.sh
RUN chmod +x /3.sh

CMD ["/1.sh"]