# Docker 部署

## 容器架构

```
docker-compose.yml
  ├── api (Go)     → :8080
  └── db (PG 16)   → :5432
```

## 快速部署

```bash
docker-compose up -d
```

## 镜像构建

```dockerfile
# 多阶段构建（与实际 Dockerfile 一致）
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o api ./cmd/api

FROM alpine:3.21
COPY --from=builder /app/api /api
COPY migrations/ /migrations/
EXPOSE 8080
CMD ["/api"]
```

> 注意：docker-compose 生产部署**未设置 `ADMIN_DEV_MODE`**（正确行为）。小程序/后台登录走
> 生产路由 `/api/v1/auth/*` 与 `/api/auth/*`（auth 兼容路由已无条件注册，bcrypt 校验）。
> 需配置 `WECHAT_APPID`/`WECHAT_APPSECRET` 启用微信登录，否则密码注册/登录可用、微信静默登录不可用。

## CI/CD

GitHub Actions: build + vet + test (每次push)
```yaml
# .github/workflows/ci.yml
```
