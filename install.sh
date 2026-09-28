#!/bin/sh
# Install the Systems & AI Academy on macOS, Linux or FreeBSD.
#
#   curl -fsSL https://raw.githubusercontent.com/pronexuscodex/learning_space/main/install.sh | sh
#
# It downloads the latest release (or ACADEMY_VERSION=vX.Y.Z), checks it
# against the release's SHA256SUMS, and installs:
#
#   ~/.local/bin/academy                        the program (ACADEMY_INSTALL_DIR to change)
#   an app-menu entry with the icon             Linux: ~/.local/share/applications
#                                               macOS: ~/Applications/Academy.app
#
# It adds ~/.local/bin to your PATH in your shell's startup file when it is
# not there yet (ACADEMY_NO_MODIFY_PATH=1 to skip). Nothing needs sudo.
# Your progress is kept in your user data folder and your PDFs in
# Documents/Academy Library, so reinstalling or upgrading never touches them.
#
# Upgrade: run the same command again.
# Remove:  curl -fsSL .../install.sh | sh -s -- --uninstall
set -eu

REPO=pronexuscodex/learning_space
APP_NAME="Systems & AI Academy"
BIN_DIR=${ACADEMY_INSTALL_DIR:-$HOME/.local/bin}
DATA_HOME=${XDG_DATA_HOME:-$HOME/.local/share}
DESKTOP_FILE=$DATA_HOME/applications/academy.desktop
ICON_FILE=$DATA_HOME/icons/hicolor/256x256/apps/academy.png
MAC_APP=$HOME/Applications/Academy.app

say() { printf '%s\n' "$*"; }
fail() { printf 'academy install: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || fail "needs '$1', which was not found"; }

fetch() { # fetch URL FILE
	if command -v curl >/dev/null 2>&1; then
		curl -fsSL --proto '=https' --tlsv1.2 --retry 3 -o "$2" "$1"
	elif command -v wget >/dev/null 2>&1; then
		wget -q --https-only -O "$2" "$1"
	else
		fail "needs curl or wget"
	fi
}

uninstall() {
	where=
	if [ -x "$BIN_DIR/academy" ]; then
		where=$("$BIN_DIR/academy" -where 2>/dev/null | grep -v '^Program:' || true)
	fi
	rm -f "$BIN_DIR/academy" "$DESKTOP_FILE" "$ICON_FILE"
	rm -rf "$MAC_APP"
	if command -v update-desktop-database >/dev/null 2>&1; then
		update-desktop-database -q "$DATA_HOME/applications" 2>/dev/null || true
	fi
	say "Removed $APP_NAME."
	say "Your progress and PDFs were kept; delete these folders yourself if you no longer want them:"
	if [ -n "$where" ]; then
		say "$where"
	else
		say "  the 'academy' folder in your user data folder, and Documents/Academy Library"
	fi
}

case "${1:-}" in
--uninstall) uninstall; exit 0 ;;
"") ;;
*) fail "unknown option '$1' (the only option is --uninstall)" ;;
esac

case $(uname -s) in
Linux) os=linux ;;
Darwin) os=darwin ;;
FreeBSD) os=freebsd ;;
*) fail "unsupported system $(uname -s); on Windows use install.ps1" ;;
esac
case $(uname -m) in
x86_64 | amd64) arch=amd64 ;;
aarch64 | arm64) arch=arm64 ;;
*) fail "unsupported processor $(uname -m)" ;;
esac
# A Mac with Apple silicon can run this script under Rosetta; install the native build.
if [ "$os" = darwin ] && [ "$arch" = amd64 ] && [ "$(sysctl -n sysctl.proc_translated 2>/dev/null || echo 0)" = 1 ]; then
	arch=arm64
fi
if [ "$os" = freebsd ] && [ "$arch" != amd64 ]; then
	fail "FreeBSD builds are amd64 only"
fi
need tar
need gzip

version=${ACADEMY_VERSION:-}
if [ -z "$version" ]; then
	# The latest-release page redirects to its tag; no API token or rate limit.
	if command -v curl >/dev/null 2>&1; then
		url=$(curl -fsSLI --proto '=https' -o /dev/null -w '%{url_effective}' "https://github.com/$REPO/releases/latest")
	else
		url=$(wget -q --https-only --max-redirect=5 -S --spider "https://github.com/$REPO/releases/latest" 2>&1 | sed -n 's/^ *[Ll]ocation: *//p' | tail -n 1)
	fi
	version=${url##*/}
	version=$(printf '%s' "$version" | tr -d '\r')
fi
case $version in
v[0-9]*.[0-9]*.[0-9]*) ;;
*) fail "'$version' is not a release version; set ACADEMY_VERSION=vX.Y.Z, or leave it unset for the latest" ;;
esac

name=academy-$version-$os-$arch
base=https://github.com/$REPO/releases/download/$version
tmp=$(mktemp -d)
trap 'rm -rf "$tmp"' EXIT INT TERM

say "Downloading $APP_NAME $version for $os/$arch..."
fetch "$base/$name.tar.gz" "$tmp/$name.tar.gz"
fetch "$base/SHA256SUMS" "$tmp/SHA256SUMS"

want=$(awk -v f="$name.tar.gz" '$2 == f { print $1 }' "$tmp/SHA256SUMS")
[ -n "$want" ] || fail "$name.tar.gz is not listed in SHA256SUMS"
if command -v sha256sum >/dev/null 2>&1; then
	got=$(sha256sum "$tmp/$name.tar.gz" | awk '{ print $1 }')
