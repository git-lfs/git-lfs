#global gemdir %(ruby -rubygems -e 'puts Gem::dir' 2>/dev/null)
%global gemdir %(IFS=: R=($(gem env gempath)); echo ${R[${#R[@]}-1]})
%define gem_name ronn

Name:           rubygem-%{gem_name}
Version:        0.7.3
Release:	1%{?dist}
Summary:        Builds manuals

Group:          Applications/Programming
License:        N/A
URL:		https://rubygems.org/gems/%{gem_name}
Source0:	https://rubygems.org/downloads/%{gem_name}-%{version}.gem
BuildRoot:      %(echo %{_topdir}/BUILDROOT/%{gem_name}-%{version})
BuildRequires:	gem
Requires:       ruby
Requires:       rubygem-hpricot >= 0.8.2
Requires:       rubygem-mustache >= 0.7.0
Requires:       rubygem-rdiscount >= 1.5.8
BuildArch:      noarch

%description
Builds Manuals

%prep
%setup -q -c -T
gem install -V --local --force --install-dir ./%{gemdir} %{SOURCE0}
mv ./%{gemdir}/bin ./usr/local

%build

%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT
mkdir -p ${RPM_BUILD_ROOT}
cp -a ./usr ${RPM_BUILD_ROOT}/usr

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%{gemdir}
/usr/local/bin/%{gem_name}

%changelog
* Wed May 20 2015 Andrew Neff <andyneff@users.noreply.github.com> - 2.1.8
- Initial Spec
