# 运维记录 (Operations and Maintenance)

> 更新日期：2026-01-20  
> 项目：低空综合服务平台

---

## 一、服务器信息

| 项目 | 值 |
|------|-----|
| 域名 | microapp.zndkfx.com |
| 服务器 IP | 39.172.120.254 |
| 部署目录 | /data/low-altitude |
| 后端目录 | /data/low-altitude/backend |
| 后端端口 | 8090 |
| PM2 进程名 | low-altitude-server |

---

## 二、2026-01-20 部署记录

### 2.1 本次更新内容

1. **更新畅行温州 SSO 测试环境配置**
   - 请求地址：`https://dev.jieyisoft.com:11296`
   - SM2 私钥：`a180a91baed06c50a92699d0fd6ca03412ad9f246a643396a07957b8933de643`
   - SM4 密钥：`67651926067651926067651926067651`（32位hex，直接使用）

2. **修复 SM4 密钥格式问题**
   - 支持32位hex字符串直接使用
   - 支持16位UTF-8字符串自动转换

3. **新增文件**
   - `backend/test-platform.js` - 平台连通性测试脚本
   - `docs/implementation/用户体系完善工作计划.md` - 工作计划文档

### 2.2 部署步骤

```bash
# 1. 进入项目目录
cd /data/low-altitude

# 2. 备份并拉取代码
git stash
git pull

# 3. 进入后端目录
cd backend

# 4. 安装依赖（重要：确保所有依赖安装完整）
rm -rf node_modules package-lock.json
npm install

# 如果缺少依赖，手动安装：
npm install axios cors sm-crypto pg xlsx lowdb multer express body-parser jsonwebtoken bcryptjs

# 5. 重启服务
pm2 delete low-altitude-server
PORT=8090 pm2 start index.js --name "low-altitude-server" --cwd /data/low-altitude/backend
pm2 save

# 6. 验证服务
pm2 status
pm2 logs low-altitude-server --lines 5
curl http://localhost:8090/api/services/config
```

### 2.3 遇到的问题

| 问题 | 原因 | 解决方案 |
|------|------|---------|
| git pull 冲突 | 服务器有本地修改的数据文件 | `git stash` 暂存后再 pull |
| MODULE_NOT_FOUND | npm install 未完整安装依赖 | 手动安装缺失的包 |
| PM2 使用旧配置 | PM2 缓存了旧的工作目录 | 删除进程后用 `--cwd` 重新创建 |

---

## 三、畅行温州 SSO 对接

### 3.1 SSO 入口地址

给畅行温州配置的入口地址：
```
https://microapp.zndkfx.com/sso/login
```

参数名：`jyauthcode`

完整 URL 格式：
```
https://microapp.zndkfx.com/sso/login?jyauthcode={authcode}
```

### 3.2 SSO 流程

1. 用户在畅行温州App点击入口
2. 畅行温州App获取临时授权码
3. WebView 打开 `https://microapp.zndkfx.com/sso/login?jyauthcode=xxx`
4. 后端重定向到 `/#/home?authcode=xxx`
5. 前端检测 authcode，调用 `/api/sso/login`
6. 后端验证授权码，返回用户信息和 JWT
7. 用户自动登录

### 3.3 测试环境配置

> 更新日期：2026-01-20

| 配置项 | 值 |
|--------|-----|
| 平台地址 | https://dev.jieyisoft.com:11296 |
| joininstid | 00000015 |
| authinstid | 00000015 |
| instid | 00000001 |
| mchntid | 150000150001 |
| chnlid | 03 |
| SM2 私钥 | cdcd3db6845a1457895328a52e109646707c6bf372ef44db69d4390989b9a5ed |
| SM4 密钥 | 12545612345648907234561434557894 |

### 3.4 连通性测试

```bash
cd /data/low-altitude/backend
node test-platform.js

# 带 authcode 测试
node test-platform.js <authcode>
```

---

## 四、常用运维命令

### 4.1 服务管理

```bash
# 查看服务状态
pm2 status

# 查看日志
pm2 logs low-altitude-server --lines 20

# 只看错误日志
pm2 logs low-altitude-server --err --lines 20

# 重启服务
pm2 restart low-altitude-server

# 停止服务
pm2 stop low-altitude-server

# 删除服务
pm2 delete low-altitude-server

# 启动服务
PORT=8090 pm2 start index.js --name "low-altitude-server" --cwd /data/low-altitude/backend

# 保存 PM2 配置
pm2 save
```

### 4.2 代码更新

```bash
cd /data/low-altitude
git stash          # 暂存本地修改
git pull           # 拉取最新代码
git stash pop      # 恢复本地修改（可选）

cd backend
npm install        # 安装依赖
pm2 restart low-altitude-server
```

### 4.3 前端构建与部署

```bash
# 本地构建
cd frontend/h5
npm install
npm run build

# 将 dist 内容复制到服务器
# backend/public/ 目录
```

### 4.4 Nginx 配置

```nginx
server {
    listen 443 ssl;
    server_name microapp.zndkfx.com;

    ssl_certificate /etc/nginx/cert/zndkfx.com.pem;
    ssl_certificate_key /etc/nginx/cert/zndkfx.com.key;

    location / {
        proxy_pass http://127.0.0.1:8090;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    }
}
```

---

## 五、待完成事项

- [ ] 服务器依赖安装完成后验证服务正常
- [ ] 前端重新构建并部署
- [ ] 与畅行温州联调测试 SSO
- [ ] Phase 1：管理员登录入口开发

---

## 六、相关文档

- [用户体系完善工作计划](./用户体系完善工作计划.md)
- [畅行温州接入文档](../畅行温州接入文档.md)
- [后端部署说明](../../backend/部署说明.md)
