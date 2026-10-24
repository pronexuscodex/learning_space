#!/bin/sh
# Print the CHANGELOG.md section of a version, for its GitHub release.
#   tools/release_notes.sh v1.9.0 > notes.md
set -eu
version=${1#v}
dir=$(dirname "$0")/..
awk -v v="$version" '
  /^## / { if (found) exit; if (index($0, "[" v "]") || index($0, " " v)) { found = 1; next } }
  found { print }
' "$dir/CHANGELOG.md" | sed '/./,$!d'
