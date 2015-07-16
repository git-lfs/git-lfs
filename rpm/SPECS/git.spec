%define gitcoredir %{_libexecdir}/git-core

# Pass --without docs to rpmbuild if you don't want the documentation
Name:           git
Version:        2.4.5
Release:        1%{?dist}
Summary:        Fast Version Control System
License:        GPLv2
Group:          Development/Tools
URL:            http://git-scm.com/
Source0:        http://kernel.org/pub/software/scm/git/%{name}-%{version}.tar.xz
Source1:        git-init.el
Source2:        git.xinetd.in
Source3:        git.conf.httpd
Source4:        git-gui.desktop
Source5:        gitweb.conf.in
BuildRoot:      %{_tmppath}/%{name}-%{version}-%{release}-root-%(%{__id_u} -n)

%if 0%{?fedora} || 0%{?rhel} >= 6
BuildRequires:  emacs >= 22.2
BuildRequires:  libcurl-devel
%else
BuildRequires:  curl-devel
%endif
BuildRequires:  expat-devel
BuildRequires:  gettext
BuildRequires:  openssl-devel
BuildRequires:  zlib-devel >= 1.2
%{!?_without_docs:BuildRequires: asciidoc > 6.0.3, xmlto}
BuildRequires:  perl-ExtUtils-MakeMaker

Requires:       less
Requires:       openssh-clients
Requires:       perl(Error)
Requires:       perl-Git = %{version}-%{release}
Requires:       rsync
Requires:       zlib >= 1.2

Provides:       git-core = %{version}-%{release}
Obsoletes:      git-core < 2.4.4

%description
Git is a fast, scalable, distributed revision control system with an
unusually rich command set that provides both high-level operations
and full access to internals.

The git rpm installs the core tools with minimal dependencies.  To
install all git packages, including tools for integrating with other
SCMs, install the git-all meta-package.

%package all
Summary:        Meta-package to pull in all git tools
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
%if 0%{?fedora}
Requires:       git-arch = %{version}-%{release}
%endif
Requires:       git = %{version}-%{release}
Requires:       git-svn = %{version}-%{release}
Requires:       git-cvs = %{version}-%{release}
Requires:       git-email = %{version}-%{release}
Requires:       gitk = %{version}-%{release}
Requires:       git-gui = %{version}-%{release}
Requires:       perl-Git = %{version}-%{release}
%if 0%{?fedora} || 0%{?rhel} >= 6
Requires:       emacs-git = %{version}-%{release}
%endif
Obsoletes:      git < 2.4.4
%description all
Git is a fast, scalable, distributed revision control system with an
unusually rich command set that provides both high-level operations
and full access to internals.

This is a dummy package which brings in all subpackages.

%package daemon
Summary:        Git protocol dæmon
Group:          Development/Tools
Requires:       git = %{version}-%{release}, xinetd
%description daemon
The git dæmon for supporting git:// access to git repositories

%package -n gitweb
Summary:        Simple web interface to git repositories
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}

%description -n gitweb
Simple web interface to track changes in git repositories


%package svn
Summary:        Git tools for importing Subversion repositories
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, subversion, perl(Term::ReadKey)
%description svn
Git tools for importing Subversion repositories.

%package cvs
Summary:        Git tools for importing CVS repositories
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, cvs, cvsps
%description cvs
Git tools for importing CVS repositories.

%if 0%{?fedora}
%package arch
Summary:        Git tools for importing Arch repositories
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, tla
%description arch
Git tools for importing Arch repositories.
%endif

%package email
Summary:        Git tools for sending email
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, perl-Git = %{version}-%{release}
Requires:       perl(Net::SMTP::SSL), perl(Authen::SASL)
%description email
Git tools for sending email.

%package gui
Summary:        Git GUI tool
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, tk >= 8.4
Requires:       gitk = %{version}-%{release}
%description gui
Git GUI tool.

%package -n gitk
Summary:        Git revision tree visualiser
Group:          Development/Tools
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, tk >= 8.4
%description -n gitk
Git revision tree visualiser.

%package -n perl-Git
Summary:        Perl interface to Git
Group:          Development/Libraries
%if 0%{?fedora} >= 10 || 0%{?rhel} >= 6
BuildArch:      noarch
%endif
Requires:       git = %{version}-%{release}, perl(Error)
Requires:       perl(:MODULE_COMPAT_%(eval "`%{__perl} -V:version`"; echo $version))
BuildRequires:  perl(Error), perl(ExtUtils::MakeMaker)

%description -n perl-Git
Perl interface to Git.

