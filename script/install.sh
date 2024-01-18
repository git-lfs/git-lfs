#!/usr/bin/env bash
set -eu

prefix="/usr/local"

if [ "${PREFIX:-}" != "" ] ; then
  prefix=${PREFIX:-}
elif [ "${BOXEN_HOME:-}" != "" ] ; then
  prefix=${BOXEN_HOME:-}
fi

while [[ $# -gt 0 ]]; do
  case "$1" in
    --local)
      prefix="$HOME/.local"
      shift
      ;;
    *)
      echo "Unknown option: $1"
      exit 1
      ;;
  esac
done

# Check if the user has permission to install in the specified prefix
if [ ! -w "$prefix" ]; then
  echo "Error: Insufficient permissions to install in $prefix. Try running with sudo or choose a different prefix.">&2
  exit 1
fi

mkdir -p "$prefix/bin"
rm -rf "$prefix/bin/git-lfs*"

pushd "$( dirname "${BASH_SOURCE[0]}" )" > /dev/null
  for g in git*; do
    install "$g" "$prefix/bin/$g"
  done
popd > /dev/null

PATH+=:"$prefix/bin"
git lfs install
