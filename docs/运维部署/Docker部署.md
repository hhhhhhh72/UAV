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

## 部署后必做：修复 .sh 行尾与执行位

Windows 工作区打包的 tar 会把 `.sh` 带成 CRLF 且丢失执行位——
线上曾因此导致**每日备份 cron 连续失败**（`Permission denied` + shebang `#!/bin/bash\r` 无法执行）。
同步源码后必须执行：

```bash
find ~/UAV -name '*.sh' -exec sed -i 's/\r$//' {} +
sudo chmod +x ~/UAV/deploy/*.sh
crontab -l | grep -q 'bash /home/ubuntu/UAV/deploy/db-backup.sh' || echo '请确认 cron 用 bash 显式执行'
```

（备份 cron 已改为 `bash .../db-backup.sh` 显式执行，不再依赖执行位。）

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

## 上传/日志/配置卷所有权（新环境首次部署必做）

api 以非 root 用户（uid=100, gid=101）运行，业务上传写入命名卷 `uploads`、日志写入 `logs`、
平台配置（banner/公告，`PLATFORM_CONFIG_PATH`）写入 `config`。
**命名卷首次创建为 root:root 755，会覆盖镜像内 chown 好的目录**——不初始化所有权会导致
`POST /api/v1/files/upload` 报 500 `create file: ... permission denied`（图片上传/发布全部失败），
以及管理端保存首页 banner/公告配置失败（始终回退代码默认值）。

首次部署（或重建卷后）执行一次：

```bash
docker run --rm \
  -v uav_uploads:/uploads -v uav_logs:/logs -v uav_config:/config \
  alpine chown -R 100:101 /uploads /logs /config
docker restart uav-api-1
docker exec uav-api-1 touch /uploads/write-test && echo WRITE_OK   # 验证
```

卷持久化，执行一次即可；`docker compose down -v` 重建卷后须重新执行。

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
