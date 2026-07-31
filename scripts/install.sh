#!/usr/bin/env bash
# momo — AI coding agent CLI installer
# Usage: curl -fsSL https://raw.githubusercontent.com/Bhagattji/momo/main/scripts/install.sh | bash

set -euo pipefail

REPO="Bhagattji/momo"
INSTALL_DIR="${HOME}/.momo/bin"
BINARY="${INSTALL_DIR}/momo"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

echo ""
echo -e "${CYAN}  __  __                  ${NC}"
echo -e "${CYAN} |  \\/  | ___  _ __ ___   ${NC}"
echo -e "${CYAN} | |\\/| |/ _ \\| '_ \` _ \\ ${NC}"
echo -e "${CYAN} | |  | | (_) | | | | | |  ${NC}"
echo -e "${CYAN} |_|  |_|\\___/|_| |_| |_| ${NC}"
echo ""

detect_platform() {
    local os arch
    os=$(uname -s | tr '[:upper:]' '[:lower:]')
    arch=$(uname -m)
    case "$arch" in
        x86_64|amd64) arch="amd64" ;;
        aarch64|arm64) arch="arm64" ;;
        *) echo -e "${RED}Unsupported architecture: $arch${NC}" >&2; exit 1 ;;
    esac
    echo "${os}_${arch}"
}

get_latest_tag() {
    curl -sLf --retry 3 "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name":' | sed -E 's/.*"([^"]+)".*/\1/'
}

download_and_install() {
    local tag platform asset_url
    platform=$(detect_platform)
    tag=$(get_latest_tag)

    if [[ "$platform" == darwin* ]]; then
        asset_url="https://github.com/${REPO}/releases/download/${tag}/momo_${tag#v}_darwin_${platform#darwin_}.tar.gz"
    elif [[ "$platform" == linux* ]]; then
        asset_url="https://github.com/${REPO}/releases/download/${tag}/momo_${tag#v}_linux_${platform#linux_}.tar.gz"
    else
        echo -e "${RED}Unsupported platform: $platform${NC}" >&2
        exit 1
    fi

    echo -e "${YELLOW}  Downloading momo ${tag} (${platform})...${NC}"
    mkdir -p "${INSTALL_DIR}"

    local tmpdir
    tmpdir=$(mktemp -d)
    trap 'rm -rf "$tmpdir"' EXIT

    if command -v curl &>/dev/null; then
        curl -fsSL --retry 3 "$asset_url" -o "${tmpdir}/momo.tar.gz"
    elif command -v wget &>/dev/null; then
        wget -q --tries=3 "$asset_url" -O "${tmpdir}/momo.tar.gz"
    else
        echo -e "${RED}curl or wget required${NC}" >&2
        exit 1
    fi

    tar -xzf "${tmpdir}/momo.tar.gz" -C "${tmpdir}"
    mv "${tmpdir}/momo" "${BINARY}"
    chmod +x "${BINARY}"
}

add_to_path() {
    local shell_profile
    case "$SHELL" in
        */zsh)  shell_profile="${HOME}/.zshrc" ;;
        */bash) shell_profile="${HOME}/.bashrc" ;;
        */fish) echo "  Set PATH manually: fish_add_path ${INSTALL_DIR}"; return 0 ;;
        *) shell_profile="${HOME}/.profile" ;;
    esac

    if ! grep -q "${INSTALL_DIR}" "${shell_profile}" 2>/dev/null; then
        echo "" >> "${shell_profile}"
        echo "# momo CLI" >> "${shell_profile}"
        echo "export PATH=\"\$PATH:${INSTALL_DIR}\"" >> "${shell_profile}"
        echo -e "${GREEN}  Added ${INSTALL_DIR} to PATH in ${shell_profile}${NC}"
    fi
}

verify() {
    local ver
    ver=$("${BINARY}" --version 2>&1)
    echo ""
    echo -e "${GREEN}  momo ${ver} installed successfully!${NC}"
    echo -e "  Binary: ${BINARY}"
    echo ""
    echo -e "${CYAN}  Open a NEW terminal and type:${NC}"
    echo -e "    momo"
    echo ""
}

download_and_install
add_to_path
verify