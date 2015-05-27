## Building on Linux

There are build scripts for recent versions of CentOS- and Debian-flavored
Linuxes in `../scripts/{centos,debian}-build`. Both install all prerequisites,
then build the client and the man pages in Docker containers for CentOS 7,
Debian 8, and Ubuntu 14.04.

On CentOS 6, the client builds, but not the man pages, because of problems
getting the right version of Ruby.

Earlier versions of CentOS and Debian/Ubuntu have trouble building go, so they
are non-starters.

## Building a deb

A debian package can be built by running `dpkg-buildpackage -us -uc` from the
root of the repo.  It is currently confirmed to work on Debian jessie and
wheezy.  On wheezy it requires `wheezy-backports` versions of `dh-golang`,
`git`, and `golang`.
