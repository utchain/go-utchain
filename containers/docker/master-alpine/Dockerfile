FROM alpine:3.7

RUN \
  apk add --update go git make gcc musl-dev linux-headers ca-certificates && \
  git clone --depth 1 --branch release/1.8 https://github.com/utchain/go-utchain && \
  (cd go-utchain && make gut) && \
  cp go-utchain/build/bin/gut /gut && \
  apk del go git make gcc musl-dev linux-headers && \
  rm -rf /go-utchain && rm -rf /var/cache/apk/*

EXPOSE 8545
EXPOSE 30303

ENTRYPOINT ["/gut"]
