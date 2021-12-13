FROM alpine:3.15

COPY shield /usr/bin/shield

EXPOSE 8080
EXPOSE 5556
ENTRYPOINT ["shield"]