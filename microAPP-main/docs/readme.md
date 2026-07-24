# 低空综合服务平台 (Low Altitude Comprehensive Service Platform)

本项目旨在为低空经济提供综合性的服务对接平台，包含 H5 移动端、微信小程序和 Node.js 后端服务。通过用户填报表单递交申请，后续人工联系用户进行业务对接。

## 🚀 快速启动

### 1. 环境准备
- **Node.js**: 版本 >= 16.0
- **npm**: 版本 >= 8.0
- **PostgreSQL**: 版本 >= 14 (可选，默认使用 JSON 文件存储)

### 2. 启动后端服务 (Backend)
后端服务基于 Express，提供 API 接口和文件上传功能。

```bash
cd backend
npm install

# 开发环境启动
npm run dev

# 生产环境启动
npm start

# 或使用 pm2
pm2 start index.js --name "low-altitude-server"
```
后端默认端口: `3000`

### 3. 启动 H5 前端
前端基于 Vue 3 + Vite + Vant 4。

```bash
cd frontend/h5
npm install
npm run dev
```
H5 访问地址: `http://localhost:5173`

### 4. 启动小程序 (微信/支付宝等)
小程序基于 uni-app (Vue 3)，支持多端编译。

```bash
cd frontend/miniprogram
# 使用 HBuilderX 打开项目进行开发
# 或使用 CLI:
npm install -g @dcloudio/uni-cli
npx uni dev:mp-weixin
```

---

## 📱 功能特性

### 核心业务 (12项服务)
| 服务 | 说明 |
|------|------|
| 飞行服务 | 通航/直升机飞行计划申报与预约 |
| 无人机外卖 | 园区/景区内的餐饮即时配送 |
| 无人机物流 | 医疗/急件等物资的点对点运输 |
| 政务巡检 | 森林防火、河道治理等常态化巡查 |
| 无人机托管 | 设备存储、保养与充放电管理 |
| 无人机吊运 | 山区/工地等复杂地形物资吊装 |
| 无人机表演 | 节日庆典、品牌宣传编队飞行 |
| 无人机培训 | CAAC/AOPA 执照培训与行业进修 |
| 无人机租赁 | 多机型短租/长租服务 |
| 低空研学 | 青少年航空科普与基地参观 |
| 无人机二手交易 | 设备评估、置换与交易担保 |
| 无人机金融服务 | 分期购买、融资租赁与商业保险 |

### 用户功能
- **服务大厅**: 快捷服务入口、精选案例、平台优势展示
- **全部服务**: 12项服务网格展示与搜索
- **我的申请**: 申请记录状态跟踪（待处理/处理中/已完成）
- **个人中心**: 用户信息管理、实名认证入口
- **登录注册**: 手机号验证码登录流程

### 后台管理
- **案例管理**: 支持管理员上传/编辑/删除案例
- **申请管理**: 查看并处理用户提交的服务申请

---

## 📂 项目结构

```
.
├── backend/                 # 后端项目 (Express + Node.js)
│   ├── db/                  # 数据库相关 (PostgreSQL 迁移脚本)
│   ├── public/              # 静态资源 (uploads, assets)
│   ├── index.js             # 服务入口
│   ├── storage.js           # 数据存储封装
│   └── 部署说明.md          # 后端部署文档
├── frontend/                # 前端项目
│   ├── h5/                  # H5 移动端 (Vue 3 + Vite + Vant 4)
│   │   ├── src/
│   │   │   ├── views/       # 页面组件
│   │   │   ├── router/      # 路由配置
│   │   │   └── styles/      # 全局样式
│   │   └── vite.config.js   # Vite 配置
│   └── miniprogram/         # 小程序 (uni-app Vue 3)
│       ├── pages/           # 页面
│       ├── components/      # 公共组件
│       └── static/          # 静态资源
├── docs/                    # 项目文档
│   ├── readme.md            # 本说明文件
│   ├── 用户体系完善指南.md
│   └── ...
└── sql/                     # 数据库脚本
    └── schema.sql           # 表结构定义
```

---

## 🛠 技术栈

| 端 | 技术 |
|----|------|
| **H5 前端** | Vue 3, Vite 5, Vant 4, Axios, Pinia, Vue Router, Leaflet |
| **小程序** | uni-app (Vue 3), 支持微信/支付宝/百度/字节小程序 |
| **后端** | Node.js, Express, Multer, JWT, bcryptjs |
| **数据库** | PostgreSQL / JSON 文件存储 (LowDB) |
| **部署** | Linux (Ubuntu), PM2, Nginx |

---

## 📄 更多文档

详细的需求说明、部署指南等文件请查看 `docs/` 和 `backend/部署说明.md`。
