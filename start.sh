#!/usr/bin/env bash
# 无人机产业综合服务平台 — 一键启动 (Linux/macOS/Git Bash)
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
cd "$SCRIPT_DIR"

export ADMIN_DEV_MODE=true
export AUTH_SECRET="dev-secret-32-bytes-long-key-here"
export CORS_ORIGINS="http://localhost:5173"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

cleanup() {
    echo -e "\n${YELLOW}正在停止服务...${NC}"
    kill $API_PID 2>/dev/null
    kill $FRONTEND_PID 2>/dev/null
    echo -e "${GREEN}已停止${NC}"
    exit 0
}
trap cleanup SIGINT SIGTERM

echo ""
echo "  ============================================"
echo "     无人机产业综合服务平台 v1.0"
echo "  ============================================"
echo ""

# ── [1/4] 构建后端 ──
echo "[1/4] 构建后端..."
if [ ! -f drone-api.exe ] && [ ! -f drone-api ]; then
    echo "  首次运行，正在编译..."
    go build -o drone-api ./cmd/api
    if [ $? -ne 0 ]; then
        echo -e "${RED}  X 编译失败！${NC}"
        exit 1
    fi
    echo -e "${GREEN}  ✓ 编译完成${NC}"
fi
echo -e "${GREEN}  ✓ 后端就绪${NC}"

# ── [2/4] 安装前端依赖 ──
echo "[2/4] 检查前端依赖..."
if [ ! -d "frontend/node_modules" ]; then
    echo "  首次运行，正在安装依赖（约 1-2 分钟）..."
    cd frontend && npm install && cd ..
    echo -e "${GREEN}  ✓ 依赖安装完成${NC}"
fi
echo -e "${GREEN}  ✓ 前端依赖就绪${NC}"

# ── [3/4] 启动后端 ──
echo "[3/4] 启动后端 (http://localhost:8080)..."
./drone-api 2>&1 &
API_PID=$!

echo "  等待后端就绪..."
for i in $(seq 1 30); do
    if curl -s -o /dev/null http://localhost:8080/healthz 2>/dev/null; then
        echo -e "${GREEN}  ✓ 后端已就绪${NC}"
        break
    fi
    sleep 1
done

# ── [4/4] 启动前端 ──
echo "[4/4] 启动前端 (http://localhost:5173)..."
cd frontend && npm run dev &
FRONTEND_PID=$!
cd ..

echo "  等待前端就绪..."
for i in $(seq 1 30); do
    if curl -s -o /dev/null http://localhost:5173 2>/dev/null; then
        echo -e "${GREEN}  ✓ 前端已就绪${NC}"
        break
    fi
    sleep 1
done

echo ""
echo "  ============================================"
echo -e "  ${GREEN}✓ 全部启动完成！${NC}"
echo ""
echo "  后端 API : http://localhost:8080"
echo "  Swagger  : http://localhost:8080/swagger/index.html"
echo "  前端页面 : http://localhost:5173"
echo "  管理后台 : http://localhost:5173/admin"
echo "  ============================================"
echo ""
echo "  按 Ctrl+C 停止所有服务"
echo ""

wait
