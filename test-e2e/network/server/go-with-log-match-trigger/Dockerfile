FROM golang

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY server.go /server.go
RUN go build -o /server /server.go
RUN chmod +x /server

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

RUN curl --help

ENTRYPOINT ["bash"]
CMD ["/entrypoint.sh"]