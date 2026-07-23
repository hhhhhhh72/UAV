#!/bin/bash
# 后台管理系统优化后快速启动脚本

echo "========================================="
echo "  低空综合服务平台 - 快速配置"
echo "========================================="
echo ""

# 检查 .env 文件
if [ ! -f .env ]; then
    echo "创建 .env 配置文件..."
    cp .env.example .env
    echo "✓ 已创建 .env 文件"
    echo ""
    echo "⚠  请编辑 .env 文件设置以下关键配置:"
    echo "   - JWT_SECRET (必须)"
    echo "   - DSL_ADMIN_PASSWORD (必须)"
    echo "   - STUDY_ADMIN_PASSWORD (必须)"
    echo ""
    read -p "按任意键继续..."
else
    echo "✓ .env 文件已存在"
    echo ""
fi

# 安装依赖
echo "安装依赖..."
npm install
echo "✓ 依赖安装完成"
echo ""

# 创建日志目录
if [ ! -d logs ]; then
    mkdir logs
    echo "✓ 创建日志目录"
    echo ""
fi

# 创建上传目录
if [ ! -d uploads ]; then
    mkdir uploads
    echo "✓ 创建上传目录"
    echo ""
fi

echo "========================================="
echo "  配置完成!"
echo "========================================="
echo ""
echo "启动服务:"
echo "  开发模式: npm run dev"
echo "  生产模式: npm start"
echo ""
echo "查看日志:"
echo "  tail -f logs/$(date +%Y-%m-%d).log"
echo ""
echo "========================================="
