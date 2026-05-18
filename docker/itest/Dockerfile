FROM alpine:3.23

RUN apk add --no-cache jq diffutils

COPY run.sh /usr/local/bin/test.sh

ENTRYPOINT ["test.sh"]
