FROM golang:1.13 as BUILDER
WORKDIR /app
COPY . .
WORKDIR /app/backend
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main

FROM alpine:3.12
WORKDIR /app
COPY --from=BUILDER /app/backend/main main
ENTRYPOINT ["./main", "server"]
