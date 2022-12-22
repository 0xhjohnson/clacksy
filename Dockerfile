FROM golang:1.19 AS builder

WORKDIR /src/clacksy

COPY ./go.mod .
COPY ./go.sum .
RUN go mod download

COPY . .
RUN go build -ldflags "-s -w -extldflags '-static'" -tags osusergo,netgo -o /usr/local/bin/clacksy ./

FROM alpine

COPY --from=builder /usr/local/bin/clacksy /usr/local/bin/clacksy

RUN apk add bash curl

ENTRYPOINT [ "clacksy" ]
