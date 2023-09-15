FROM golang:1.20 as builder

WORKDIR /src/
COPY . .

RUN go build

FROM debian:stable-slim

COPY --from=builder /src/fly /fly

ENTRYPOINT /fly
