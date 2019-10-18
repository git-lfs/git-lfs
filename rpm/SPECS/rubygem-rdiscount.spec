#global gemdir %(ruby -rubygems -e 'puts Gem::dir' 2>/dev/null)
%global gemdir %(IFS=: R=($(gem env gempath)); echo ${R[${#R[@]}-1]})
%define gem_name rdiscount
Name:           rubygem-%{gem_name}
Version:        2.1.8
Release:	1%{?dist}
Summary:        Fast Implementation of Gruber's Markdown in C

Group:          Applications/Programming
License:        BSD
URL:		https://rubygems.org/gems/%{gem_name}
Source0:	https://rubygems.org/downloads/%{gem_name}-%{version}.gem
BuildRoot:      %(echo %{_topdir}/BUILDROOT/%{gem_name}-%{version})
BuildRequires:	gem > 1.9.2
Requires:       ruby > 1.9.2

%description
Fast Implementation of Gruber's Markdown in C

%prep
%if 0%{?el6}
%setup -q -c -T
%else
%setup -q -n %{gem_name}-%{version}
%endif
%if ! 0%{?el8}
gem install -V --local --force --install-dir ./%{gemdir} %{SOURCE0}
mv ./%{gemdir}/bin ./usr/local
%endif

%build
%if 0%{?el8}
gem build ../%{gem_name}-%{version}.gemspec
%gem_install
%endif

%install
mkdir -p ${RPM_BUILD_ROOT}
cp -a ./usr ${RPM_BUILD_ROOT}/usr

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%if 0%{?el8}
%dir %{gem_instdir}
%{gem_extdir_mri}
%{gem_libdir}
%exclude %{gem_cache}
/usr/share/gems
/usr/bin/%{gem_name}
%{gem_spec}
%else
%{gemdir}
/usr/local/bin/%{gem_name}
%endif

%changelog
* Wed May 20 2015 Andrew Neff <andyneff@users.noreply.github.com> - 2.1.8
- Initial Spec
