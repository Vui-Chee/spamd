#!/usr/bin/env sh

detect_platform() {
  platform="$(uname -s | tr '[:upper:]' '[:lower:]')"
  case "${platform}" in
    linux) platform="linux" ;;
    darwin) platform="darwin" ;;
  esac
  printf '%s' "${platform}"
}

detect_arch() {
  arch="$(uname -m | tr '[:upper:]' '[:lower:]')"
  case "${arch}" in
    x86_64) arch="amd64" ;;
    armv*) arch="arm" ;;
    arm64) arch="aarch64" ;;
  esac
  printf '%s' "${arch}"
}

has_bindir() {
  if [ ! -d "$BIN_DIR" ]; then
    echo "Installation location $BIN_DIR does not appear to be a directory"
    echo "Make sure the location exists and is a directory, then try again."
    exit 1
  fi
}

can_write() {
  fake_file=$1/total_bs.txt
  if touch "${fake_file}" 2>/dev/null; then
    rm "${fake_file}"
    return 0
  fi
  return 1
}

download() {
  platform=$1
  arch=$2

  sudo=""
  if ! can_write $BIN_DIR; then
    sudo="sudo"
  fi

  ${sudo} curl --fail --location --output $BIN_DIR/spamd $BASE_URL/spamd_${platform}_${arch}
  ${sudo} chmod +x $BIN_DIR/spamd
}

do_install() {
  platform=$(detect_platform)
  arch=$(detect_arch)
  has_bindir
  download "${platform}" "${arch}"
}

BIN_DIR=/usr/local/bin
BASE_URL=https://github.com/Vui-Chee/spamd/releases/latest/download
do_install
