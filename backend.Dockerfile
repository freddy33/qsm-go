FROM fredsimon-docker.jfrog.io/golang:1.14 as BUILDER
COPY m3util /app/m3util
COPY model /app/model
COPY backend/go.mod backend/go.sum /app/backend/
WORKDIR /app/backend
RUN GOPROXY="https://fredsimon.jfrog.io/artifactory/api/go/go" go mod download
COPY . /app
RUN GOPROXY="https://fredsimon.jfrog.io/artifactory/api/go/go" GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o main

FROM fredsimon-docker.jfrog.io/alpine:3.12
WORKDIR /app
COPY --from=BUILDER /app/backend/main main
ENTRYPOINT ["./main", "server", "-test"]
