ARG GO_VERSION=1.25.1
FROM golang:${GO_VERSION}-alpine AS builder

WORKDIR /bot

RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -o support-bot ./cmd/bot

FROM alpine:3.20

WORKDIR /bot

COPY --from=builder /bot/support-bot ./support-bot

ENTRYPOINT ["./support-bot"]
