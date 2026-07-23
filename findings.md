# 前端研究发现

## 设计参考
- 重庆市无人机产业协会 — 官网: https://www.cq-uav.com
- 低空经济生态服务平台 — 产业级工具，非消费级应用
- 用户群: 180+会员单位(企业/院校/政府部门)

## 技术决策
- 微信原生 > uni-app (用户偏好)
- Vant Weapp: 最成熟的微信小程序组件库
- icon: 已有15个SVG在icons/目录

## API对前端的影响
- 212条端点全在后端
- 公开GET无需认证
- POST/PUT/DELETE需Bearer Token
- 分页统一 page/page_size
