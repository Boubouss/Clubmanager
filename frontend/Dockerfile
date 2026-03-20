FROM golang:1.25-alpine AS builder

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN apk add --no-cache protobuf nodejs npm
RUN npm install -g sass

WORKDIR /app/src

COPY ./frontend . 
COPY ./proto ./api/grpc/proto

RUN go mod download
RUN templ generate
RUN sass ./public/scss/main.scss ./public/css/main.css
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
COPY --from=builder app/src/public/css ./public/css
COPY --from=builder app/src/public/js ./public/js

ENTRYPOINT ["/clubmanager"]
