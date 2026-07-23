# Docker 部署

## 容器架构

```
docker-compose.yml
  ├── api (Go)     → :8080
  └── db (PG 15)   → :5432
```

## 快速部署

```bash
docker-compose up -d
```

## 镜像构建

```dockerfile
# 多阶段构建
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build -o api ./cmd/api

FROM alpine:3.19
COPY --from=builder /app/api /api
COPY migrations/ /migrations/
EXPOSE 8080
CMD ["/api"]
```

## CI/CD

GitHub Actions: build + vet + test (每次push)
```yaml
# .github/workflows/ci.yml
```
