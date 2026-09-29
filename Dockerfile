# syntax=docker/dockerfile:1

FROM golang:1.23-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY cmd ./cmd
COPY internal ./internal
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags='-s -w' -o /authenticator ./cmd/authenticator

FROM alpine:3.20
RUN adduser -D -u 10001 cube
USER cube
COPY --from=build /authenticator /usr/local/bin/cube-authenticator
ENTRYPOINT ["/usr/local/bin/cube-authenticator"]
