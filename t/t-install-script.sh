#!/usr/bin/env bash

. "$(dirname "$0")/testlib.sh"

begin_test "binary installer can skip Git configuration"
(
  set -e

  # The installer is a Unix shell script and is not used by Windows packages.
  [ "$IS_WINDOWS" -eq 1 ] && exit 0

  package="$TRASHDIR/install-script-package"
  prefix="$TRASHDIR/install-script-prefix"
  fake_bin="$TRASHDIR/install-script-bin"
  mkdir -p "$package" "$prefix" "$fake_bin"
  cp "$ROOTDIR/script/install.sh" "$package/install.sh"
  printf '#!/usr/bin/env bash\n' > "$package/git-lfs"
  chmod +x "$package/install.sh" "$package/git-lfs"

  cat > "$fake_bin/git" <<EOF
#!/usr/bin/env bash
printf '%s\\n' "\$*" >> "$TRASHDIR/git-invocations"
exit 99
EOF
  chmod +x "$fake_bin/git"

  PATH="$fake_bin:$PATH" PREFIX="$prefix" "$package/install.sh" --bin

  [ -x "$prefix/bin/git-lfs" ]
  [ ! -e "$TRASHDIR/git-invocations" ]
)
end_test
