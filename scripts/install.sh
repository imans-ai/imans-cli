#!/usr/bin/env bash
#
# Imans CLI installer.
#
#   curl -fsSL https://imans.ai/install | bash
#
# Downloads the correct release archive for your OS/arch from GitHub Releases,
# verifies its SHA-256 checksum, and installs the `imans` binary.
#
# Environment overrides:
#   IMANS_VERSION       install a specific version (e.g. v0.1.0); default: latest
#   IMANS_INSTALL_DIR   install location; default: /usr/local/bin if writable,
#                       otherwise $HOME/.local/bin
#   IMANS_RELEASE_BASE  releases base URL (for a mirror); default GitHub Releases
#
set -euo pipefail

REPO="imans-ai/imans-cli"
BINARY="imans"

err() { printf 'error: %s\n' "$1" >&2; exit 1; }
info() { printf '%s\n' "$1" >&2; }

need() { command -v "$1" >/dev/null 2>&1 || err "required command not found: $1"; }

main() {
	need curl
	need tar

	# --- platform detection ---
	local os arch uname_s uname_m
	uname_s="$(uname -s)"
	uname_m="$(uname -m)"
	case "$uname_s" in
		Linux)  os="linux" ;;
		Darwin) os="darwin" ;;
		*) err "unsupported OS: $uname_s (Windows users: use Scoop — see ${REPO} README)" ;;
	esac
	case "$uname_m" in
		x86_64|amd64) arch="amd64" ;;
		aarch64|arm64) arch="arm64" ;;
		*) err "unsupported architecture: $uname_m" ;;
	esac

	# --- sha256 tool ---
	local sha_cmd
	if command -v sha256sum >/dev/null 2>&1; then
		sha_cmd="sha256sum"
	elif command -v shasum >/dev/null 2>&1; then
		sha_cmd="shasum -a 256"
	else
		err "need sha256sum or shasum to verify the download"
	fi

	# --- resolve version ---
	local version="${IMANS_VERSION:-}"
	if [ -z "$version" ]; then
		info "Resolving latest release..."
		local latest_json
		latest_json="$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest")" \
			|| err "could not reach the GitHub API (network or rate limit?) — set IMANS_VERSION to install a specific version"
		version="$(printf '%s' "$latest_json" | grep '"tag_name":' | head -n1 | sed -E 's/.*"tag_name": *"([^"]+)".*/\1/')"
		[ -n "$version" ] || err "could not parse the latest release tag — set IMANS_VERSION to install a specific version"
	fi
	local num_version="${version#v}"

	local archive="${BINARY}_${num_version}_${os}_${arch}.tar.gz"
	# IMANS_RELEASE_BASE allows pointing at a mirror (or a local server for
	# testing); defaults to GitHub Releases.
	local releases_base="${IMANS_RELEASE_BASE:-https://github.com/${REPO}/releases}"
	local base="${releases_base}/download/${version}"

	# --- download + verify in a temp dir ---
	# tmp is intentionally global (not `local`) so the EXIT trap can still see it
	# after main() returns; ${tmp:-} keeps the trap safe under `set -u`.
	tmp="$(mktemp -d)"
	trap 'rm -rf "${tmp:-}"' EXIT

	info "Downloading ${archive} (${version})..."
	curl -fSL -o "$tmp/$archive" "$base/$archive" \
		|| err "download failed: $base/$archive"
	curl -fsSL -o "$tmp/checksums.txt" "$base/checksums.txt" \
		|| err "could not download checksums.txt"

	info "Verifying checksum..."
	local expected actual
	expected="$(grep " ${archive}\$" "$tmp/checksums.txt" | awk '{print $1}')"
	[ -n "$expected" ] || err "no checksum listed for ${archive}"
	actual="$(cd "$tmp" && $sha_cmd "$archive" | awk '{print $1}')"
	[ "$expected" = "$actual" ] || err "checksum mismatch for ${archive} (expected $expected, got $actual)"

	tar -xzf "$tmp/$archive" -C "$tmp" "$BINARY" || err "failed to extract $BINARY"

	# --- choose install dir ---
	local dir="${IMANS_INSTALL_DIR:-}"
	if [ -z "$dir" ]; then
		if [ -w /usr/local/bin ] || { mkdir -p /usr/local/bin 2>/dev/null && [ -w /usr/local/bin ]; }; then
			dir="/usr/local/bin"
		else
			dir="$HOME/.local/bin"
		fi
	fi
	mkdir -p "$dir" || err "cannot create install directory: $dir"

	# --- install ---
	local sudo=""
	if [ ! -w "$dir" ]; then
		if command -v sudo >/dev/null 2>&1; then
			info "Elevated permissions needed to write to $dir"
			sudo="sudo"
		else
			err "no write permission for $dir (set IMANS_INSTALL_DIR to a writable path)"
		fi
	fi
	$sudo install -m 0755 "$tmp/$BINARY" "$dir/$BINARY" || err "failed to install to $dir"

	info ""
	info "✓ Installed $BINARY $version to $dir/$BINARY"
	case ":$PATH:" in
		*":$dir:"*) info "Run: $BINARY login" ;;
		*) info "Add it to your PATH, then run '$BINARY login':"
		   info "  export PATH=\"$dir:\$PATH\"" ;;
	esac
}

main "$@"
