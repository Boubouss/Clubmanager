FROM golang:1.25-alpine AS builder

RUN go install github.com/a-h/templ/cmd/templ@latest
RUN apk add --no-cache nodejs npm
RUN npm install -g sass

WORKDIR /app/src

COPY . . 

RUN go mod download
RUN templ generate
RUN sass ./public/scss/main.scss ./public/css/main.css
RUN go build -o /clubmanager /app/src/cmd/main.go

FROM alpine:latest 

COPY --from=builder /clubmanager .
COPY --from=builder app/src/public/css ./public/css
COPY --from=builder app/src/public/js ./public/js

ENTRYPOINT ["/clubmanager"]
