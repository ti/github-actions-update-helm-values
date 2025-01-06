# Build stage
FROM golang:1.23-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod init updater
RUN CGO_ENABLED=0 go build -ldflags '-w -s' -o updater .

FROM alpine:3.21
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/updater .
ENTRYPOINT ["/app/updater"]
