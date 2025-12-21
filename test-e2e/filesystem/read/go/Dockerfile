FROM golang

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY read.go /read.go
RUN go build -o /read /read.go
RUN chmod +x /read

CMD ["/read"]