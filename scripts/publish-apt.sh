#!/usr/bin/env bash
# Publishes a .deb to the gh-pages apt repository.
# Usage: GPG_KEY_ID=<key-id> bash scripts/publish-apt.sh <path-to-deb> <version>
# Example: GPG_KEY_ID=ABC123 bash scripts/publish-apt.sh bin/flexcli_0.2.34_amd64.deb 0.2.34
set -euo pipefail

DEB_FILE="${1:?Usage: publish-apt.sh <path-to-deb> <version>}"
VERSION="${2:?Usage: publish-apt.sh <path-to-deb> <version>}"
GPG_KEY_ID="${GPG_KEY_ID:?GPG_KEY_ID env var is required}"

REPO_ROOT="$(git rev-parse --show-toplevel)"
WORKTREE_DIR="$(mktemp -d)"

cleanup() {
  git -C "$REPO_ROOT" worktree remove --force "$WORKTREE_DIR" 2>/dev/null || true
  rm -rf "$WORKTREE_DIR"
}
trap cleanup EXIT

echo "→ Setting up gh-pages worktree..."
if git -C "$REPO_ROOT" ls-remote --exit-code --heads origin gh-pages &>/dev/null; then
  git -C "$REPO_ROOT" fetch origin gh-pages:gh-pages 2>/dev/null || true
  git -C "$REPO_ROOT" worktree add "$WORKTREE_DIR" gh-pages
else
  echo "  gh-pages branch not found — creating orphan branch"
  git -C "$REPO_ROOT" worktree add --orphan -b gh-pages "$WORKTREE_DIR"
fi

# ── helpers ──────────────────────────────────────────────────────────────────

md5_hash()    { openssl dgst -md5    "$1" | awk '{print $NF}'; }
sha1_hash()   { openssl dgst -sha1   "$1" | awk '{print $NF}'; }
sha256_hash() { openssl dgst -sha256 "$1" | awk '{print $NF}'; }

file_size() {
  # macOS: stat -f%z   Linux: stat -c%s
  stat -f%z "$1" 2>/dev/null || stat -c%s "$1"
}

# ensure Jekyll doesn't interfere with serving apt repo files
touch "$WORKTREE_DIR/.nojekyll"

# ── copy .deb into pool ───────────────────────────────────────────────────────

POOL_DIR="$WORKTREE_DIR/pool/main/f/flexcli"
mkdir -p "$POOL_DIR"
DEB_FILENAME="$(basename "$DEB_FILE")"
cp "$DEB_FILE" "$POOL_DIR/$DEB_FILENAME"
echo "→ Copied $DEB_FILENAME to pool"

# ── generate Packages ─────────────────────────────────────────────────────────

DIST_DIR="$WORKTREE_DIR/dists/stable/main/binary-amd64"
mkdir -p "$DIST_DIR"

DEB_ABS="$POOL_DIR/$DEB_FILENAME"
DEB_SIZE="$(file_size "$DEB_ABS")"
DEB_MD5="$(md5_hash "$DEB_ABS")"
DEB_SHA1="$(sha1_hash "$DEB_ABS")"
DEB_SHA256="$(sha256_hash "$DEB_ABS")"

# Installed-Size is the unpacked binary size in KiB
INSTALLED_KB="$(du -sk "$REPO_ROOT/bin/flexcli-linux" 2>/dev/null | cut -f1)" || INSTALLED_KB=0

cat > "$DIST_DIR/Packages" <<EOF
Package: flexcli
Version: ${VERSION}
Architecture: amd64
Maintainer: Lukas Leidinger <lukas.leidinger@willhaben.at>
Installed-Size: ${INSTALLED_KB}
Section: utils
Priority: optional
Homepage: https://github.com/f1dot4/homebrew-flexcli
Description: FlexCLI - Management CLI for FlexCoach AI fitness platform
 Allows users and developers to interact with the backend API to manage
 profiles, training plans, goals, and system status.
Filename: pool/main/f/flexcli/${DEB_FILENAME}
Size: ${DEB_SIZE}
MD5sum: ${DEB_MD5}
SHA1: ${DEB_SHA1}
SHA256: ${DEB_SHA256}
EOF

gzip -c "$DIST_DIR/Packages" > "$DIST_DIR/Packages.gz"
echo "→ Generated Packages / Packages.gz"

# ── generate Release ──────────────────────────────────────────────────────────

RELEASE="$WORKTREE_DIR/dists/stable/Release"

PKG_SIZE="$(file_size "$DIST_DIR/Packages")"
PKG_GZ_SIZE="$(file_size "$DIST_DIR/Packages.gz")"
PKG_MD5="$(md5_hash "$DIST_DIR/Packages")"
PKG_GZ_MD5="$(md5_hash "$DIST_DIR/Packages.gz")"
PKG_SHA256="$(sha256_hash "$DIST_DIR/Packages")"
PKG_GZ_SHA256="$(sha256_hash "$DIST_DIR/Packages.gz")"

cat > "$RELEASE" <<EOF
Origin: flexcli
Label: flexcli
Suite: stable
Codename: stable
Architectures: amd64
Components: main
Description: FlexCLI apt repository
Date: $(date -u '+%a, %d %b %Y %H:%M:%S UTC')
MD5Sum:
 ${PKG_MD5} ${PKG_SIZE} main/binary-amd64/Packages
 ${PKG_GZ_MD5} ${PKG_GZ_SIZE} main/binary-amd64/Packages.gz
SHA256:
 ${PKG_SHA256} ${PKG_SIZE} main/binary-amd64/Packages
 ${PKG_GZ_SHA256} ${PKG_GZ_SIZE} main/binary-amd64/Packages.gz
EOF

echo "→ Generated Release"

# ── sign ──────────────────────────────────────────────────────────────────────

gpg --batch --yes --default-key "$GPG_KEY_ID" \
  --clearsign -o "$WORKTREE_DIR/dists/stable/InRelease" "$RELEASE"

gpg --batch --yes --default-key "$GPG_KEY_ID" \
  -abs -o "$WORKTREE_DIR/dists/stable/Release.gpg" "$RELEASE"

echo "→ Signed InRelease / Release.gpg"

# ── export public key (idempotent) ────────────────────────────────────────────

gpg --armor --export "$GPG_KEY_ID" > "$WORKTREE_DIR/flexcli.gpg"
echo "→ Exported public key to flexcli.gpg"

# ── commit and push ───────────────────────────────────────────────────────────

cd "$WORKTREE_DIR"
git add -A
git commit -m "apt: release v${VERSION}"
git push origin gh-pages
echo "→ Pushed gh-pages — apt repo updated for v${VERSION}"
