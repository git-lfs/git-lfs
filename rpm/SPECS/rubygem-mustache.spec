#global gemdir %(ruby -rubygems -e 'puts Gem::dir' 2>/dev/null)
%global gemdir %(IFS=: R=($(gem env gempath)); echo ${R[${#R[@]}-1]})
%define gem_name mustache

Name:           rubygem-%{gem_name}
Version:        1.0.1
Release:	1%{?dist}
Summary:        A framework-agnostic way to render logic-free views

Group:          Applications/Programming
License:        MIT
URL:		https://rubygems.org/gems/%{gem_name}
Source0:	https://rubygems.org/downloads/%{gem_name}-%{version}.gem
BuildRoot:      %(echo %{_topdir}/BUILDROOT/%{gem_name}-%{version})
BuildRequires:	gem > 2.0
Requires:       ruby > 2.0
BuildArch:      noarch

%description
Inspired by ctemplate, Mustache is a framework-agnostic way to render logic-free views. As ctemplates says, "It emphasizes separating logic from presentation: it is impossible to embed application logic in this template language. Think of Mustache as a replacement for your views. Instead of views consisting of ERB or HAML with random helpers and arbitrary logic, your views are broken into two parts: a Ruby class and an HTML template.

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
%{gem_libdir}
%exclude %{gem_cache}
/usr/share/gems
/usr/bin/%{gem_name}
%else
%{gemdir}
/usr/local/bin/%{gem_name}
%endif

%changelog
* Wed May 20 2015 Andrew Neff <andyneff@users.noreply.github.com> - 2.1.8
- Initial Spec
