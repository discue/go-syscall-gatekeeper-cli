FROM golang

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper

COPY concurrent-file-reads.go /concurrent-file-reads.go
RUN go build -o /concurrent-file-reads /concurrent-file-reads.go
RUN chmod +x /concurrent-file-reads

CMD ["/concurrent-file-reads"]