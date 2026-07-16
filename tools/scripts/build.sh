#!/bin/bash
# GameAgentEngine cross-platform packaging script
# Usage:
#   ./build.sh                       build current platform
#   ./build.sh windows/amd64         build one target
#   ./build.sh linux/amd64 darwin/arm64  build multiple targets
#   ./build.sh all/all               build all supported targets

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
VERSION="v0.4.6"
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

echo "Generating Creator component metadata..."
go run ./tools/scripts/generate_component_meta.go
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

  echo "[${GOOS}/${GOARCH}] Building GameAgentWorker..."
  GOOS=${GOOS} GOARCH=${GOARCH} CGO_ENABLED=0 go build \
    -trimpath \
    -ldflags="${LDFLAGS}" \
    -o "${OUT}/GameAgentWorker${EXT}" \
    ./cmd/gameagentworker/

  if [ -f "gameagentengine.conf.yaml" ]; then
    cp gameagentengine.conf.yaml "${OUT}/"
  fi
  if [ -d "${SOURCE_DIR}" ]; then
    cp -r "${SOURCE_DIR}/"* "${OUT}/" 2>/dev/null || true
  fi

  # Inject version into packaged Creator asset without mutating source files
  if [ -f "${OUT}/web/GameAgentCreator/js/version.js" ]; then
    sed -i.bak "s/CREATOR_MIN_COMPATIBLE = \"v[0-9]\+\.[0-9]\+\.[0-9]\+\"/CREATOR_MIN_COMPATIBLE = \"${VERSION}\"/" "${OUT}/web/GameAgentCreator/js/version.js"
    rm -f "${OUT}/web/GameAgentCreator/js/version.js.bak"
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
