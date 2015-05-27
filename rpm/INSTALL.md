# Building RPMs #

All of the code to build the RPM is stored in a SPECS/git-lfs.spec file. The source 
code tarball needs to be put in a SOURCES directory. A BUILD and BUILDROOT directory 
is used during the build process, and the final RPM ends up in the RPMS directory, 
and a source-rpm in SRPMS

In order to expedite installing all dependencies (mainly ruby-rconn and golang) and 
download any files (outside of yum) a build_rpms.bsh script is included. This is the 
**RECOMMENDED** way to build the rpms. It will install all yum packages in order to
build the rpm. This can be especially difficult in CentOS 5 and 6, but it will build
and install a suitable golang/ruby so that git-lfs can be built.

Simple run:

```
./build_rpms.bsh
```

Practice is to run rpmbuild as non-root user. This prevents inadvertently installing
files in the operating system. The intent was to run build_rpms.bsh as a non-root user
with sudo privileges. If you have a different command for sudo, or do not have sudo
installed (which is possible, but unlikely), you can set the SUDO environment variable
to nothing or another command and you can run as root if that is your style. Example:

```
SUDO=echo ./build_rpms.bsh
  or
(as root) SUDO= ./build_rpms.bsh
```

(The echo example will let you know what yum commands you need to run to make the build
work. Not ideal, but 95% of people will just run ```./build_rpms.bsh``` and have it work)

When all is down, install (or distribute) RPMS/git-lfs.rpm 

### Alternative build ###

If you want to use your own ruby/golang without using build_rpms.bsh, just make sure
rconn and go are in the path, and run

```
rpmbuild --define "_topdir `pwd`" -bb SPECS/git-lfs.spec --nodeps

#(and optionally)
rpmbuild --define "_topdir `pwd`" -bs SPECS/git-lfs.spec --nodeps
```

### Releasing ###

The only thing that needs to be updated with a new version is the version number in 
git-lfs.spec needs to be updated. It will download:

https://github.com/github/git-lfs/archive/v%{version}.tar.gz 

This way when a new version is archived, it will always download get downloaded. Of
course this is a bit of a chicken/egg issue with the spec being stored in the repo... 
detail details... If you always want the master branch, I guess you can change the 
version to master, but I'm not not sure why you would bother making an rpm for that.

### Troubleshooting ###

**Q**) "error: Bad owner/group" when building SRPM (rpmbuild -bs command)

**A**) For some... STUPID reason, git-lfs.spec has to be OWNED by a valid used. Just chown 
git-lfs.spec to a valid user AND group. root will do

### TODO ###

- Add a "use current checkout" mode to create a tar.gz out of the current checkout instead
of downloading the archive, for release testing BEFORE release is released. 
