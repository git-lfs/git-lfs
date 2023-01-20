# Installing on Linux using packagecloud

[packagecloud](https://packagecloud.io) hosts [`git-lfs` packages](https://packagecloud.io/github/git-lfs) for popular Linux distributions with apt/deb and yum/rpm based package-managers.  Installing from packagecloud is reasonably straightforward and involves two steps:

## 1. Adding the packagecloud repository

packagecloud provides scripts to automate the process of configuring the package repository on your system, importing signing-keys etc.  These scripts must be run sudo root, and you should review them first.  The scripts are:

* apt/deb repositories: https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh
* yum/rpm repositories: https://packagecloud.io/install/repositories/github/git-lfs/script.rpm.sh

The scripts check your Linux distribution and version, and use those parameters to create the best repository URL.  If you are running one of the distributions listed for the latest version of Git LFS listed at [packagecloud](https://packagecloud.io/github/git-lfs) e.g `debian/jessie`, `el/7`, you can run the script without parameters:

apt/deb repos:
`curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh | sudo bash`

yum/rpm repos:
`curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.rpm.sh | sudo bash`

If you are running a distribution which does not match exactly a repository uploaded for Git LFS, but for which there is a repository for a compatible upstream distribution, you can either run the script with some additional parameters, or run it and then manually-correct the resulting repository URLs.  See [#1074](https://github.com/git-lfs/git-lfs/issues/1074) for details.

If you are running LinuxMint 17.1 Rebecca, which is downstream of Ubuntu Trusty and Debian Jessie, you can run:

`curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh | os=debian dist=jessie sudo -E bash`

The `os` and `dist` variables passed-in will override what would be detected for your system and force the selection of the upstream distribution's repository.

You may also be able to run the following to automatically detect the dist for Ubuntu based distributions such as Pop OS:
```
(. /etc/lsb-release &&
curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh |
sudo env os=ubuntu dist="${DISTRIB_CODENAME}" bash)
```

## 2. Installing packages

With the packagecloud repository configured for your system, you can install Git LFS:

* apt/deb: `sudo apt-get install git-lfs`
* yum/rpm: `sudo yum install git-lfs`

## A note about proxies

Several of the commands above assume internet access and use `sudo`. If your host is behind a proxy-server that is required for internet access, you may depend on environment-variables `http_proxy` or `https_proxy` being set, and these might not survive the switch to root with `sudo`, which resets environment by-default.  To get around this, you can run `sudo` with the `-E` switch, `sudo -E ...`, which retains environment variables.
