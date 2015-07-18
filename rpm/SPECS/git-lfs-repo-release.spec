Name:           git-lfs-repo-release
Version:        1
Release:        1%{?dist}
Summary:        Packges for git-lfs for Enterprise Linux repository configuration

Group:          System Environment/Base
License:        MIT
%if 0%{?fedora} 
URL:            https://git-lfs.github.com/fedora/%{fedora}/
%elseif 0%{?rhel}
URL:            https://git-lfs.github.com/centos/%{rhel}/
%endif
Source0:        RPM-GPG-KEY-GITLFS
Source1:        git-lfs.repo
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

BuildArch:      noarch

%description
This package contains the Extra Packages for Enterprise Linux (EPEL) repository
GPG key as well as configuration for yum.

%prep
%setup -q -c -T

%build


%install
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT

#GPG Key
install -Dpm 644 %{SOURCE0} \
    $RPM_BUILD_ROOT%{_sysconfdir}/pki/rpm-gpg/RPM-GPG-KEY-GITLFS

# yum
install -dm 755 $RPM_BUILD_ROOT%{_sysconfdir}/yum.repos.d
install -pm 644 %{SOURCE1}  \
    $RPM_BUILD_ROOT%{_sysconfdir}/yum.repos.d

%clean
[ "$RPM_BUILD_ROOT" != "/" ] && rm -rf $RPM_BUILD_ROOT

%files
%defattr(-,root,root,-)
%config(noreplace) /etc/yum.repos.d/*
/etc/pki/rpm-gpg/*
