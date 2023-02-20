FROM golang:1.19 AS builder

WORKDIR /src/clacksy

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w -extldflags '-static'" -tags osusergo,netgo -o /usr/local/bin/clacksy ./cmd/clacksy

FROM alpine

COPY --from=flyio/litefs:0.3 /usr/local/bin/litefs /usr/local/bin/litefs
COPY --from=builder /usr/local/bin/clacksy /usr/local/bin/clacksy

ADD etc/litefs.yml /etc/litefs.yml

RUN apk add bash fuse sqlite ca-certificates curl ffmpeg

ENTRYPOINT litefs mount -- clacksy -dsn /litefs/clacksy
