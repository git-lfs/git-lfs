Name:           git-lfs
Version:        1.0.2
Release:        1%{?dist}
Summary:        Git extension for versioning large files

Group:          Applications/Archiving
License:        MIT
URL:            https://git-lfs.github.com/
Source0:        https://github.com/github/git-lfs/archive/%{name}-%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
BuildRequires:  golang, tar, which, bison, rubygem-ronn, git
BuildRequires:  perl-Digest-SHA

%define debug_package %{nil}
#I think this is because go links with --build-id=none for linux

%description
Git Large File Storage (LFS) replaces large files such as audio samples,
videos, datasets, and graphics with text pointers inside Git, while
storing the file contents on a remote server like GitHub.com or GitHub
Enterprise.

%prep
%setup -q -n %{name}-%{version}
mkdir -p src/github.com/github
ln -s $(pwd) src/github.com/github/%{name}

%build
%if %{_arch} == i386
  GOARCH=386 GOPATH=`pwd` ./script/bootstrap
%else
  GOARCH=amd64 GOPATH=`pwd` ./script/bootstrap
%endif
GOPATH=`pwd` ./script/man

%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
install -D bin/git-lfs ${RPM_BUILD_ROOT}/usr/bin/git-lfs
mkdir -p -m 755 ${RPM_BUILD_ROOT}/usr/share/man/man1
install -D man/*.1 ${RPM_BUILD_ROOT}/usr/share/man/man1

%check
GOPATH=`pwd` ./script/test
GOPATH=`pwd` ./script/integration

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%doc LICENSE.md README.md
/usr/bin/git-lfs
/usr/share/man/man1/*.1.gz

%changelog
* Sun Aug 2 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.4-1
- Added tests back in

* Sat Jul 18 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.2-1
- Changed Source0 filename

* Mon May 18 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.1-1
- Initial Spec
