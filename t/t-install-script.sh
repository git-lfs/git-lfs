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

  # We create an executable with the same name as the Git LFS client
  # in a directory that we place first in the system path when running
  # our installation script.  If the script ignores its --skip-config
  # flag and invokes Git LFS, our executable will leave a sentinel file
  # and log messages we can detect, and should also cause the script
  # to exit with an error status.
  cp "$BINPATH/lfstest-executecheck" "$binpath/git-lfs"

  PATH="$fake_bin:$PATH" PREFIX="$prefix" \
    "$package/install.sh" --skip-config 2>&1 | tee install.log
  [ 0 -eq "${PIPESTATUS[0]}" ]

  [ ! -f execute ]

  [ 0 -eq "$(grep -c "execute: .*git-lfs install" install.log)" ]

  [ -x "$prefix/bin/git-lfs" ]
)
end_test
