#!/bin/sh -e

KEYCHAIN_ID="lfs.keychain-db"
KEYCHAIN_PATH="$HOME/Library/Keychains/${KEYCHAIN_ID}"

check_var_exists() {
  var="$1"
  val="$(eval echo "\$$var")"

  if [ -z "$val" ]; then
    echo "The $var variable is missing." >&2
    exit 1
  fi
}

check_keychain_exists() {
  if [ ! -f "$KEYCHAIN_PATH" ]; then
    echo "The '$KEYCHAIN_PATH' file is missing." >&2
    exit 1
  fi
}
