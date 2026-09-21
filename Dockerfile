# ---- Build stage ----
FROM golang:1.25-alpine AS builder

# Alpine 源换国内镜像：dl-cdn 从这台机器下载**每个包要 60–190 秒**（2026-09-21 实测），
# 一次构建 12 个包 → 十几分钟；中途一抖动整个发布就失败（当天真的发生过一次，
# 报 apk add exit 1）。阿里云镜像实测 200 且快两个数量级。
RUN sed -i 's#dl-cdn.alpinelinux.org#mirrors.aliyun.com#g' /etc/apk/repositories && \
    apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go env -w GOPROXY=https://goproxy.cn,direct && go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /app ./cmd/api

# ---- Run stage ----
FROM alpine:3.21

RUN sed -i 's#dl-cdn.alpinelinux.org#mirrors.aliyun.com#g' /etc/apk/repositories && \
    apk add --no-cache ca-certificates tzdata
ENV TZ=Asia/Shanghai

# 安全加固（审计 H1）：非 root 运行 + 最小权限——容器被 RCE 后爆炸半径受限于 appuser
RUN addgroup -S app && adduser -S appuser -G app && \
    mkdir -p /uploads /logs /migrations && \
    chown -R appuser:app /uploads /logs /migrations

COPY --from=builder /app /app
# 迁移目录：MigrationsDir() 优先读 MIGRATIONS_DIR 环境变量（-trimpath 下 runtime.Caller 推导不可靠）
ENV MIGRATIONS_DIR=/migrations
COPY migrations/ /migrations/
# 种子图片（000051 迁移引用的 sl-*.jpg；uploads 卷首次挂载时 Docker 自动填充）。
# FileService("uploads/") 是相对路径，容器 cwd=/，实际目录是 /uploads
COPY deploy/seed-images/ /uploads/

# 非 root 用户运行（卷仍可写：uploads/logs 已 chown）
USER appuser

EXPOSE 8080
HEALTHCHECK --interval=15s --timeout=3s CMD wget -qO- http://localhost:8080/healthz || exit 1

ENTRYPOINT ["/app"]
