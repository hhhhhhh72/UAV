<!-- Generated: 2026-08-19 | Files scanned: 12+ | Token estimate: ~400 -->

# 依赖与外部服务地图

## 运行时依赖

| 依赖 | 用途 | 备注 |
|------|------|------|
| PostgreSQL 16 | 生产存储 | docker-compose 端口 **5433:5432**，用户 drone |
| 微信开放平台 | 小程序登录 code2Session | WECHAT_APPID/APPSECRET（manifest AppID `wx10842887836afd68`） |
| 电子签服务 | 合同 webhook 回调 | SIGNING_SECRET HMAC 校验 + event_id 去重（对接方未实指） |
| 文件存储 | 上传文件 /uploads/{file_id} | 本地磁盘，10MB 限制；OSS/COS 待接入（PRD Q-5） |
| 短信通道 | 验证码 | 未接入（PRD Q-4，待签名模板） |

## 后端 Go 依赖（go.mod）

标准库为主：net/http（Go 1.22+ 路由模式）、database/pgx(v5/pgxpool)、golang.org/x/crypto(bcrypt)、swaggo（dev 文档生成）。

## 前端依赖

| 项目 | 关键依赖 |
|------|---------|
| frontend/ | vue3.4 + vue-router4 + pinia2(未用) + axios + echarts6/vue-echarts8 + @arco-design/web-vue2.58 + vue-cropper |
| miniprogram/ | uni-app（HBuilderX 编译链，无 npm 依赖）+ 自研 14 个 u- 组件 |

## 部署与 CI

```
Dockerfile      多阶段：golang:1.25-alpine 构建 → alpine:3.21 运行（TZ=Asia/Shanghai）
docker-compose  db(postgres:16-alpine, 512M) + api(256M, uploads卷, depends_on db healthy)
CI (ci.yml)     build: go build+vet+test+race | integration: 容器 PG 跑 repository 集成测试
nginx           (生产) 前端静态 + /api 转发 172.17.0.1:8090
```

⚠️ docker-compose 中 AUTH_SECRET/ENCRYPTION_KEY 为占位符；CORS_ORIGINS 端口与前端实际(5173)不一致。

## 工具脚本

| 脚本 | 用途 |
|------|------|
| run_api.bat / cmd/cli/main.go | Windows 开发启动（构建+设 PG 连接+开浏览器） |
| scripts/fix_prd.py, fix_prd_tree.py | 一次性 PRD docx 修改工具（未提交） |

## 文档新鲜度（2026-08-19 已清理）

- ✅ README/CLAUDE.md — 已同步 Arco Design Vue、小程序 103 页、63 组迁移、443 条路由
- ✅ docs/数据设计/数据模型.md — 已更新 63 组迁移 + 表分组补齐
- ✅ CLAUDE.md 死链 — 已改为 README.md / PRD-四人并行开发方案.md
- ℹ️ 小程序开发规格.md / 功能方案评审.md / 功能方案修订版.md — 已由用户删除（过时需求文档）