elif command -v shasum >/dev/null 2>&1; then
	got=$(shasum -a 256 "$tmp/$name.tar.gz" | awk '{ print $1 }')
else
	got=$(sha256 -q "$tmp/$name.tar.gz") # FreeBSD
fi
[ "$got" = "$want" ] || fail "checksum mismatch for $name.tar.gz (expected $want, got $got); nothing was installed"
say "Checksum verified."

tar -xzf "$tmp/$name.tar.gz" -C "$tmp"
[ -f "$tmp/$name/academy" ] || fail "the archive does not contain academy"
mkdir -p "$BIN_DIR"
# Replace atomically, so a running copy keeps working during an upgrade.
cp "$tmp/$name/academy" "$BIN_DIR/.academy.new"
chmod 755 "$BIN_DIR/.academy.new"
mv -f "$BIN_DIR/.academy.new" "$BIN_DIR/academy"
say "Installed $BIN_DIR/academy"

# The icon comes from the same tagged source as the release. It is only
# decoration, so a failure here does not stop the install.
icon=$tmp/icon.png
fetch "https://raw.githubusercontent.com/$REPO/$version/academy/assets/png/icon-256.png" "$icon" 2>/dev/null || icon=

if [ "$os" = darwin ]; then
	# A small app bundle that opens the academy in Terminal, so it can be
	# found in Launchpad and Spotlight and kept in the Dock.
	rm -rf "$MAC_APP"
	mkdir -p "$MAC_APP/Contents/MacOS" "$MAC_APP/Contents/Resources"
	cat >"$MAC_APP/Contents/MacOS/Academy" <<EOF
#!/bin/sh
exec open -a Terminal "$BIN_DIR/academy"
EOF
	chmod 755 "$MAC_APP/Contents/MacOS/Academy"
	icon_key=
	if [ -n "$icon" ] && command -v sips >/dev/null 2>&1 && command -v iconutil >/dev/null 2>&1; then
		set_dir=$tmp/Academy.iconset
		mkdir -p "$set_dir"
		for s in 16 32 128 256; do
			sips -z $s $s "$icon" --out "$set_dir/icon_${s}x${s}.png" >/dev/null 2>&1 || true
			d=$((s * 2))
			if [ $d -le 256 ]; then
				sips -z $d $d "$icon" --out "$set_dir/icon_${s}x${s}@2x.png" >/dev/null 2>&1 || true
			fi
		done
		if iconutil -c icns "$set_dir" -o "$MAC_APP/Contents/Resources/Academy.icns" 2>/dev/null; then
			icon_key="<key>CFBundleIconFile</key><string>Academy</string>"
		fi
	fi
	cat >"$MAC_APP/Contents/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>CFBundleName</key><string>Academy</string>
	<key>CFBundleDisplayName</key><string>$APP_NAME</string>
	<key>CFBundleIdentifier</key><string>io.github.pronexuscodex.academy</string>
	<key>CFBundleExecutable</key><string>Academy</string>
	<key>CFBundlePackageType</key><string>APPL</string>
	<key>CFBundleShortVersionString</key><string>${version#v}</string>
	$icon_key
</dict>
</plist>
EOF
	say "Added $MAC_APP (Launchpad and Spotlight: \"Academy\")"
else
	mkdir -p "$(dirname "$DESKTOP_FILE")"
	icon_line="Icon=utilities-terminal"
	if [ -n "$icon" ]; then
		mkdir -p "$(dirname "$ICON_FILE")"
		cp "$icon" "$ICON_FILE"
		icon_line="Icon=$ICON_FILE"
	fi
	cat >"$DESKTOP_FILE" <<EOF
[Desktop Entry]
Type=Application
Name=$APP_NAME
GenericName=Study companion
Comment=Computer science, AI and security, stage by stage
Exec="$BIN_DIR/academy"
Terminal=true
$icon_line
Categories=Education;ComputerScience;
Keywords=learn;study;computer science;C;AI;security;
EOF
	if command -v update-desktop-database >/dev/null 2>&1; then
		update-desktop-database -q "$(dirname "$DESKTOP_FILE")" 2>/dev/null || true
	fi
	say "Added \"$APP_NAME\" to your applications menu"
fi

on_path=no
case ":$PATH:" in *":$BIN_DIR:"*) on_path=yes ;; esac
if [ "$on_path" = no ] && [ "${ACADEMY_NO_MODIFY_PATH:-}" != 1 ]; then
	line="export PATH=\"$BIN_DIR:\$PATH\""
	case ${SHELL:-} in
	*/zsh) rc=$HOME/.zshrc ;;
	*/bash) if [ "$os" = darwin ]; then rc=$HOME/.bash_profile; else rc=$HOME/.bashrc; fi ;;
	*/fish) rc=$HOME/.config/fish/conf.d/academy.fish; line="fish_add_path \"$BIN_DIR\"" ;;
	*) rc=$HOME/.profile ;;
	esac
	mkdir -p "$(dirname "$rc")"
	if ! grep -qsF "$line" "$rc"; then
		printf '\n# Added by the Systems & AI Academy installer\n%s\n' "$line" >>"$rc"
	fi
	say "Added $BIN_DIR to your PATH in $rc (open a new terminal for it to apply)"
fi

say ""
say "Done. Start it from your applications menu, or type:  academy"
if [ "$on_path" = no ]; then
	say "(in this terminal, until you open a new one:  $BIN_DIR/academy)"
fi
say "See where your progress and PDFs are kept:  academy -where"
say "Upgrade later by running the same install command again."
