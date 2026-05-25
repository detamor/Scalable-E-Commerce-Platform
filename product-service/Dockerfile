FROM golang:1.26-alpine AS builder

WORKDIR /app

ENV GOPROXY=https://goproxy.cn,https://goproxy.io,https://proxy.golang.org,direct

RUN apk --no-cache add git

COPY go.mod ./
COPY go.sum* ./
RUN go mod download

COPY . .
RUN go mod tidy
RUN CGO_ENABLED=0 GOOS=linux go build -o /product-service ./cmd/api

FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /product-service .
COPY migrations/ ./migrations/

EXPOSE 8082
CMD ["./product-service"]
