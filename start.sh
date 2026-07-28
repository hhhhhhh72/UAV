#!/usr/bin/env bash
# ============================================================
#  无人机产业综合服务平台 — 一键启动 (Linux/macOS/Git Bash)
#  启动顺序: PostgreSQL → Go API → Vue Frontend
# ============================================================

set -euo pipefail
cd "$(dirname "$0")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}  ╔══════════════════════════════════════════════╗${NC}"
echo -e "${CYAN}  ║    无人机产业综合服务平台 — 开发环境启动      ║${NC}"
echo -e "${CYAN}  ╚══════════════════════════════════════════════╝${NC}"
echo ""

# ---- 0. 检查必要工具 -------------------------------------------
echo -e " ${GREEN}[1/5]${NC} 检查环境..."

command -v go >/dev/null 2>&1   || { echo -e "${RED} [错误] 未找到 Go，请安装 Go 1.22+${NC}"; exit 1; }
command -v node >/dev/null 2>&1 || { echo -e "${RED} [错误] 未找到 Node.js，请安装 Node.js 18+${NC}"; exit 1; }

echo "        Go:       $(go version | head -c20)"
echo "        Node.js:  $(node -v)"
echo ""

# ---- 1. 启动 PostgreSQL (docker-compose) -----------------------
echo -e " ${GREEN}[2/5]${NC} 启动 PostgreSQL..."

if command -v docker >/dev/null 2>&1; then
  docker compose up -d db 2>/dev/null || docker-compose up -d db 2>/dev/null || true

  echo "        等待 PostgreSQL 就绪..."
  for i in $(seq 1 15); do
    if docker compose exec db pg_isready -U drone -d drone_platform >/dev/null 2>&1; then
      echo -e "        ${GREEN}PostgreSQL 就绪 ✓${NC}"
      break
    fi
    sleep 2
  done
else
  echo -e " ${YELLOW}[警告] 未找到 Docker，跳过数据库启动${NC}"
fi
echo ""

# ---- 2. 启动 Go 后端 API ---------------------------------------
echo -e " ${GREEN}[3/5]${NC} 启动 Go API (http://localhost:8080)..."

go run ./cmd/api &
API_PID=$!
echo "        Go API PID: $API_PID"

# 等待后端就绪
echo "        等待 Go API 就绪..."
for i in $(seq 1 15); do
  if curl -s http://localhost:8080/api/v1/health >/dev/null 2>&1; then
    echo -e "        ${GREEN}Go API 就绪 ✓${NC}"
    break
  fi
  sleep 1
done
echo ""

# ---- 3. 启动 Vue 前端 (Vite) -----------------------------------
echo -e " ${GREEN}[4/5]${NC} 启动 Vue 前端 (http://localhost:5173)..."

# 安装依赖 (首次)
if [ ! -d "frontend/node_modules" ]; then
  echo "        首次运行，安装前端依赖..."
  (cd frontend && npm install)
fi

(cd frontend && npm run dev) &
VITE_PID=$!
echo "        Vite PID: $VITE_PID"

echo "        等待 Vite 就绪..."
for i in $(seq 1 20); do
  if curl -s http://localhost:5173 >/dev/null 2>&1; then
    echo -e "        ${GREEN}Vue 前端就绪 ✓${NC}"
    break
  fi
  sleep 1
done
echo ""

# ---- 4. 保存 PID 文件 ------------------------------------------
echo "$API_PID" > .api.pid
echo "$VITE_PID" > .vite.pid

echo -e " ${GREEN}[5/5]${NC} 启动完成！"
echo ""
echo -e "  ${CYAN}╔══════════════════════════════════════════════╗${NC}"
echo -e "  ${CYAN}║  服务                     地址                ║${NC}"
echo -e "  ${CYAN}╠══════════════════════════════════════════════╣${NC}"
echo -e "  ${CYAN}║${NC}  后台管理 (推荐入口)      ${GREEN}http://localhost:5173/admin${NC}  ${CYAN}║${NC}"
echo -e "  ${CYAN}║${NC}  Vue 前端                 http://localhost:5173          ${CYAN}║${NC}"
echo -e "  ${CYAN}║${NC}  Go API                   http://localhost:8080          ${CYAN}║${NC}"
echo -e "  ${CYAN}║${NC}  PostgreSQL               localhost:5433                 ${CYAN}║${NC}"
echo -e "  ${CYAN}╚══════════════════════════════════════════════╝${NC}"
echo ""
echo -e "  停止所有服务: ${YELLOW}bash stop.sh${NC}"
echo ""

# 等待用户 Ctrl+C
wait
