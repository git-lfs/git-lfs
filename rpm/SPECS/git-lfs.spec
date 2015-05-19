Name:           git-lfs	
Version:        0.5.1
Release:	1%{?dist}
Summary:        Git extension for versioning large files

Group:          Applications/Archiving
License:        MIT
URL:		https://git-lfs.github.com/
Source0:	https://github.com/github/git-lfs/archive/v%{version}.tar.gz
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)
BuildRequires:	golang, tar, which, bison, rubygem-ronn
Requires:	git

%if 0%{?rhel} == 7
  #Umm... excuse me what?
  %define debug_package %{nil}
  #I think this is because go links with --build-id=none for linux
  #Uhhh... HOW DO I FIX THAT? Using an external linker
%endif

%description


%prep
%setup -q -n %{name}-%{version}

%build
./script/bootstrap
./script/man

%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
install -D bin/git-lfs ${RPM_BUILD_ROOT}/usr/bin/git-lfs
mkdir -p -m 755 ${RPM_BUILD_ROOT}/usr/share/man/man1
install -D man/*.1 ${RPM_BUILD_ROOT}/usr/share/man/man1

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
