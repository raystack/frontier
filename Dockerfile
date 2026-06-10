FROM alpine:3.21

RUN adduser -D -h /home/frontier frontier
COPY --chown=frontier frontier /usr/bin/frontier
USER frontier
ENTRYPOINT ["frontier"]
