FROM golang:1.24-alpine as modules
COPY go.mod go.sum /modules/
WORKDIR /modules
RUN go mod download

FROM golang:1.24-alpine as builder
COPY --from=modules /go/pkg go/pkg
RUN go install github.com/pressly/goose/v3/cmd/goose@latest
COPY . /app
WORKDIR /app
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/app ./cmd/app/main.go

FROM alpine:latest

WORKDIR /root/

COPY --from=builder /bin/app .
COPY --from=builder /go/bin/goose /bin/goose 
COPY --from=builder /app/migrations ./migrations

COPY .env .

CMD ["./app"]