%if 0%{?fedora} || 0%{?rhel} >= 6
%package -n emacs-git
Summary:        Git version control system support for Emacs
Group:          Applications/Editors
BuildArch:      noarch
Requires:       git = %{version}-%{release}, emacs(bin) >= %{_emacs_version}

%description -n emacs-git
%{summary}.
%endif

%package -n emacs-git-el
Summary:        Elisp source files for git version control system support for Emacs
Group:          Applications/Editors
BuildArch:      noarch
Requires:       emacs-git = %{version}-%{release}

%description -n emacs-git-el
%{summary}.

%prep
%setup -q

# Use these same options for every invocation of 'make'.
# Otherwise it will rebuild in %%install due to flags changes.
cat << \EOF > config.mak
V = 1
CFLAGS = %{optflags}
BLK_SHA1 = 1
NEEDS_CRYPTO_WITH_SSL = 1
NO_PYTHON = 1
ETC_GITCONFIG = %{_sysconfdir}/gitconfig
DESTDIR = %{buildroot}
INSTALL = install -p
GITWEB_PROJECTROOT = %{_var}/lib/git
htmldir = %{_docdir}/%{name}-%{version}
prefix = %{_prefix}
ASCIIDOC8 = 1
ASCIIDOC_NO_ROFF = 1
EOF

# Filter bogus perl requires
# packed-refs comes from a comment in contrib/hooks/update-paranoid
cat << \EOF > %{name}-req
#!/bin/sh
%{__perl_requires} $* |\
sed -e '/perl(packed-refs)/d'
EOF

%global __perl_requires %{_builddir}/%{name}-%{version}/%{name}-req
chmod +x %{__perl_requires}

%build
make %{?_smp_mflags} all %{!?_without_docs: doc}

%if 0%{?fedora} || 0%{?rhel} >= 6
make -C contrib/emacs
%endif

# Remove shebang from bash-completion script
sed -i '/^#!bash/,+1 d' contrib/completion/git-completion.bash

%install
rm -rf %{buildroot}
make %{?_smp_mflags} INSTALLDIRS=vendor install %{!?_without_docs: install-doc}

%if 0%{?fedora} || 0%{?rhel} >= 6
%global elispdir %{_emacs_sitelispdir}/git
make -C contrib/emacs install \
    emacsdir=%{buildroot}%{elispdir}
for elc in %{buildroot}%{elispdir}/*.elc ; do
    install -pm 644 contrib/emacs/$(basename $elc .elc).el \
    %{buildroot}%{elispdir}
done
install -Dpm 644 %{SOURCE1} \
    %{buildroot}%{_emacs_sitestartdir}/git-init.el
%endif

mkdir -p %{buildroot}%{_var}/www/git/static/js/lib
install -pm 644 gitweb/static/*.{css,js,png} %{buildroot}%{_var}/www/git/static
install -pm 644 gitweb/static/js/*.js %{buildroot}%{_var}/www/git/static/js
install -pm 644 gitweb/static/js/lib/*.js %{buildroot}%{_var}/www/git/static/js/lib
install -pm 755 gitweb/gitweb.cgi %{buildroot}%{_var}/www/git
install -pm 755 gitweb/gitweb.perl %{buildroot}%{_var}/www/git
mkdir -p %{buildroot}%{_sysconfdir}/httpd/conf.d
install -pm 0644 %{SOURCE3} %{buildroot}%{_sysconfdir}/httpd/conf.d/git.conf
sed "s|@PROJECTROOT@|%{_var}/lib/git|g" \
    %{SOURCE5} > %{buildroot}%{_sysconfdir}/gitweb.conf

find %{buildroot} -type f -name .packlist -exec rm -f {} ';'
find %{buildroot} -type f -name '*.bs' -empty -exec rm -f {} ';'
find %{buildroot} -type f -name perllocal.pod -exec rm -f {} ';'

%if ! 0%{?fedora}
find $RPM_BUILD_ROOT Documentation -type f -name 'git-archimport*' -exec rm -f {} ';'
%endif

(find %{buildroot}{%{_bindir},%{_libexecdir}} -type f | grep -vE "archimport|svn|cvs|email|gitk|git-gui|git-citool|git-daemon" | sed -e s@^%{buildroot}@@) > bin-man-doc-files
(find %{buildroot}%{perl_vendorlib} -type f | sed -e s@^%{buildroot}@@) >> perl-files
(find %{buildroot}%{_datadir} -type f -name \*.mo | sed -e s@^%{buildroot}@@ ) >> bin-man-doc-files
%if %{!?_without_docs:1}0
(find %{buildroot}%{_mandir} -type f | grep -vE "archimport|svn|git-cvs|email|gitk|git-gui|git-citool|git-daemon" | sed -e s@^%{buildroot}@@ -e 's/$/*/' ) >> bin-man-doc-files
%else
rm -rf %{buildroot}%{_mandir}
%endif

