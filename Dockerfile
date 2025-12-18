# 构建镜像，基于golang:1.24.4-alpine
FROM golang:1.24.4-alpine AS builder

WORKDIR /workspace/wcd

COPY . .

RUN go mod tidy && go build -o app

# 运行时镜像，基于alpine
FROM alpine:3.20

WORKDIR /app

RUN mkdir -p logs
COPY --from=builder /workspace/wcd/app .
COPY --from=builder /workspace/wcd/main.go .
COPY --from=builder /workspace/wcd/conf ./conf
COPY --from=builder /workspace/wcd/static ./static
COPY --from=builder /workspace/wcd/data ./data

EXPOSE 8080

CMD ["./app"]