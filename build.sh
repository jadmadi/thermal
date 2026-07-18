#!/usr/bin/env bash
#
# build.sh - Build script for thermal CLI
#
# Provides versatile build options across development, release, and ultra-compressed
# modes with automatic version tag injection and cross-compilation support.

set -euo pipefail

# ANSI color codes for terminal output
BOLD=$'\033[1m'
RED=$'\033[31m'
GREEN=$'\033[32m'
YELLOW=$'\033[33m'
CYAN=$'\033[36m'
RESET=$'\033[0m'

show_help() {
    cat <<EOF
${BOLD}thermal Build Script (${CYAN}build.sh${RESET}${BOLD})${RESET}

${BOLD}USAGE:${RESET}
    ./build.sh [MODE/OPTIONS] [TARGETS]

${BOLD}BUILD MODES (Select one, defaults to --release):${RESET}

  ${GREEN}--release${RESET}   ${BOLD}Production / Stripped Build (Recommended Standard)${RESET}
              Command: CGO_ENABLED=0 go build -ldflags="-s -w -X ..." -trimpath -o thermal ./cmd/thermal
              • ${BOLD}Pros:${RESET}  ~35% smaller binary size (~9.8MB vs ~15MB), fully self-contained (CGO_ENABLED=0),
                        strips DWARF symbol tables (-s -w), strips local host paths (-trimpath) for
                        reproducible builds, and auto-injects Git tag/commit/date into version.
              • ${BOLD}Cons:${RESET}  Slightly longer build time (~1s more), production stack traces do not include
                        DWARF local variable debug info or absolute filepaths.

  ${YELLOW}--dev${RESET}       ${BOLD}Development / Debug Build${RESET}
              Command: go build -o thermal ./cmd/thermal
              • ${BOLD}Pros:${RESET}  Fastest compile time, retains full DWARF debug symbols (-g) and symbol table,
                        enables step-by-step debugging via gdb/dlv with exact file and line numbers.
              • ${BOLD}Cons:${RESET}  Large binary size (~15MB), embeds host absolute filepaths inside binary.

  ${CYAN}--upx${RESET}       ${BOLD}Ultra-Compressed Distribution Build (LZMA Compression)${RESET}
              Command: Builds with --release flags, then compresses binary using upx --best --lzma
              • ${BOLD}Pros:${RESET}  Ultra-compact binary size (~2.8MB, an ~80% reduction from dev build!),
                        minimum disk space and lightning-fast network download.
              • ${BOLD}Cons:${RESET}  Requires 'upx' tool installed (sudo apt install upx-ucl), slightly increases
                        cold startup time (~15-20ms) due to in-memory decompression on launch,
                        and compressed binaries can trigger false positives on some antivirus scanners.

${BOLD}OPTIONS & TARGETS:${RESET}
  ${BOLD}-o, --output <path>${RESET}   Specify output binary filepath or directory (default: ./thermal)
  ${BOLD}--target <os/arch>${RESET}    Cross-compile for target (e.g., --target linux/amd64 or --target darwin/arm64)
  ${BOLD}--all${RESET}                 Cross-compile for all major platforms (linux, darwin, windows x amd64, arm64)
                        Outputs to dist/<project>_<version>_<os>_<arch>
  ${BOLD}-h, --help${RESET}            Display this help explanation and exit

${BOLD}EXAMPLES:${RESET}
    ./build.sh --release                      # Build standard stripped release binary for current OS/Arch
    ./build.sh --dev -o thermal-debug         # Build debug binary with symbols named thermal-debug
    ./build.sh --upx                          # Build ultra-small compressed binary (requires upx)
    ./build.sh --release --target darwin/arm64  # Cross-compile release binary for macOS Apple Silicon
    ./build.sh --release --all                # Build release binaries for all primary OS/Arch targets
EOF
}

