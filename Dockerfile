FROM alpine:3.18

RUN apk --no-cache add curl
COPY frontier /usr/bin/frontier

ENTRYPOINT ["frontier"]
