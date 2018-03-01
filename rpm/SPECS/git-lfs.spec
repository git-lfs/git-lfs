Name:           git-lfs
Version:        2.4.0
Release:        1%{?dist}
Summary:        Git extension for versioning large files

Group:          Applications/Archiving
License:        MIT
URL:            https://git-lfs.github.com/
Source0:        https://github.com/git-lfs/git-lfs/archive/v%{version}/%{name}-%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
BuildRequires:  perl-Digest-SHA
BuildRequires:  golang, tar, rubygem-ronn, which, git >= 1.8.2

Requires: git >= 1.8.2

%define debug_package %{nil}
#I think this is because go links with --build-id=none for linux

%description
Git Large File Storage (LFS) replaces large files such as audio samples,
videos, datasets, and graphics with text pointers inside Git, while
storing the file contents on a remote server like GitHub.com or GitHub
Enterprise.

%prep
%setup -q -n %{name}-%{version}
export GOPATH=`pwd`
mkdir -p src/github.com/git-lfs
ln -s $(pwd) src/github.com/git-lfs/%{name}

%build
%if 0%{?rhel} == 5
  export CGO_ENABLED=0
%endif

pushd src/github.com/git-lfs/%{name}
  %if %{_arch} == i386
    GOARCH=386 ./script/bootstrap
  %else
    GOARCH=amd64 ./script/bootstrap
  %endif
popd
./script/man

%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
install -D bin/git-lfs ${RPM_BUILD_ROOT}/usr/bin/git-lfs
mkdir -p -m 755 ${RPM_BUILD_ROOT}/usr/share/man/man1
mkdir -p -m 755 ${RPM_BUILD_ROOT}/usr/share/man/man5
install -D man/*.1 ${RPM_BUILD_ROOT}/usr/share/man/man1
install -D man/*.5 ${RPM_BUILD_ROOT}/usr/share/man/man5

%post
git lfs install --system

%preun
git lfs uninstall

%check
export GOPATH=`pwd`
export GIT_LFS_TEST_DIR=$(mktemp -d)

# test/git-lfs-server-api/main.go does not compile because github.com/spf13/cobra
# cannot be found in vendor, for some reason. It's not needed for installs, so
# skip it.
export SKIPAPITESTCOMPILE=1

pushd src/github.com/git-lfs/%{name}
  ./script/test
  go get github.com/ThomsonReutersEikon/go-ntlm/ntlm
  ./script/integration
popd

rmdir ${GIT_LFS_TEST_DIR}

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%doc LICENSE.md README.md
/usr/bin/git-lfs
/usr/share/man/man1/*.1.gz
/usr/share/man/man5/*.5.gz

%changelog
* Sun Dec 6 2015 Andrew Neff <andyneff@users.noreply.github.com> - 1.1.0-1
- Added Requires and version for git back in

* Sat Oct 31 2015 Andrew Neff <andyneff@users.noreply.github.com> - 1.0.3-1
- Added GIT_LFS_TEST_DIR to prevent future test race condition

* Sun Aug 2 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.4-1
- Added tests back in

* Sat Jul 18 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.2-1
- Changed Source0 filename

* Mon May 18 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.1-1
- Initial Spec
