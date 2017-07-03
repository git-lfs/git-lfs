Name:           ruby
Version:        2.2.2
Release:	1%{?dist}
Summary:        Ruby Programming Language

Group:          Applications/Programming
License:        BSDL
URL:		https://www.ruby-lang.org/
Source0:	http://cache.ruby-lang.org/pub/ruby/2.2/ruby-2.2.2.tar.gz
BuildRoot:      %(echo %{_topdir}/BUILDROOT/%{name}-%{version})
BuildRequires:	patch, libyaml-devel, glibc-headers, autoconf, gcc-c++, glibc-devel, patch, readline-devel, zlib-devel, libffi-devel, openssl-devel, automake, libtool, sqlite-devel
Provides:       gem = %{version}-%{release}

%description
A dynamic, open source programming language with a focus on simplicity and productivity. It has an elegant syntax that is natural to read and easy to write. 

%prep
%setup -q

%build
./configure
make -j 8

%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
make install DESTDIR=${RPM_BUILD_ROOT}

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
/

%changelog
* Tue May 19 2015 Andrew Neff <andyneff@users.noreply.github.com> - 2.2.2-1
- Initial Spec
