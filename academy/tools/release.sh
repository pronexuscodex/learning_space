#!/bin/sh
# Build release archives for every supported platform.
#
#   tools/release.sh v1.0.0        # from the academy/ directory
#
# Produces dist/academy-<version>-<os>-<arch>.tar.gz (.zip for Windows),
# each holding the executable, README.md, CHANGELOG.md and LICENSE (when
# present), plus dist/SHA256SUMS. Needs only Go, tar, zip and sha256sum
# (or shasum on macOS). Builds are static (CGO_ENABLED=0), reproducible
# (-trimpath) and stripped (-s -w), and so are the archives: every file
# gets the commit's timestamp (or $SOURCE_DATE_EPOCH), a fixed owner and a
# fixed order, so building the same commit twice gives identical checksums.
set -eu

VERSION=${1:?usage: tools/release.sh vX.Y.Z}
case "$VERSION" in
v[0-9]*.[0-9]*.[0-9]*) ;;
*) echo "version must look like v1.2.3, got $VERSION" >&2; exit 1 ;;
esac

cd "$(dirname "$0")/.."
EPOCH=${SOURCE_DATE_EPOCH:-$(git log -1 --format=%ct 2>/dev/null || date +%s)}
STAMP=$(date -u -d "@$EPOCH" +%Y%m%d%H%M.%S 2>/dev/null || date -u -r "$EPOCH" +%Y%m%d%H%M.%S)
export TZ=UTC # touch -t and zip store local times
if tar --version 2>/dev/null | grep -q GNU; then
	TAR_FLAGS="--format=ustar --owner=0 --group=0 --numeric-owner"
else
	TAR_FLAGS="--format=ustar --uid 0 --gid 0 --numeric-owner" # bsdtar (macOS)
fi
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
	# The license may live at the repository root instead.
	if [ ! -f LICENSE ] && [ -f ../LICENSE ]; then
		cp ../LICENSE "$dir/"
	fi
	chmod 755 "$dir" "$dir/$exe"
	chmod 644 "$dir"/*.md "$dir/LICENSE" 2>/dev/null || true
	touch -t "$STAMP" "$dir" "$dir"/*
	files=$(cd "$OUT" && ls -d "$name" "$name"/* | LC_ALL=C sort)
	if [ "$os" = windows ]; then
		# -X: no extra attributes; -D: no directory entries.
		(cd "$OUT" && zip -qXD "$name.zip" $files)
	else
		# -n: gzip without the file name and time.
		(cd "$OUT" && tar $TAR_FLAGS --no-recursion -cf - $files | gzip -n9 > "$name.tar.gz")
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
