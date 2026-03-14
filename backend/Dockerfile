FROM golang:1.25-alpine AS builder

WORKDIR /app/src

COPY . . 

RUN go mod download

RUN go build -o /clubmanager /app/src/cmd/main.go

FROM alpine:latest 

COPY --from=builder /clubmanager .

RUN wget https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.46/grpc_health_probe-linux-amd64 -O /usr/local/bin/grpc_health_probe && \
    chmod +x /usr/local/bin/grpc_health_probe

EXPOSE "50051"

ENTRYPOINT ["/clubmanager"]
