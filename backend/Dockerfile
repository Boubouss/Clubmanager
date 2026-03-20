FROM golang:1.25-alpine AS builder

RUN apk add --no-cache protobuf
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

WORKDIR /app/src

COPY ./backend . 
COPY ./proto ./api/grpc/proto

RUN go mod download

RUN protoc \
    --proto_path=api/grpc/proto \
    --go_opt=paths=source_relative \
    --go_out=api/grpc/proto \
    --go-grpc_out=api/grpc/proto \
    --go-grpc_opt=paths=source_relative \
    api/grpc/proto/*.proto

RUN go build -o /clubmanager /app/src/cmd/main.go

FROM alpine:latest 

COPY --from=builder /clubmanager .

RUN wget https://github.com/grpc-ecosystem/grpc-health-probe/releases/download/v0.4.46/grpc_health_probe-linux-amd64 -O /usr/local/bin/grpc_health_probe && \
    chmod +x /usr/local/bin/grpc_health_probe

EXPOSE "50051"

ENTRYPOINT ["/clubmanager"]
