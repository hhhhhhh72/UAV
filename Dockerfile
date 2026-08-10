# ---- Build stage ----
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app ./cmd/api

# ---- Run stage ----
FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

COPY --from=builder /app /app
# 迁移目录：MigrationsDir() 优先读 MIGRATIONS_DIR 环境变量（-trimpath 下 runtime.Caller 推导不可靠）
ENV MIGRATIONS_DIR=/migrations
COPY migrations/ /migrations/
# 种子图片（000051 迁移引用的 sl-*.jpg；uploads 卷首次挂载时 Docker 自动填充）。
# FileService("uploads/") 是相对路径，容器 cwd=/，实际目录是 /uploads
COPY deploy/seed-images/ /uploads/

EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app"]
