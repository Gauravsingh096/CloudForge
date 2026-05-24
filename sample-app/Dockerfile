FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY main.go .
RUN go mod init sample-app && \
    CGO_ENABLED=0 GOOS=linux go build -o sample-app .

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/sample-app /sample-app
EXPOSE 8080
ENTRYPOINT ["/sample-app"]
