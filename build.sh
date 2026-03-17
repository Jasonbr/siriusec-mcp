#!/bin/sh
set -e

export CGO_ENABLED=0

APP_NAME="siriusec-mcp"
VERSION=$(grep 'Version.*=' internal/mcp/server.go 2>/dev/null | grep -o '".*"' | tr -d '"' || echo "1.0.0")
TIMESTAMP=$(date +%Y%m%d%H%M%S)
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS="-s -w -X main.Version=${VERSION} -X main.GitCommit=${GIT_COMMIT} -X main.BuildTime=${TIMESTAMP}"

# 颜色输出
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo "${RED}[ERROR]${NC} $1"
}

_package() {
    GOOS_VAL="${1}"
    GOARCH_VAL="${2}"

    log_info "Building ${GOOS_VAL}/${GOARCH_VAL} ..."
    
    OUTPUT_NAME="${APP_NAME}-${GOOS_VAL}-${GOARCH_VAL}"
    if [ "${GOOS_VAL}" = "windows" ]; then
        OUTPUT_NAME="${OUTPUT_NAME}.exe"
    fi
    
    GOOS=${GOOS_VAL} GOARCH=${GOARCH_VAL} go build -ldflags "${LDFLAGS}" -o "bin/${OUTPUT_NAME}" ./cmd/server/main.go

    RELEASE_DIR="release/${APP_NAME}-${VERSION}-${GOOS_VAL}-${GOARCH_VAL}-${TIMESTAMP}"
    mkdir -p "${RELEASE_DIR}"
    cp "bin/${OUTPUT_NAME}" "${RELEASE_DIR}/${APP_NAME}"
    
    # 复制配置文件
    cp -r k8s "${RELEASE_DIR}/" 2>/dev/null || true
    cp Dockerfile "${RELEASE_DIR}/" 2>/dev/null || true
    cp docker-compose.yml "${RELEASE_DIR}/" 2>/dev/null || true
    cp .env.example "${RELEASE_DIR}/" 2>/dev/null || true
    cp README.md "${RELEASE_DIR}/" 2>/dev/null || true
    
    # 创建压缩包
    tar czf "${RELEASE_DIR}.tar.gz" -C release "$(basename ${RELEASE_DIR})"
    rm -rf "${RELEASE_DIR}"

    log_info "Done: ${RELEASE_DIR}.tar.gz"
}

build_local() {
    log_info "Building for local platform ..."
    go build -ldflags "${LDFLAGS}" -o "bin/${APP_NAME}" ./cmd/server/main.go
    log_info "Done: bin/${APP_NAME}"
}

build_linux_amd64() {
    _package linux amd64
}

build_linux_arm64() {
    _package linux arm64
}

build_darwin_amd64() {
    _package darwin amd64
}

build_darwin_arm64() {
    _package darwin arm64
}

build_windows_amd64() {
    _package windows amd64
}

build_all() {
    log_info "Building all platforms ..."
    build_linux_amd64
    build_linux_arm64
    build_darwin_amd64
    build_darwin_arm64
    build_windows_amd64
    log_info "All builds completed!"
}

clean() {
    log_info "Cleaning build artifacts ..."
    rm -rf bin/ release/
    log_info "Clean done"
}

test() {
    log_info "Running tests ..."
    go test ./... -v
    log_info "Tests completed"
}

lint() {
    log_info "Running linter ..."
    go vet ./...
    log_info "Lint completed"
}

version() {
    echo "Version: ${VERSION}"
    echo "Git Commit: ${GIT_COMMIT}"
    echo "Build Time: ${TIMESTAMP}"
}

usage() {
    echo "Usage: $0 {local|linux-amd64|linux-arm64|darwin-amd64|darwin-arm64|windows-amd64|all|clean|test|lint|version}"
    echo ""
    echo "Commands:"
    echo "  local         Build for current platform"
    echo "  linux-amd64   Cross-compile for linux/amd64"
    echo "  linux-arm64   Cross-compile for linux/arm64"
    echo "  darwin-amd64  Cross-compile for darwin/amd64 (macOS Intel)"
    echo "  darwin-arm64  Cross-compile for darwin/arm64 (macOS M1/M2)"
    echo "  windows-amd64 Cross-compile for windows/amd64"
    echo "  all           Build all platforms"
    echo "  clean         Clean build artifacts"
    echo "  test          Run all tests"
    echo "  lint          Run linter"
    echo "  version       Show version info"
    echo ""
    echo "Examples:"
    echo "  $0 local                    # Build for local platform"
    echo "  $0 linux-amd64              # Build for Linux AMD64"
    echo "  $0 all                      # Build all platforms"
    echo "  $0 clean && $0 all          # Clean and rebuild all"
}

# 创建 bin 目录
mkdir -p bin

case "${1}" in
    local)
        build_local
        ;;
    linux-amd64)
        build_linux_amd64
        ;;
    linux-arm64)
        build_linux_arm64
        ;;
    darwin-amd64)
        build_darwin_amd64
        ;;
    darwin-arm64)
        build_darwin_arm64
        ;;
    windows-amd64)
        build_windows_amd64
        ;;
    all)
        build_all
        ;;
    clean)
        clean
        ;;
    test)
        test
        ;;
    lint)
        lint
        ;;
    version)
        version
        ;;
    "")
        build_local
        ;;
    *)
        usage
        exit 1
        ;;
esac
