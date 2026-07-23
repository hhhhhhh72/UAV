#!/bin/bash
# 无人机平台 — 开发环境启动 (Linux/macOS/Git Bash)
set -e

# 加载 .env
[ -f .env ] && export $(grep -v '^#' .env | xargs)

# 默认值
export AUTH_SECRET="${AUTH_SECRET:-drone-platform-dev-secret-32bytes!}"
export ADMIN_DEV_MODE="${ADMIN_DEV_MODE:-true}"
export ENV="${ENV:-development}"
export HTTP_ADDR="${HTTP_ADDR:-:8080}"

show_help() {
  echo "用法: bash start.sh [api|mp|all]"
  echo "  api   启动后端 API (端口 8080)"
  echo "  mp    初始化小程序开发环境"
  echo "  all   后端 + 小程序全部就绪"
  echo ""
  echo "管理后台: http://localhost:8080/admin"
  echo "OpenAPI:  docs/openapi.yaml"
}

case "${1:-}" in
  --help|-h)
    show_help; exit 0 ;;

  mp|--mp)
    echo "=== 小程序开发环境 ==="
    cd miniprogram
    if [ ! -d node_modules ]; then
      echo "[1/2] npm install..."
      npm install
    else
      echo "[1/2] npm 依赖已存在"
    fi
    cd ..
    echo "[2/2] 请手动打开微信开发者工具 → 导入 miniprogram 目录"
    echo "后端地址: http://localhost:8080 (需先启动 start.sh api)"
    ;;

  all|--all)
    echo "=== 全环境启动 ==="
    cd miniprogram && [ ! -d node_modules ] && npm install && cd ..
    echo "[编译] go build..."
    go build -o drone-api ./cmd/api
    echo "[启动] http://localhost:8080"
    echo "[小程序] 微信开发者工具 → 导入 miniprogram 目录"
    ./drone-api
    ;;

  api|--api|*)
    echo "=== 无人机平台 API ==="
    echo "[编译] go build..."
    go build -o drone-api ./cmd/api
    echo "[启动] http://localhost:8080"
    echo "[管理后台] http://localhost:8080/admin"
    echo "[健康检查] http://localhost:8080/healthz"
    echo "按 Ctrl+C 停止"
    ./drone-api
    ;;
esac
