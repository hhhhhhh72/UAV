# Build stage (可选，如果您想在 Docker 里构建)
# FROM node:18 AS build-stage
# WORKDIR /app
# COPY package*.json ./
# RUN npm install
# COPY . .
# RUN npm run build

# Production stage
FROM nginx:alpine

# 复制构建好的文件到 Nginx 目录
# 假设您已经在本地运行了 npm run build，这里直接复制 dist
COPY dist /usr/share/nginx/html

# 复制自定义 Nginx 配置
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
