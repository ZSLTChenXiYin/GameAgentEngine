#!/bin/bash
# GameAgentEngine 多平台编译打包脚本
# 用法:
#   ./build.sh                       编译当前平台
#   ./build.sh windows/amd64         编译指定平台
#   ./build.sh linux/amd64 darwin/arm64  编译多个平台
#   ./build.sh all/all               自动编译全部平台

set -euo pipefail

cd "$(dirname "$0")/../.." || { echo "Failed to locate project root"; exit 1; }

# ============ CONFIG ============
ALL_PLATFORMS=(
  "windows/amd64"
  "windows/arm64"
  "linux/amd64"
  "linux/arm64"
  "darwin/amd64"
  "darwin/arm64"
)
SOURCE_DIR="tools/source"
OUTPUT_DIR="dist"
VERSION="v0.4.3"
# ==============================

detect_os="$(uname -s | tr '[:upper:]' '[:lower:]')"
detect_arch="$(uname -m)"
case "$detect_os" in
  linux)   GOOS="linux" ;;
  darwin)  GOOS="darwin" ;;
  mingw*|msys*|cygwin*) GOOS="windows" ;;
  *)       GOOS="linux" ;;
esac
case "$detect_arch" in
  x86_64|amd64) GOARCH="amd64" ;;
  i386|i686)    GOARCH="386" ;;
  aarch64|arm64) GOARCH="arm64" ;;
  *)            GOARCH="amd64" ;;
esac
CURRENT_PLATFORM="${GOOS}/${GOARCH}"

if [ $# -eq 0 ]; then
  TARGETS=("$CURRENT_PLATFORM")
else
  TARGETS=()
  for arg in "$@"; do
    if [ "$arg" = "all/all" ] || [ "$arg" = "all" ]; then
      TARGETS=("${ALL_PLATFORMS[@]}")
      break
    else
      TARGETS+=("$arg")
    fi
  done
fi

LDFLAGS="-s -w -X github.com/ZSLTChenXiYin/GameAgentEngine/internal/version.Version=${VERSION} -X github.com/ZSLTChenXiYin/GameAgentEngine/cmd/gameagentdevcli.devCliVersion=${VERSION}"

echo "========================================="
echo " GameAgentEngine Build Script"
echo " Version: ${VERSION}"
echo " Current: ${CURRENT_PLATFORM}"
echo " Targets: ${TARGETS[*]}"
echo " Output:  ${OUTPUT_DIR}/"
echo "========================================="
echo ""

for target in "${TARGETS[@]}"; do
  GOOS="${target%%/*}"
  GOARCH="${target##*/}"
  OUT="${OUTPUT_DIR}/GameAgentEngine-${GOOS}-${GOARCH}-${VERSION}"
  EXT=""
  [ "$GOOS" = "windows" ] && EXT=".exe"

  mkdir -p "$OUT"

  echo "[${GOOS}/${GOARCH}] Building GameAgentEngine..."
  GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="${LDFLAGS}" \
    -o "${OUT}/GameAgentEngine${EXT}" \
    ./cmd/gameagentengine/

  echo "[${GOOS}/${GOARCH}] Building GameAgentDevCli..."
  GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="${LDFLAGS}" \
    -o "${OUT}/GameAgentDevCli${EXT}" \
    ./cmd/gameagentdevcli/

  # Inject version into Creator JS before copying
  if [ -f "${SOURCE_DIR}/web/GameAgentCreator/js/version.js" ]; then
    sed -i.bak "s/CREATOR_MIN_COMPATIBLE = \"v[0-9]\+\.[0-9]\+\.[0-9]\+\"/CREATOR_MIN_COMPATIBLE = \"${VERSION}\"/" "${SOURCE_DIR}/web/GameAgentCreator/js/version.js"
    rm -f "${SOURCE_DIR}/web/GameAgentCreator/js/version.js.bak"
  fi

  if [ -f "gameagentengine.conf.yaml" ]; then
    cp gameagentengine.conf.yaml "${OUT}/"
  fi
  if [ -d "${SOURCE_DIR}" ]; then
    cp -r "${SOURCE_DIR}/"* "${OUT}/" 2>/dev/null || true
  fi

  echo "[${GOOS}/${GOARCH}] -> ${OUT}/"
  ls -lh "${OUT}/"
  echo ""

  echo "[${GOOS}/${GOARCH}] Packaging..."
  zip -r "${OUT}.zip" "${OUT}/" > /dev/null 2>&1
done

echo "========================================="
echo " Build complete."
for target in "${TARGETS[@]}"; do
  echo "  ${OUTPUT_DIR}/GameAgentEngine-${target%%/*}-${target##*/}-${VERSION}/"
done
echo "========================================="


