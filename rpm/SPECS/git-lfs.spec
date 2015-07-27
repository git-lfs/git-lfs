Name:           git-lfs	
Version:        0.5.2
Release:        1%{?dist}
Summary:        Git extension for versioning large files

Group:          Applications/Archiving
License:        MIT
URL:            https://git-lfs.github.com/
Source0:        https://github.com/github/git-lfs/archive/v%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
BuildRequires:  golang, tar, which, bison, rubygem-ronn, git
BuildRequires:  perl-Digest-SHA
Requires:       git

%if 0%{?rhel} == 7
  #Umm... excuse me what?
  %define debug_package %{nil}
  #I think this is because go links with --build-id=none for linux
  #Uhhh... HOW DO I FIX THAT? The answer is: go -ldflags '-linkmode=external'
%endif

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
GOPATH=`pwd` ./script/bootstrap
GOPATH=`pwd` ./script/man

%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
install -D bin/git-lfs ${RPM_BUILD_ROOT}/usr/bin/git-lfs
mkdir -p -m 755 ${RPM_BUILD_ROOT}/usr/share/man/man1
install -D man/*.1 ${RPM_BUILD_ROOT}/usr/share/man/man1

%check
if ! git config --global user.name; then
  RPM_GIT_USER_NAME=1
  git config --global user.name "User Name"
fi
if ! git config --global user.email; then
  RPM_GIT_USER_EMAIL=1
  git config --global user.email "user@email.com"
fi
#GOPATH=`pwd` ./script/test
#GOPATH=`pwd` ./script/integration
if [ "${RPM_GIT_USER_NAME}" == "1" ]; then
  git config --global --unset user.name
fi
if [ "${RPM_GIT_USER_EMAIL}" == "1" ]; then
  git config --global --unset user.email
fi

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%doc LICENSE README.md
/usr/bin/git-lfs
/usr/share/man/man1/*.1.gz

%changelog
* Mon May 18 2015 Andrew Neff <andyneff@users.noreply.github.com> - 0.5.1-1
- Initial Spec