MODE="release"
OUTPUT=""
TARGET_OS_ARCH=""
BUILD_ALL=false

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --release)
            MODE="release"
            shift
            ;;
        --dev)
            MODE="dev"
            shift
            ;;
        --upx)
            MODE="upx"
            shift
            ;;
        -o|--output)
            OUTPUT="$2"
            shift 2
            ;;
        --target)
            TARGET_OS_ARCH="$2"
            shift 2
            ;;
        --all)
            BUILD_ALL=true
            shift
            ;;
        -h|--help|help)
            show_help
            exit 0
            ;;
        *)
            echo -e "${RED}Error: Unknown argument '$1'${RESET}" >&2
            echo "Run './build.sh --help' for usage information." >&2
            exit 1
            ;;
    esac
done

# Resolve version tags from Git
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
DATE=$(date -u '+%Y-%m-%dT%H:%M:%SZ')

LDFLAGS="-X github.com/jadmadi/thermal/internal/version.Version=${VERSION} \
-X github.com/jadmadi/thermal/internal/version.Commit=${COMMIT} \
-X github.com/jadmadi/thermal/internal/version.Date=${DATE}"

build_single() {
    local target_goos="${1:-$(go env GOOS)}"
    local target_goarch="${2:-$(go env GOARCH)}"
    local out_path="${3:-}"

    if [[ -z "$out_path" ]]; then
        if [[ "$target_goos" == "windows" ]]; then
            out_path="thermal.exe"
        else
            out_path="thermal"
        fi
    fi

    echo -e "${BOLD}Building thermal (${CYAN}${MODE}${RESET}${BOLD} mode) for ${YELLOW}${target_goos}/${target_goarch}${RESET}${BOLD}...${RESET}"

    local go_flags=()
    local env_vars=("GOOS=${target_goos}" "GOARCH=${target_goarch}")

    case "$MODE" in
        release|upx)
            env_vars+=("CGO_ENABLED=0")
            go_flags+=("-trimpath" "-ldflags=-s -w ${LDFLAGS}")
            ;;
        dev)
            go_flags+=("-ldflags=${LDFLAGS}")
            ;;
    esac

    # Run go build
    env "${env_vars[@]}" go build "${go_flags[@]}" -o "$out_path" ./cmd/thermal

    if [[ "$MODE" == "upx" ]]; then
        if ! command -v upx >/dev/null 2>&1; then
            echo -e "${RED}Warning: 'upx' command not found. Skipping LZMA compression.${RESET}" >&2
            echo -e "Install upx (e.g. 'sudo apt install upx-ucl') to enable ultra-compression." >&2
        else
            echo -e "Compressing ${CYAN}${out_path}${RESET} with UPX (--best --lzma)..."
            local upx_flags=("--best" "--lzma")
            if [[ "$target_goos" == "darwin" ]]; then
                upx_flags+=("--force-macos")
            fi
            upx "${upx_flags[@]}" "$out_path" >/dev/null
        fi
    fi

    local file_size
    if [[ "$target_goos" == "windows" ]] || [[ "$(uname -s)" == "Linux" ]]; then
        file_size=$(ls -lh "$out_path" | awk '{print $5}')
    else
        file_size=$(ls -lh "$out_path" | awk '{print $5}')
    fi

    echo -e "${GREEN}✔ Built successfully:${RESET} ${BOLD}${out_path}${RESET} (${CYAN}${file_size}${RESET})"
}

if [[ "$BUILD_ALL" == "true" ]]; then
    mkdir -p dist
    targets=(
        "linux/amd64"
        "linux/arm64"
        "darwin/amd64"
        "darwin/arm64"
        "windows/amd64"
    )
    for t in "${targets[@]}"; do
        IFS='/' read -r goos goarch <<< "$t"
        ext=""
        if [[ "$goos" == "windows" ]]; then ext=".exe"; fi
        out_file="dist/thermal_${VERSION}_${goos}_${goarch}${ext}"
        build_single "$goos" "$goarch" "$out_file"
    done
    echo -e "\n${GREEN}${BOLD}✔ All target builds completed inside dist/ directory.${RESET}"
elif [[ -n "$TARGET_OS_ARCH" ]]; then
    IFS='/' read -r goos goarch <<< "$TARGET_OS_ARCH"
    build_single "$goos" "$goarch" "$OUTPUT"
else
    build_single "$(go env GOOS)" "$(go env GOARCH)" "$OUTPUT"
fi
