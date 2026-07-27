#!/usr/bin/env bash
# ============================================================
#  无人机产业综合服务平台 — 停止所有服务 (Linux/macOS/Git Bash)
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
echo -e "${CYAN}  ║    无人机产业综合服务平台 — 停止所有服务      ║${NC}"
echo -e "${CYAN}  ╚══════════════════════════════════════════════╝${NC}"
echo ""

# ---- 1. 用 PID 文件停止 ----------------------------------------
echo -e " ${GREEN}[1/4]${NC} 停止 Go API (PID 文件)..."

if [ -f .api.pid ]; then
  PID=$(cat .api.pid)
  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null && echo "        终止 Go API PID=$PID" || true
  fi
  rm -f .api.pid
fi
echo "        完成"
echo ""

# ---- 2. 停止 Vue 前端 (Vite) -----------------------------------
echo -e " ${GREEN}[2/4]${NC} 停止 Vue 前端 (PID 文件)..."

if [ -f .vite.pid ]; then
  PID=$(cat .vite.pid)
  if kill -0 "$PID" 2>/dev/null; then
    kill "$PID" 2>/dev/null && echo "        终止 Vite PID=$PID" || true
  fi
  rm -f .vite.pid
fi
echo "        完成"
echo ""

# ---- 3. 按端口杀进程 (兜底) ------------------------------------
echo -e " ${GREEN}[3/4]${NC} 清理端口占用..."

# 杀 8080 端口 (Go API)
lsof -ti:8080 2>/dev/null | xargs kill -9 2>/dev/null && echo "        终止端口 8080 进程" || true

# 杀 5173 端口 (Vite)
lsof -ti:5173 2>/dev/null | xargs kill -9 2>/dev/null && echo "        终止端口 5173 进程" || true

# 杀 drone-platform 进程
pkill -f "drone-platform" 2>/dev/null && echo "        终止 drone-platform 进程" || true

echo "        完成"
echo ""

# ---- 4. 停止 PostgreSQL (docker-compose) -----------------------
echo -e " ${GREEN}[4/4]${NC} 停止 PostgreSQL..."

docker compose down 2>/dev/null || docker-compose down 2>/dev/null || true
echo "        完成"
echo ""

echo -e " ${CYAN}╔══════════════════════════════════════════════╗${NC}"
echo -e " ${CYAN}║           所有服务已停止 ✓                   ║${NC}"
echo -e " ${CYAN}╚══════════════════════════════════════════════╝${NC}"
echo ""
