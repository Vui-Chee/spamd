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
    amd64) arch="x86_64" ;;
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

download() {
  platform=$1
  arch=$2
	echo $BASE_URL/spamd_${platform}_${arch}
  curl --fail --location --output $BIN_DIR/spamd $BASE_URL/spamd_${platform}_${arch}
	chmod +x $BIN_DIR/spamd
}

do_install() {
	platform=$(detect_platform)
	arch=$(detect_arch)
	has_bindir
	download "${platform}" "${arch}"
}

BASE_URL=https://storage.googleapis.com/spamd-releases/install.sh
do_install
