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

## Building an rpm

An rpm package can be built by running ```./rpm/build_rpms.bsh```. All 
dependencies will be downloaded, compiled, and installed for you, provided
you have sudo/root permissions. The resulting ./rpm/RPMS/x86_64/git-lfs*.rpm
Can be installed using ```yum install``` or distributed. 

-CentOS 7 - build_rpms.bsh will take care of everything. You only need the
git-lfs rpm
-CentOS 6 - build_rpms.bsh will take care of everything. You will need to
distribute both the git-lfs rpms and the git rpms, as CentOS 6 does not
have a current enough version available
-CentOS 5 - build_rpms.bsh will take care of everything. You only need the
git-lfs rpm. When distributing to CentOS 5, they will need git from the epel
repo
```
yum install epel-release
yum install git
```

See ./rpm/INSTALL.md for more detail