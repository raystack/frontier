FROM alpine:3.13

COPY shield /usr/bin/shield

EXPOSE 8080
ENTRYPOINT ["shield", "serve", "proxy"]