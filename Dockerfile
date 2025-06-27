FROM alpine:3.18

COPY frontier /usr/bin/frontier

ENTRYPOINT ["frontier"]