mkdir -p %{buildroot}%{_var}/lib/git
mkdir -p %{buildroot}%{_sysconfdir}/xinetd.d
perl -p \
    -e "s|\@GITCOREDIR\@|%{gitcoredir}|g;" \
    -e "s|\@BASE_PATH\@|%{_var}/lib/git|g;" \
    %{SOURCE2} > %{buildroot}%{_sysconfdir}/xinetd.d/git

mkdir -p %{buildroot}%{_sysconfdir}/bash_completion.d
install -pm 644 contrib/completion/git-completion.bash %{buildroot}%{_sysconfdir}/bash_completion.d/git

# Move contrib/hooks out of %%docdir and make them executable
mkdir -p %{buildroot}%{_datadir}/git-core/contrib
mv contrib/hooks %{buildroot}%{_datadir}/git-core/contrib
chmod +x %{buildroot}%{_datadir}/git-core/contrib/hooks/*
pushd contrib > /dev/null
ln -s ../../../git-core/contrib/hooks
popd > /dev/null

# install git-gui .desktop file
desktop-file-install \
%if 0%{?rhel} && 0%{?rhel} <= 5
    --vendor fedora \
%endif
    --dir=%{buildroot}%{_datadir}/applications %{SOURCE4}

# quiet some rpmlint complaints
chmod -R g-w %{buildroot}
find %{buildroot} -name git-mergetool--lib | xargs chmod a-x
rm -f {Documentation/technical,contrib/emacs}/.gitignore
chmod a-x Documentation/technical/api-index.sh
find contrib -type f | xargs chmod -x

rm -f %{buildroot}%{gitcoredir}/git-remote-testsvn
rm -rf %{buildroot}%{_datadir}/gitweb

%clean
rm -rf %{buildroot}


%files -f bin-man-doc-files
%defattr(-,root,root)
%{_datadir}/git-core/
%dir %{gitcoredir}
%doc README COPYING Documentation/*.txt contrib/
%{!?_without_docs: %doc Documentation/*.html Documentation/docbook-xsl.css}
%{!?_without_docs: %doc Documentation/howto Documentation/technical}
%{_sysconfdir}/bash_completion.d


%files svn
%defattr(-,root,root)
%{gitcoredir}/*svn*
%doc Documentation/*svn*.txt
%{!?_without_docs: %{_mandir}/man1/*svn*.1*}
%{!?_without_docs: %doc Documentation/*svn*.html }

%files cvs
%defattr(-,root,root)
#%doc Documentation/*git-cvs*.txt
%{_bindir}/git-cvsserver
%{gitcoredir}/*cvs*
%{!?_without_docs: %{_mandir}/man1/*cvs*.1*}
#%{!?_without_docs: %doc Documentation/*git-cvs*.html }

%if 0%{?fedora}
%files arch
%defattr(-,root,root)
%doc Documentation/git-archimport.txt
%{gitcoredir}/git-archimport
%{!?_without_docs: %{_mandir}/man1/git-archimport.1*}
%{!?_without_docs: %doc Documentation/git-archimport.html }
%endif

%files email
%defattr(-,root,root)
#%doc Documentation/*email*.txt
%{gitcoredir}/*email*
%{!?_without_docs: %{_mandir}/man1/*email*.1*}
#%{!?_without_docs: %doc Documentation/*email*.html }

%files gui
%defattr(-,root,root)
%{gitcoredir}/git-gui*
%{gitcoredir}/git-citool
%{_datadir}/applications/*git-gui.desktop
%{_datadir}/git-gui/
%{!?_without_docs: %{_mandir}/man1/git-gui.1*}
%{!?_without_docs: %doc Documentation/git-gui.html}
%{!?_without_docs: %{_mandir}/man1/git-citool.1*}
%{!?_without_docs: %doc Documentation/git-citool.html}

%files -n gitk
%defattr(-,root,root)
%doc Documentation/*gitk*.txt
%{_bindir}/*gitk*
%{_datadir}/gitk
%{!?_without_docs: %{_mandir}/man1/*gitk*.1*}
%{!?_without_docs: %doc Documentation/*gitk*.html }

%files -n perl-Git -f perl-files
%defattr(-,root,root)

%if 0%{?fedora} || 0%{?rhel} >= 6
%files -n emacs-git
%defattr(-,root,root)
%doc contrib/emacs/README
%dir %{elispdir}
%{elispdir}/*.elc
%{_emacs_sitestartdir}/git-init.el

%files -n emacs-git-el
%defattr(-,root,root)
%{elispdir}/*.el
%endif

%files daemon
%defattr(-,root,root)
%doc Documentation/*daemon*.txt
%config(noreplace)%{_sysconfdir}/xinetd.d/git
%{gitcoredir}/git-daemon
%{_var}/lib/git
%{!?_without_docs: %{_mandir}/man1/*daemon*.1*}
%{!?_without_docs: %doc Documentation/*daemon*.html}

%files -n gitweb
%defattr(-,root,root)
%doc gitweb/INSTALL gitweb/README
%{_var}/www/git/
%config(noreplace)%{_sysconfdir}/gitweb.conf
%config(noreplace)%{_sysconfdir}/httpd/conf.d/git.conf


%files all
# No files for you!

%changelog
* Wed Jun 08 2015 Andy Neff <andrew.neff visionsystemsinc com> 2.4.5
- Update to version 2.4.5. Did not keep any patches from 1.7.1-3.1
- Had to comment out a few docuemnts.

* Tue Feb 26 2013 Adam Tkac <atkac redhat com> 1.7.1-3.1
- fix CVE-2013-0308

* Thu Dec 16 2010 Adam Tkac <atkac redhat com> 1.7.1-3
- fix CVE-2010-3906

* Tue Jul 27 2010 Adam Tkac <atkac redhat com> 1.7.1-2
- fix CVE-2010-2542 (#618108)

* Thu Jun 10 2010 Adam Tkac <atkac redhat com> - 1.7.1-1
- update to 1.7.1 (#585990)
- Use %%{gitcoredir}/git-daemon as xinetd server (#529682) [Todd Zullinger]
- Remove long fixed xinetd IPv6 workaround (#557528) [Todd Zullinger]
- Make %%{_var}/lib/git the default gitweb root (#556299) [Todd Zullinger]
- Include gitweb/INSTALL file as documentation [Todd Zullinger]
- Ship example gitweb config (%%{_sysconfdir}/gitweb.conf) [Todd Zullinger]
- Install missing gitweb.js (#558740) [Todd Zullinger]
- Only BR perl(Error) [Todd Zullinger]
- Use config.mak to set build options [Todd Zullinger]
- Disable building of unused python remote helper libs [Todd Zullinger]
- Comply with Emacs add-on packaging guidelines (#573423) [Jonathan Underwood]
  - Place elisp source files in separate emacs-git-el package
  - Place git support files in own directory under site-lisp
  - Use Emacs packaging macros
- Replace $RPM_BUILD_ROOT with %%{buildroot} [Todd Zullinger]

* Thu Dec 03 2009 Dennis Gregorovic <dgregor@redhat.com> - 1.6.5.2-1.2
- Rebuilt for RHEL 6

* Fri Nov 13 2009 Dennis Gregorovic <dgregor@redhat.com> - 1.6.5.2-1.1
- Fix conditional for RHEL

* Mon Oct 26 2009 Todd Zullinger <tmz@pobox.com> - 1.6.5.2-1
- git-1.6.5.2
- Drop asciidoc --unsafe option, it should not be needed anymore
- Don't use install -t/-T, they're not compatible with older coreutils
- Don't use -perm /a+x with find, it's incompatible with older findutils

* Sat Oct 17 2009 Todd Zullinger <tmz@pobox.com> - 1.6.5.1-1
- git-1.6.5.1

* Sun Oct 11 2009 Todd Zullinger <tmz@pobox.com> - 1.6.5-1
- git-1.6.5

* Mon Sep 28 2009 Todd Zullinger <tmz@pobox.com> - 1.6.5-0.2.rc2
- git-1.6.5.rc2
- Enable Linus' block-sha1 implementation

* Wed Sep 16 2009 Todd Zullinger <tmz@pobox.com> - 1.6.4.4-1
- git-1.6.4.4

* Sun Sep 13 2009 Todd Zullinger <tmz@pobox.com> - 1.6.4.3-1
- git-1.6.4.3

* Sun Aug 30 2009 Todd Zullinger <tmz@pobox.com> - 1.6.4.2-1
- git-1.6.4.2

* Sat Aug 22 2009 Todd Zullinger <tmz@pobox.com> - 1.6.4.1-1
- git-1.6.4.1

* Fri Aug 21 2009 Tomas Mraz <tmraz@redhat.com> - 1.6.4-2
- rebuilt with new openssl

* Wed Jul 29 2009 Todd Zullinger <tmz@pobox.com> - 1.6.4-1
- git-1.6.4

* Fri Jul 24 2009 Fedora Release Engineering <rel-eng@lists.fedoraproject.org> - 1.6.3.3-2
- Rebuilt for https://fedoraproject.org/wiki/Fedora_12_Mass_Rebuild

* Sun Jun 28 2009 Todd Zullinger <tmz@pobox.com> - 1.6.3.3-1
- git-1.6.3.3
- Move contributed hooks to %%{_datadir}/git-core/contrib/hooks (bug 500137)
- Fix rpmlint warnings about Summary and git-mergetool--lib missing shebang

* Fri Jun 19 2009 Todd Zullinger <tmz@pobox.com> - 1.6.3.2-3
- Temporarily disable asciidoc's safe mode until bug 506953 is fixed

* Fri Jun 19 2009 Todd Zullinger <tmz@pobox.com> - 1.6.3.2-2
- Fix git-daemon hang on invalid input (CVE-2009-2108, bug 505761)

* Fri Jun 05 2009 Todd Zullinger <tmz@pobox.com> - 1.6.3.2-1
- git-1.6.3.2
- Require emacs >= 22.2 for emacs support (bug 495312)
- Add a .desktop file for git-gui (bug 498801)
- Set ASCIIDOC8 and ASCIIDOC_NO_ROFF to correct documentation issues,
  the sed hack to fix bug 485161 should no longer be needed
- Escape newline in git-daemon xinetd description (bug 502393)
- Add xinetd to git-daemon Requires (bug 504105)
- Organize BuildRequires/Requires, drop redundant expat Requires
- Only build noarch subpackages on Fedora >= 10
- Only build emacs and arch subpackages on Fedora
- Handle curl/libcurl naming for EPEL and Fedora

* Fri Apr 03 2009 Todd Zullinger <tmz@pobox.com> - 1.6.2.2-1
- git-1.6.2.2
- Include contrib/ dir in %%doc (bug 492490)
- Don't set DOCBOOK_XSL_172, fix the '\&.ft' with sed (bug 485161)
- Ignore Branches output from cvsps-2.2b1 (bug 490602)
- Remove shebang from bash-completion script
- Include README in gitweb subpackage

* Mon Mar 09 2009 Todd Zullinger <tmz@pobox.com> - 1.6.2-1
- git-1.6.2
- Include contrib/emacs/README in emacs subpackage
- Drop upstreamed git-web--browse patch

* Tue Feb 24 2009 Todd Zullinger <tmz@pobox.com> - 1.6.1.3-2
- Require perl(Authen::SASL) in git-email (bug 483062)
- Build many of the subpackages as noarch
- Update URL field

* Mon Feb 09 2009 Todd Zullinger <tmz@pobox.com> 1.6.1.3-1
- git-1.6.1.3
- Set htmldir so "git help -w <command>" works
- Patch git-web--browse to not use "/sbin/start" to browse
- Include git-daemon documentation in the git-daemon package

* Thu Jan 29 2009 Josh Boyer <jwboyer@gmail.com> 1.6.1.2-1
- git-1.6.1.2

* Mon Jan 26 2009 Todd Zullinger <tmz@pobox.com> 1.6.1.1-1
- git-1.6.1.1
- Make compile more verbose

* Fri Jan 16 2009 Tomas Mraz <tmraz@redhat.com> 1.6.1-2
- rebuild with new openssl

* Sat Jan 03 2009 Todd Zullinger <tmz@pobox.com> 1.6.1-1
- Install git-* commands in %%{_libexecdir}/git-core, the upstream default
- Remove libcurl from Requires, rpm will pick this up automatically
- Consolidate build/install options in %%make_git (Roland McGrath)
- Include DirectoryIndex in gitweb httpd-config (bug 471692)
- Define DOCBOOK_XSL_172 to fix minor manpage issues
- Rename %%{_var}/lib/git-daemon to %%{_var}/lib/git
- Preserve timestamps on installed files
- Quiet some rpmlint complaints
- Use macros more consistently

* Sat Dec 20 2008 Todd Zullinger <tmz@pobox.com> 1.6.0.6-1
- git-1.6.0.6
- Fixes a local privilege escalation bug in gitweb
  (http://article.gmane.org/gmane.comp.version-control.git/103624)
- Add gitk Requires to git-gui (bug 476308)

* Thu Dec 11 2008 Josh Boyer <jboyer@gmail.com> 1.6.0.5-1
- git-1.6.0.5

* Mon Nov 17 2008 Seth Vidal <skvidal at fedoraproject.org>
- switch from /srv/git to /var/lib/git-daemon for packaging rules compliance

* Fri Nov 14 2008 Josh Boyer <jwboyer@gmail.com> 1.6.0.4-1
- git-1.6.0.4

* Wed Oct 22 2008 Josh Boyer <jwboyer@gmail.com> 1.6.0.3-1
- git-1.6.0.3
- Drop curl requirement in favor of libcurl (bug 449388)
- Add requires for SMTP-SSL perl module to make git-send-email work (bug 443615)

* Thu Aug 28 2008 James Bowes <jbowes@redhat.com> 1.6.0.1-1
- git-1.6.0.1

* Thu Jul 24 2008 James Bowes <jbowes@redhat.com> 1.5.6-4
- git-1.5.6.4

* Thu Jun 19 2008 James Bowes <jbowes@redhat.com> 1.5.6-1
- git-1.5.6

* Tue Jun  3 2008 Stepan Kasal <skasal@redhat.com> 1.5.5.3-2
- use tar.bz2 instead of tar.gz

* Wed May 28 2008 James Bowes <jbowes@redhat.com> 1.5.5.3-1
- git-1.5.5.3

* Mon May 26 2008 James Bowes <jbowes@redhat.com> 1.5.5.2-1
- git-1.5.5.2

* Mon Apr 21 2008 James Bowes <jbowes@redhat.com> 1.5.5.1-1
- git-1.5.5.1

* Wed Apr 09 2008 James Bowes <jbowes@redhat.com> 1.5.5-1
- git-1.5.5

* Fri Apr 04 2008 James Bowes <jbowes@redhat.com> 1.5.4.5-3
- Remove the last two requires on git-core.

* Wed Apr 02 2008 James Bowes <jbowes@redhat.com> 1.5.4.5-2
- Remove a patch that's already upstream.

* Fri Mar 28 2008 James Bowes <jbowes@redhat.com> 1.5.4.5-1
- git-1.5.4.5

* Wed Mar 26 2008 James Bowes <jbowes@redhat.com> 1.5.4.4-4
- Own /etc/bash_completion.d in case bash-completion isn't installed.

* Tue Mar 25 2008 James Bowes <jbowes@redhat.com> 1.5.4.4-3
- Include the sample hooks from contrib/hooks as docs (bug 321151).
- Install the bash completion script from contrib (bug 433255).
- Include the html docs in the 'core' package again (bug 434271).

* Wed Mar 19 2008 James Bowes 1.5.4.4-2
- Obsolete git <= 1.5.4.3, to catch going from F8 to rawhide/F9

* Thu Mar 13 2008 James Bowes <jbowes@redhat.com> 1.5.4.4-1
- git-1.5.4.4

* Mon Mar  3 2008 Tom "spot" Callaway <tcallawa@redhat.com> 1.5.4.3-3
- rebuild for new perl (again)

* Sun Feb 24 2008 Bernardo Innocenti <bernie@codewiz.org> 1.5.4.3-2
- Do not silently overwrite /etc/httpd/conf.d/git.conf

* Sat Feb 23 2008 James Bowes <jbowes@redhat.com> 1.5.4.3-1
- git-1.5.4.3
- Include Kristian Høgsberg's changes to rename git-core to
  git and git to git-all.

* Sun Feb 17 2008 James Bowes <jbowes@redhat.com> 1.5.4.2-1
- git-1.5.4.2

* Mon Feb 11 2008 Jeremy Katz <katzj@redhat.com> - 1.5.4.1-2
- Add upstream patch (e62a641de17b172ffc4d3a803085c8afbfbec3d1) to have 
  gitweb rss feeds point be commitdiffs instead of commit

* Sun Feb 10 2008 James Bowes <jbowes@redhat.com> 1.5.4.1-1
- git-1.5.4.1

* Tue Feb 05 2008 Tom "spot" Callaway <tcallawa@redhat.com> 1.5.4-3
- rebuild for new perl

* Sun Feb 03 2008 James Bowes <jbowes@redhat.com> 1.5.4-1
- Add BuidRequires on gettext.

* Sat Feb 02 2008 James Bowes <jbowes@redhat.com> 1.5.4-1
- git-1.5.4

* Tue Jan 08 2008 James Bowes <jbowes@redhat.com> 1.5.3.8-1
- git-1.5.3.8

* Fri Dec 21 2007 James Bowes <jbowes@redhat.com> 1.5.3.7-2
- Have git metapackage require explicit versions (bug 247214)

* Mon Dec 03 2007 Josh Boyer <jwboyer@gmail.com> 1.5.3.7-1
- git-1.5.3.7

* Tue Nov 27 2007 Josh Boyer <jwboyer@gmail.com> 1.5.3.6-1
- git-1.5.3.6
- git-core requires perl(Error) (bug 367861)
- git-svn requires perl(Term:ReadKey) (bug 261361)
- git-email requires perl-Git (bug 333061)

* Wed Oct 24 2007 Lubomir Kundrak <lkundrak@redhat.com> 1.5.3.4-2
- git-Perl requires Error package

* Tue Oct 09 2007 James Bowes <jbowes@redhat.com> 1.5.3.4-1
- git-1.5.3.4

* Sun Sep 30 2007 James Bowes <jbowes@redhat.com> 1.5.3.3-1
- git-1.5.3.3

* Wed Sep 26 2007 James Bowes <jbowes@redhat.com> 1.5.3.2-1
- git-1.5.3.2

* Thu Sep 06 2007 Josh Boyer <jwboyer@jdub.homelinux.org> 1.5.3.1-2
- Include git-gui and git-citool docs

* Thu Sep 06 2007 Josh Boyer <jwboyer@jdub.homelinux.org> 1.5.3.1-1
- git-1.5.3.1-1

* Thu Aug 23 2007 James Bowes <jbowes@redhat.com> 1.5.2.5-1
- git-1.5.2.5-1

* Fri Aug 03 2007 Josh Boyer <jwboyer@jdub.homelinux.org> 1.5.2.4-1
- git-1.5.2.4-1

* Tue Jul 03 2007 Josh Boyer <jwboyer@jdub.homelinux.org> 1.5.2.2-3
- Add git-daemon and gitweb packages

* Thu Jun 21 2007 Josh Boyer <jwboyer@jdub.homelinux.org> 1.5.2.2-2
- Add emacs-git package (#235431)

* Mon Jun 18 2007 James Bowes <jbowes@redhat.com> 1.5.2.2-1
- git-1.5.2.2

* Fri Jun 08 2007 James Bowes <jbowes@redhat.com> 1.5.2.1-1
- git-1.5.2.1

* Tue May 13 2007 Quy Tonthat <qtonthat@gmail.com>
- Added lib files for git-gui
- Added Documentation/technical (As needed by Git Users Manual)

* Tue May 8 2007 Quy Tonthat <qtonthat@gmail.com>
- Added howto files

* Fri Mar 30 2007 Chris Wright <chrisw@redhat.com> 1.5.0.6-1
- git-1.5.0.6

* Mon Mar 19 2007 Chris Wright <chrisw@redhat.com> 1.5.0.5-1
- git-1.5.0.5

* Tue Mar 13 2007 Chris Wright <chrisw@redhat.com> 1.5.0.3-1
- git-1.5.0.3

* Fri Mar 2 2007 Chris Wright <chrisw@redhat.com> 1.5.0.2-2
- BuildRequires perl-devel as of perl-5.8.8-14 (bz 230680)

* Mon Feb 26 2007 Chris Wright <chrisw@redhat.com> 1.5.0.2-1
- git-1.5.0.2

* Mon Feb 13 2007 Nicolas Pitre <nico@cam.org>
- Update core package description (Git isn't as stupid as it used to be)

* Mon Feb 12 2007 Junio C Hamano <junkio@cox.net>
- Add git-gui and git-citool.

* Sun Dec 10 2006 Chris Wright <chrisw@redhat.com> 1.4.4.2-2
- no need to install manpages executable (bz 216790)
- use bytes for git-cvsserver

* Sun Dec 10 2006 Chris Wright <chrisw@redhat.com> 1.4.4.2-1
- git-1.4.4.2

* Mon Nov 6 2006 Jindrich Novy <jnovy@redhat.com> 1.4.2.4-2
- rebuild against the new curl

* Tue Oct 17 2006 Chris Wright <chrisw@redhat.com> 1.4.2.4-1
- git-1.4.2.4

* Wed Oct 4 2006 Chris Wright <chrisw@redhat.com> 1.4.2.3-1
- git-1.4.2.3

* Fri Sep 22 2006 Chris Wright <chrisw@redhat.com> 1.4.2.1-1
- git-1.4.2.1

* Mon Sep 11 2006 Chris Wright <chrisw@redhat.com> 1.4.2-1
- git-1.4.2

* Thu Jul 6 2006 Chris Wright <chrisw@redhat.com> 1.4.1-1
- git-1.4.1

* Tue Jun 13 2006 Chris Wright <chrisw@redhat.com> 1.4.0-1
- git-1.4.0

* Thu May 4 2006 Chris Wright <chrisw@redhat.com> 1.3.3-1
- git-1.3.3
- enable git-email building, prereqs have been relaxed

* Thu May 4 2006 Chris Wright <chrisw@redhat.com> 1.3.2-1
- git-1.3.2

* Fri Apr 28 2006 Chris Wright <chrisw@redhat.com> 1.3.1-1
- git-1.3.1

* Wed Apr 19 2006 Chris Wright <chrisw@redhat.com> 1.3.0-1
- git-1.3.0

* Mon Apr 10 2006 Chris Wright <chrisw@redhat.com> 1.2.6-1
- git-1.2.6

* Wed Apr 5 2006 Chris Wright <chrisw@redhat.com> 1.2.5-1
- git-1.2.5

* Wed Mar 1 2006 Chris Wright <chrisw@redhat.com> 1.2.4-1
- git-1.2.4

* Wed Feb 22 2006 Chris Wright <chrisw@redhat.com> 1.2.3-1
- git-1.2.3

* Tue Feb 21 2006 Chris Wright <chrisw@redhat.com> 1.2.2-1
- git-1.2.2

* Thu Feb 16 2006 Chris Wright <chrisw@redhat.com> 1.2.1-1
- git-1.2.1

* Mon Feb 13 2006 Chris Wright <chrisw@redhat.com> 1.2.0-1
- git-1.2.0

* Tue Feb 1 2006 Chris Wright <chrisw@redhat.com> 1.1.6-1
- git-1.1.6

* Tue Jan 24 2006 Chris Wright <chrisw@redhat.com> 1.1.4-1
- git-1.1.4

* Sun Jan 15 2006 Chris Wright <chrisw@redhat.com> 1.1.2-1
- git-1.1.2

* Tue Jan 10 2006 Chris Wright <chrisw@redhat.com> 1.1.1-1
- git-1.1.1

* Tue Jan 10 2006 Chris Wright <chrisw@redhat.com> 1.1.0-1
- Update to latest git-1.1.0 (drop git-email for now)
- Now creates multiple packages:
-        git-core, git-svn, git-cvs, git-arch, gitk

* Mon Nov 14 2005 H. Peter Anvin <hpa@zytor.com> 0.99.9j-1
- Change subpackage names to git-<name> instead of git-core-<name>
- Create empty root package which brings in all subpackages
- Rename git-tk -> gitk

* Thu Nov 10 2005 Chris Wright <chrisw@osdl.org> 0.99.9g-1
- zlib dependency fix
- Minor cleanups from split
- Move arch import to separate package as well

* Tue Sep 27 2005 Jim Radford <radford@blackbean.org>
- Move programs with non-standard dependencies (svn, cvs, email)
  into separate packages

* Tue Sep 27 2005 H. Peter Anvin <hpa@zytor.com>
- parallelize build
- COPTS -> CFLAGS

* Fri Sep 16 2005 Chris Wright <chrisw@osdl.org> 0.99.6-1
- update to 0.99.6

* Fri Sep 16 2005 Horst H. von Brand <vonbrand@inf.utfsm.cl>
- Linus noticed that less is required, added to the dependencies

* Sun Sep 11 2005 Horst H. von Brand <vonbrand@inf.utfsm.cl>
- Updated dependencies
- Don't assume manpages are gzipped

* Thu Aug 18 2005 Chris Wright <chrisw@osdl.org> 0.99.4-4
- drop sh_utils, sh-utils, diffutils, mktemp, and openssl Requires
- use RPM_OPT_FLAGS in spec file, drop patch0

* Wed Aug 17 2005 Tom "spot" Callaway <tcallawa@redhat.com> 0.99.4-3
- use dist tag to differentiate between branches
- use rpm optflags by default (patch0)
- own %%{_datadir}/git-core/

* Mon Aug 15 2005 Chris Wright <chrisw@osdl.org>
- update spec file to fix Buildroot, Requires, and drop Vendor

* Sun Aug 07 2005 Horst H. von Brand <vonbrand@inf.utfsm.cl>
- Redid the description
- Cut overlong make line, loosened changelog a bit
- I think Junio (or perhaps OSDL?) should be vendor...

* Thu Jul 14 2005 Eric Biederman <ebiederm@xmission.com>
- Add the man pages, and the --without docs build option

* Wed Jul 7 2005 Chris Wright <chris@osdl.org>
- initial git spec file
