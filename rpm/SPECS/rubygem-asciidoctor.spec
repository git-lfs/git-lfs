#global gemdir %(ruby -rubygems -e 'puts Gem::dir' 2>/dev/null)
%global gemdir %(IFS=: R=($(gem env gempath)); echo ${R[${#R[@]}-1]})
%define gem_name asciidoctor

Name:           rubygem-%{gem_name}
Version:        2.0.17
Release:        1%{?dist}
Summary:        Builds manuals

Group:          Applications/Programming
License:        N/A
URL:            https://rubygems.org/gems/%{gem_name}
Source0:        https://rubygems.org/downloads/%{gem_name}-%{version}.gem
BuildRoot:      %(echo %{_topdir}/BUILDROOT/%{gem_name}-%{version})
%if 0%{?el7}
BuildRequires:  rh-ruby30-ruby, rh-ruby30-build
Requires:       rh-ruby30-ruby
%else
BuildRequires:  gem
Requires:       ruby
%endif
BuildArch:      noarch

%description
Builds Manuals

%prep
%if 0%{?el7}
%setup -q -c -T
%else
%setup -q -n %{gem_name}-%{version}
%endif
%if 0%{?el7}
mkdir -p ./usr/local
gem install -V --local --force --install-dir ./%{gemdir} --wrappers --bindir ./usr/local/bin %{SOURCE0}
%endif

%build
%if 0%{?el8}%{?el9}
gem build ../%{gem_name}-%{version}.gemspec
gem install -V --local --build-root . --force --no-document %{gem_name}-%{version}.gem
%endif

%install
mkdir -p ${RPM_BUILD_ROOT}
cp -a ./usr ${RPM_BUILD_ROOT}/usr
%if 0%{?el7}
cp -a ./opt ${RPM_BUILD_ROOT}/opt
%endif

%clean
rm -rf %{buildroot}

%files
%defattr(-,root,root,-)
%if 0%{?el8}%{?el9}
%dir %{gem_instdir}
%{gem_libdir}
%exclude %{gem_cache}
%{gem_spec}
/usr/share/gems
/usr/bin/%{gem_name}
%else
%{gemdir}/gems/%{gem_name}-%{version}
/opt/rh/rh-ruby30/root/usr/local/share/gems/cache/%{gem_name}-%{version}.gem
/opt/rh/rh-ruby30/root/usr/local/share/gems/doc/%{gem_name}-%{version}
/opt/rh/rh-ruby30/root/usr/local/share/gems/specifications/%{gem_name}-%{version}.gemspec
/usr/local/bin
%endif

%changelog
