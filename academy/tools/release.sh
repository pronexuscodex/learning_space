#!/bin/sh
# Build release archives for every supported platform.
#
#   tools/release.sh v1.0.0        # from the academy/ directory
#
# Produces dist/academy-<version>-<os>-<arch>.tar.gz (.zip for Windows),
# each holding the executable, README.md, CHANGELOG.md and LICENSE (when
# present), plus dist/SHA256SUMS. Needs only Go, tar, zip and sha256sum
# (or shasum on macOS). Builds are static (CGO_ENABLED=0), reproducible
# (-trimpath) and stripped (-s -w).
set -eu

VERSION=${1:?usage: tools/release.sh vX.Y.Z}
case "$VERSION" in
v[0-9]*.[0-9]*.[0-9]*) ;;
*) echo "version must look like v1.2.3, got $VERSION" >&2; exit 1 ;;
esac

cd "$(dirname "$0")/.."
TARGETS="linux/amd64 linux/arm64 darwin/amd64 darwin/arm64 windows/amd64 windows/arm64 freebsd/amd64"
OUT=dist
rm -rf "$OUT"
mkdir -p "$OUT"

for target in $TARGETS; do
	os=${target%/*}
	arch=${target#*/}
	name="academy-$VERSION-$os-$arch"
	dir="$OUT/$name"
	mkdir -p "$dir"
	exe=academy
	[ "$os" = windows ] && exe=academy.exe
	echo "building $name"
	CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build -trimpath \
		-ldflags "-s -w -X main.version=$VERSION" -o "$dir/$exe" .
	for f in README.md CHANGELOG.md LICENSE; do
		[ -f "$f" ] && cp "$f" "$dir/"
	done
	if [ "$os" = windows ]; then
		(cd "$OUT" && zip -qr "$name.zip" "$name")
	else
		tar -C "$OUT" -czf "$OUT/$name.tar.gz" "$name"
	fi
	rm -rf "$dir"
done

cd "$OUT"
if command -v sha256sum >/dev/null 2>&1; then
	sha256sum academy-* > SHA256SUMS
else
	shasum -a 256 academy-* > SHA256SUMS
fi
echo
cat SHA256SUMS
