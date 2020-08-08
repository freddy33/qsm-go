FROM golang:1.13 as BUILDER
WORKDIR /app
COPY . .
WORKDIR /app/backend
RUN go mod download
RUN GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main

FROM postgres:12.3-alpine
USER postgres
WORKDIR /app
COPY --from=BUILDER /app/backend/main main
ENTRYPOINT ["./main", "filldb"]