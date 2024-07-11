FROM golang:1.21-alpine as builder

WORKDIR /app

COPY . .

RUN go build -o /app/go_auth

FROM scratch

COPY --from=builder /app/go_auth /app/go_auth

ENTRYPOINT ["/app/go_auth"]