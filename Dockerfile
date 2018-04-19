FROM alpine:3.5

RUN apk add --no-cache ca-certificates

COPY /bin/slackwork /usr/bin/

ENTRYPOINT ["/usr/bin/hydra-wework"]
