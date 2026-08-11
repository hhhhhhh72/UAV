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

## 密钥管理（2026-08 起）

**密钥一律禁止硬编码进 compose/git**，从 compose 同目录 `.env` 读取：

| 变量 | 生成方式 | 说明 |
|------|---------|------|
| `AUTH_SECRET` | `openssl rand -hex 32` | Token 签名密钥；轮换仅使已签发 access token 失效，用户自动刷新，无登出风暴 |
| `ENCRYPTION_KEY` | `openssl rand -base64 32` | AES-256-GCM 加密密钥；**轮换必须先重加密存量数据** |

### ENCRYPTION_KEY 轮换流程（生产）

1. 停 API：`docker compose stop api`
2. 全量重加密（5 列：`demands.contact`、`enterprises.license_url`/`account_name`、`users.phone_ciphertext`、`certified_pilots.id_card`）：

```bash
# 干跑确认范围
go build -o /tmp/reencrypt ./cmd/reencrypt
/tmp/reencrypt -dsn "$DATABASE_URL" -old-key "$OLD_KEY" -new-key "$NEW_KEY"
# 正式执行
/tmp/reencrypt -dsn "$DATABASE_URL" -old-key "$OLD_KEY" -new-key "$NEW_KEY" -apply
# 校验（全部密文必须能用新 key 解密；legacy 明文行会报告 undecryptable，属预期）
/tmp/reencrypt -dsn "$DATABASE_URL" -new-key "$NEW_KEY" -old-key "$OLD_KEY" -verify
```

3. 更新 `.env` 的 `ENCRYPTION_KEY`，`docker compose up -d --build api`
4. 验证：管理端 `GET /api/v1/admin/demands` 的 `contact` 应解出明文手机号

> 迁移的 SQL 文件均在启动时执行且**无版本追踪**（必须幂等，`IF NOT EXISTS`）。
> 管理后台账号：手机号 + bcrypt 密码，`role=platform_admin`，经 `POST /api/auth/login` 登录。

## CI/CD

GitHub Actions: build + vet + test (每次push)
```yaml
# .github/workflows/ci.yml
```
