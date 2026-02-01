FROM golang:1.22-alpine as modules
COPY go.mod go. /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.22-alpine as builder
COPY --from=modules /go/pkg go/pkg
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build -o /bin/app ./cmd/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /bin/app .
COPY .env .

CMD ["./app"]