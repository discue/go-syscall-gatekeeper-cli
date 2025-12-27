FROM debian:stable-slim

COPY gatekeeper /gatekeeper
RUN chmod +x /gatekeeper
# Try fallocate (util-linux). Also do a cp that may use copy_file_range depending on kernel/libc.
RUN sh -lc "echo data > /tmp/srcfile"

CMD ["/bin/sh", "-lc", "fallocate -l 1024 /tmp/allocfile && cp /tmp/srcfile /tmp/dstfile"]
