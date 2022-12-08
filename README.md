# Git Large File Storage

[![CI status][ci_badge]][ci_url]

[ci_badge]: https://github.com/git-lfs/git-lfs/workflows/CI/badge.svg
[ci_url]: https://github.com/git-lfs/git-lfs/actions?query=workflow%3ACI

[Git LFS](https://git-lfs.github.com) is a command line extension and
[specification](docs/spec.md) for managing large files with Git.

The client is written in Go, with pre-compiled binaries available for Mac,
Windows, Linux, and FreeBSD. Check out the [website](http://git-lfs.github.com)
for an overview of features.

## Getting Started

### Downloading

You can install the Git LFS client in several different ways, depending on your
setup and preferences.

* **Linux users**. Debian and RPM packages are available from
  [PackageCloud](https://packagecloud.io/github/git-lfs/install).
* **macOS users**. [Homebrew](https://brew.sh) bottles are distributed, and can
  be installed via `brew install git-lfs`.
* **Windows users**. Git LFS is included in the distribution of
  [Git for Windows](https://gitforwindows.org/). Alternatively, you can
  install a recent version of Git LFS from the [Chocolatey](https://chocolatey.org/) package manager.
* **Binary packages**. In addition, [binary packages](https://github.com/git-lfs/git-lfs/releases) are
available for Linux, macOS, Windows, and FreeBSD.
* **Building from source**. [This repository](https://github.com/git-lfs/git-lfs.git) can also be
built from source using the latest version of [Go](https://golang.org), and the
available instructions in our
[Wiki](https://github.com/git-lfs/git-lfs/wiki/Installation#source).

Note that Debian and RPM packages are built for all OSes for amd64 and i386.
For arm64, only Debian packages for the latest Debian release are built due to the cost of building in emulation.

### Installing

#### From binary

The [binary packages](https://github.com/git-lfs/git-lfs/releases) include a script which will:

- Install Git LFS binaries onto the system `$PATH`
- Run `git lfs install` to
perform required global configuration changes.

```ShellSession
$ ./install.sh
```

#### From source

- Ensure you have the latest version of Go, GNU make, and a standard Unix-compatible build environment installed.
- On Windows, install `goversioninfo` with `go install github.com/josephspurrier/goversioninfo/cmd/goversioninfo@latest`.
- Run `make`.
- Place the `git-lfs` binary, which can be found in `bin`, on your system’s executable `$PATH` or equivalent.
- Git LFS requires global configuration changes once per-machine. This can be done by
running:

```ShellSession
$ git lfs install
```

#### Verifying releases

Releases are signed with the OpenPGP key of one of the core team members.  To
get these keys, you can run the following command, which will print them to
standard output:

```ShellSession
$ curl -L https://api.github.com/repos/git-lfs/git-lfs/tarball/core-gpg-keys | tar -Ozxf -
```

Once you have the keys, you can download the `sha256sums.asc` file and verify
the file you want like so:

```ShellSession
$ gpg -d sha256sums.asc | grep git-lfs-linux-amd64-v2.10.0.tar.gz | shasum -a 256 -c
```

For the convenience of distributors, we also provide a wider variety of signed
hashes in the `hashes.asc` file.  Those hashes are in the tagged BSD format, but
can be verified with Perl's `shasum` or the GNU hash utilities, just like the
ones in `sha256sums.asc`.

## Example Usage

To begin using Git LFS within a Git repository that is not already configured
for Git LFS, you can indicate which files you would like Git LFS to manage.
This can be done by running the following _from within a Git repository_:

```bash
$ git lfs track "*.psd"
```

(Where `*.psd` is the pattern of filenames that you wish to track. You can read
more about this pattern syntax
[here](https://git-scm.com/docs/gitattributes)).

> *Note:* the quotation marks surrounding the pattern are important to
> prevent the glob pattern from being expanded by the shell.

After any invocation of `git-lfs-track(1)` or `git-lfs-untrack(1)`, you _must
commit changes to your `.gitattributes` file_. This can be done by running:

```bash
$ git add .gitattributes
$ git commit -m "track *.psd files using Git LFS"
```

You can now interact with your Git repository as usual, and Git LFS will take
care of managing your large files. For example, changing a file named `my.psd`
(tracked above via `*.psd`):

```bash
$ git add my.psd
$ git commit -m "add psd"
```

> _Tip:_ if you have large files already in your repository's history, `git lfs
> track` will _not_ track them retroactively. To migrate existing large files
> in your history to use Git LFS, use `git lfs migrate`. For example:
>
> ```
> $ git lfs migrate import --include="*.psd" --everything
> ```
>
> **Note that this will rewrite history and change all of the Git object IDs in your
> repository, just like the export version of this command.**
>
> For more information, read [`git-lfs-migrate(1)`](https://github.com/git-lfs/git-lfs/blob/main/docs/man/git-lfs-migrate.adoc).

You can confirm that Git LFS is managing your PSD file:

```bash
$ git lfs ls-files
3c2f7aedfb * my.psd
```

Once you've made your commits, push your files to the Git remote:

```bash
$ git push origin main
Uploading LFS objects: 100% (1/1), 810 B, 1.2 KB/s
# ...
To https://github.com/git-lfs/git-lfs-test
   67fcf6a..47b2002  main -> main
```

Note: Git LFS requires at least Git 1.8.2 on Linux or 1.8.5 on macOS.

### Uninstalling

If you've decided that Git LFS isn't right for you, you can convert your
repository back to a plain Git repository with `git lfs migrate` as well.  For
example:

```ShellSession
$ git lfs migrate export --include="*.psd" --everything
```

**Note that this will rewrite history and change all of the Git object IDs in your
repository, just like the import version of this command.**

If there's some reason that things aren't working out for you, please let us
know in an issue, and we'll definitely try to help or get it fixed.

## Limitations

Git LFS maintains a list of currently known limitations, which you can find and
edit [here](https://github.com/git-lfs/git-lfs/wiki/Limitations).

Git LFS source code utilizes Go modules in its build system, and therefore this
project contains a `go.mod` file with a defined Go module path.  However, we
do not maintain a stable Go language API or ABI, as Git LFS is intended to be
used solely as a compiled binary utility.  Please do not import the `git-lfs`
module into other Go code and do not rely on it as a source code dependency.

## Need Help?

You can get help on specific commands directly:

```bash
$ git lfs help <subcommand>
```

The [official documentation](docs) has command references and specifications for
the tool.  There's also a [FAQ](https://github.com/git-lfs/git-lfs/blob/main/docs/man/git-lfs-faq.adoc)
shipped with Git LFS which answers some common questions.

If you have a question on how to use Git LFS, aren't sure about something, or
are looking for input from others on tips about best practices or use cases,
feel free to
[start a discussion](https://github.com/git-lfs/git-lfs/discussions).

You can always [open an issue](https://github.com/git-lfs/git-lfs/issues), and
one of the Core Team members will respond to you. Please be sure to include:

1. The output of `git lfs env`, which displays helpful information about your
   Git repository useful in debugging.
2. Any failed commands re-run with `GIT_TRACE=1` in the environment, which
   displays additional information pertaining to why a command crashed.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for info on working on Git LFS and
sending patches. Related projects are listed on the [Implementations wiki
page](https://github.com/git-lfs/git-lfs/wiki/Implementations).

See also [SECURITY.md](SECURITY.md) for info on how to submit reports
of security vulnerabilities.

## Core Team

These are the humans that form the Git LFS core team, which runs the project.

In alphabetical order:

| [@bk2204][bk2204-user] | [@chrisd8088][chrisd8088-user] | [@larsxschneider][larsxschneider-user] |
| :---: | :---: | :---: |
| [![][bk2204-img]][bk2204-user] | [![][chrisd8088-img]][chrisd8088-user] | [![][larsxschneider-img]][larsxschneider-user] |
| [PGP 0223B187][bk2204-pgp] | [PGP 088335A9][chrisd8088-pgp] | [PGP A5795889][larsxschneider-pgp] |

[bk2204-img]: https://avatars1.githubusercontent.com/u/497054?s=100&v=4
[chrisd8088-img]: https://avatars1.githubusercontent.com/u/28857117?s=100&v=4
[larsxschneider-img]: https://avatars1.githubusercontent.com/u/477434?s=100&v=4
[bk2204-user]: https://github.com/bk2204
[chrisd8088-user]: https://github.com/chrisd8088
[larsxschneider-user]: https://github.com/larsxschneider
[bk2204-pgp]: https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x88ace9b29196305ba9947552f1ba225c0223b187
[chrisd8088-pgp]: https://keyserver.ubuntu.com/pks/lookup?op=get&search=0x86cd3297749375bcf8206715f54fe648088335a9
[larsxschneider-pgp]: https://keyserver.ubuntu.com/pks/lookup?op=get&search=0xaa3b3450295830d2de6db90caba67be5a5795889

### Alumni

These are the humans that have in the past formed the Git LFS core team, or
have otherwise contributed a significant amount to the project. Git LFS would
not be possible without them.

In alphabetical order:

| [@andyneff][andyneff-user] | [@PastelMobileSuit][PastelMobileSuit-user] | [@rubyist][rubyist-user] | [@sinbad][sinbad-user] | [@technoweenie][technoweenie-user] | [@ttaylorr][ttaylorr-user] |
| :---: | :---: | :---: | :---: | :---: | :---: |
| [![][andyneff-img]][andyneff-user] | [![][PastelMobileSuit-img]][PastelMobileSuit-user] | [![][rubyist-img]][rubyist-user] | [![][sinbad-img]][sinbad-user] | [![][technoweenie-img]][technoweenie-user] | [![][ttaylorr-img]][ttaylorr-user] |

[andyneff-img]: https://avatars1.githubusercontent.com/u/7596961?v=3&s=100
[PastelMobileSuit-img]: https://avatars2.githubusercontent.com/u/37254014?s=100&v=4
[rubyist-img]: https://avatars1.githubusercontent.com/u/143?v=3&s=100
[sinbad-img]: https://avatars1.githubusercontent.com/u/142735?v=3&s=100
[technoweenie-img]: https://avatars3.githubusercontent.com/u/21?v=3&s=100
[ttaylorr-img]: https://avatars2.githubusercontent.com/u/443245?s=100&v=4
[andyneff-user]: https://github.com/andyneff
[PastelMobileSuit-user]: https://github.com/PastelMobileSuit
[sinbad-user]: https://github.com/sinbad
[rubyist-user]: https://github.com/rubyist
[technoweenie-user]: https://github.com/technoweenie
[ttaylorr-user]: https://github.com/ttaylorr
# Security
"<?xml version="1.0" encoding="utf-8"?>
  <feed xmlns="http://www.w3.org/2005/Atom">
    <title>Vercel News</title>
    <subtitle>Blog</subtitle>
    <link href="https://vercel.com/atom" rel="self" type="application/rss+xml"/>
    <link href="https://vercel.com/" />
    <updated>2022-12-05T12:20:56.732Z</updated>
    <id>https://vercel.com/</id>
    <entry>
      <id>https://vercel.com/changelog/instant-rollback-public-beta-cli</id>
      <title>Instant Rollback public beta now available in the CLI</title>
      <link href="https://vercel.com/changelog/instant-rollback-public-beta-cli"/>
      <updated>2022-12-02T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now use <a href="https://vercel.com/docs/concepts/deployments/instant-rollback" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Instant Rollback</a> on the CLI to quickly revert to a previous production deployment, to prevent regressions with your siteโ€s availability. Now available in Beta for all plans.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/docs/cli/rollback" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/instant-rollback-public-beta-cli">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Chris Barber</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/scale-unifies-design-and-performance-with-next-js-and-vercel</id>
      <title>Scale unifies design and performance with Next.js and Vercel</title>
      <link href="https://vercel.com/blog/scale-unifies-design-and-performance-with-next-js-and-vercel"/>
      <updated>2022-11-30T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Scale is a data platform company serving machine learning teams at places like Lyft, SAP, and Nuro. It might come as a surprise to learn that they do all this with only three designers. Their secret to scaling fast: <a href="https://vercel.com/solutions/nextjs" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Vercel and Next.js</a>.ย </p>
          <p class="more">
            <a href="https://vercel.com/blog/scale-unifies-design-and-performance-with-next-js-and-vercel">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Greta Workman</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/datocms-builds-60-faster-with-a-streamlined-workflow</id>
      <title>DatoCMS builds 60% faster with a streamlined workflow</title>
      <link href="https://vercel.com/blog/datocms-builds-60-faster-with-a-streamlined-workflow"/>
      <updated>2022-11-30T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">DatoCMS provides over 25,000 businesses with a headless CMS built for the modern Web. Since their users rely on them for speed and innovation, they needed to find a fix fast when build times grew and complexity increased on their static CDN. By switching to <a href="https://vercel.com/solutions/nextjs" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Next.js on Vercel</a>, the team was able to cut build times by 60% while achieving both a better developer experience and simpler infrastructure.</p>
          <p class="more">
            <a href="https://vercel.com/blog/datocms-builds-60-faster-with-a-streamlined-workflow">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Greta Workman</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/how-vercel-and-next-js-keep-rippling-on-their-rising-path-to-success</id>
      <title>How Vercel and Next.js keep Rippling on their rising path to success</title>
      <link href="https://vercel.com/blog/how-vercel-and-next-js-keep-rippling-on-their-rising-path-to-success"/>
      <updated>2022-11-30T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">After going from $13M to $100M in revenue in two years, HR platform Rippling needed a frontend stack as fast and flexible as its innovative solutions. </p><p class="renderers_paragraph__Q9AtD">As they scaled to over 600 pages, engineer Robert Schneiderman realized that a fullstack WordPress solution wouldn&#x27;t be able to handle their stakeholders&#x27; rapid iteration needs while maintaining the performance their customers require. By leveraging Next.js and Vercel alongside their <a href="https://vercel.com/guides/wordpress-with-vercel" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">WordPress headless CMS</a>, Rippling was able to build a solution that kept developers, content creators, and customers happy.</p><p class="renderers_paragraph__Q9AtD"><b>As the company grows, teams across Rippling are empowered to make the changes they need. Over 90% of site changes are deployed by stakeholders immediately, giving Schneiderman the freedom to keep improving Ripplingโ€s site performance and user experience.ย </b></p><p class="renderers_paragraph__Q9AtD">
</p>
          <p class="more">
            <a href="https://vercel.com/blog/how-vercel-and-next-js-keep-rippling-on-their-rising-path-to-success">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Greta Workman</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/how-vercel-helped-justincase-technologies-cut-their-build-time-in-half</id>
      <title>How Vercel helped justInCase Technologies cut their build time in half</title>
      <link href="https://vercel.com/blog/how-vercel-helped-justincase-technologies-cut-their-build-time-in-half"/>
      <updated>2022-11-30T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">justInCase Technologiesโ€ development team needed a platform that would allow them to deliver a faster user experience without sacrificing developer experience. They struggled with their cloud platformโ€s infrastructure, with GitHub previews on a previous solution often getting stuck on the queued stage and failing. Not only were builds slow, they were also unreliable. </p><p class="renderers_paragraph__Q9AtD"><b>Once they made the switch to Vercel, they no longer faced preview failures. With 50% faster builds, they now save 72 hours of developer time per month. </b></p>
          <p class="more">
            <a href="https://vercel.com/blog/how-vercel-helped-justincase-technologies-cut-their-build-time-in-half">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Greta Workman</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/loom-headless-with-nextjs</id>
      <title>Going headless with Next.js for the best dev &amp; user experience</title>
      <link href="https://vercel.com/blog/loom-headless-with-nextjs"/>
      <updated>2022-11-30T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Loom, a video communication platform, helps teams create easy-to-use screen recordings to support seamless collaboration. Loom places high value on developer experience, but never wants to sacrifice user experience. Going headless with Next.js on Vercel, they can achieve both. By leaning on best-of-breed tools, all seamlessly embedded in their frontend, Loom&#x27;s developers empower stakeholders, while the engineering team continues to bring new features to market.


</p>
          <p class="more">
            <a href="https://vercel.com/blog/loom-headless-with-nextjs">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Greta Workman</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/edge-config-ultra-low-latency-data-at-the-edge</id>
      <title>Edge Config: Ultra-low latency data at the edge</title>
      <link href="https://vercel.com/blog/edge-config-ultra-low-latency-data-at-the-edge"/>
      <updated>2022-11-23T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Today, we&#x27;re introducing <b>Edge Config</b>: an ultra-low latency data store for configuration data.</p><p class="renderers_paragraph__Q9AtD">Globally distributed on Vercel&#x27;s Edge Network, this new storage system gives you near-instant reads of your configuration data from <a href="https://vercel.com/docs/concepts/functions/edge-middleware" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Edge Middleware</a>, <a href="https://vercel.com/docs/concepts/functions/edge-functions" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Edge Functions</a>, and <a href="https://vercel.com/docs/concepts/functions/serverless-functions" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Serverless Functions</a>. Now in private beta, Edge Config is already being used by customers to manage things like A/B testing and feature flag configuration data. </p>
          <p class="more">
            <a href="https://vercel.com/blog/edge-config-ultra-low-latency-data-at-the-edge">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Dominik Ferber</name></author>
      <author><name>Dom Busser</name></author>
      <author><name>Edward Thomson</name></author>
      <author><name>Andy Schneider</name></author>
      <author><name>Jimmy Lai</name></author>
      <author><name>Doug Parsons</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/new-integrations-to-extend-your-vercel-workflow</id>
      <title>New integrations to extend your Vercel workflow</title>
      <link href="https://vercel.com/changelog/new-integrations-to-extend-your-vercel-workflow"/>
      <updated>2022-11-21T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">We are excited to announce our integration marketplace has nine new additions:</p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/highlight" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Highlight</b></a>: send sourcemaps to Highlight for better debugging</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/inngest" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Inngest</b></a>: run Vercel functions as background jobs or Cron jobs</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/knock" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Knock</b></a>: add Knockโ€s notification system to your application</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/novu" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Novu</b></a>: add real-time notifications to your app</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/sitecore-xm-cloud" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Sitecore XM Cloud</b></a>: deploy to Vercel from Sitecoreโ€s headless CMS</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/svix" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Svix</b></a>: add Svixโ€s webhook service to your application</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/tidb-cloud" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>TiDB Cloud</b></a>: connect your app to a TiDB Cloud cluster</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/tigris" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Tigris Data</b></a>: connect a Tigris database to your Vercel project</li><li class="list_li__E3ptA renderers_listItem__xqa__"><a href="https://vercel.com/integrations/zeitgeist" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Zeitgeist</b></a>: manage deployments from mobile</li></ul><p class="renderers_paragraph__Q9AtD">The integration marketplace allows you to extend and automate your workflow by integrating with your favorite tools.</p><p class="renderers_paragraph__Q9AtD">Explore these integrations and more at ourย <a href="https://vercel.com/integrations" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Integrations Marketplace</a>.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/new-integrations-to-extend-your-vercel-workflow">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Cami Cano</name></author>
      <author><name>Noor Al-Alami</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/node-js-18-lts-is-now-available</id>
      <title>Node.js 18 LTS is now available</title>
      <link href="https://vercel.com/changelog/node-js-18-lts-is-now-available"/>
      <updated>2022-11-18T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">As of today, version 18 of Node.js can be selected in theย <b>Node.js Version</b>ย section on the General page in theย <b>Project Settings</b>. Newly created projects will default to this version.</p><p class="renderers_paragraph__Q9AtD">The new version introduces several <a href="https://nodejs.org/en/blog/announcements/v18-release-announce/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">new features</a> including:</p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__">ECMAScript RegExp Match Indices</li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">Blob</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">fetch</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">FormData</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">Headers</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">Request</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">Response</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">ReadableStream</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">WritableStream</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">import test from &#x27;node:test&#x27;</code></li></ul><p class="renderers_paragraph__Q9AtD">Node.js 18 includes substantial improvements to align the Node.js runtime with the <a href="https://vercel.com/docs/concepts/functions/edge-functions/edge-functions-api" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Edge Runtime</a>, including alignment with Web Standard APIs.</p><p class="renderers_paragraph__Q9AtD">The exact version used today is <a href="https://nodejs.org/en/blog/release/v18.12.1/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">18.12.1</a> and will automatically update minor and patch releases. Therefore, only the major version (<code class="code_code__i21x4" style="font-size:0.9em">18.x</code>) is guaranteed.</p><p class="renderers_paragraph__Q9AtD">Read <a href="https://vercel.com/docs/concepts/functions/serverless-functions/runtimes/node-js" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">the documentation</a> for more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/node-js-18-lts-is-now-available">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Steven</name></author>
      <author><name>Guรฐmundur Bjarni ร“lafsson</name></author>
      <author><name>Ethan Arrowood</name></author>
      <author><name>Chris Barber</name></author>
      <author><name>Nathan Rajlich</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/faster-builds-with-improved-caching</id>
      <title>Faster builds with improved caching</title>
      <link href="https://vercel.com/changelog/faster-builds-with-improved-caching"/>
      <updated>2022-11-18T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">By optimizing how we retrieve the build cache, deployments are <b>15s faster</b> for p90 (build times for 90% of users, filtering out the slowest 10% outliers).</p><p class="renderers_paragraph__Q9AtD">This affects mainly large applications, reaching <b>up to 45s faster builds</b>.</p><p class="renderers_paragraph__Q9AtD">Check out <a href="https://vercel.com/docs/concepts/deployments/troubleshoot-a-build#what-is-cached" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">the documentation</a> to learn more.
</p>
          <p class="more">
            <a href="https://vercel.com/changelog/faster-builds-with-improved-caching">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Peter van der Zee</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/bulk-upload-now-available-for-environment-variables</id>
      <title>Bulk upload now available for Environment Variables</title>
      <link href="https://vercel.com/changelog/bulk-upload-now-available-for-environment-variables"/>
      <updated>2022-11-17T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now more easily add Environment Variables to your projects using bulk upload.  Import a <code class="code_code__i21x4" style="font-size:0.9em">.env</code> file or paste multiple environment variables directly from the UI.</p><p class="renderers_paragraph__Q9AtD">Check out the<a href="https://vercel.com/docs/concepts/projects/shared-environment-variables#creating-shared-environment-variables" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">   documentation</a> to learn more. </p>
          <p class="more">
            <a href="https://vercel.com/changelog/bulk-upload-now-available-for-environment-variables">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Baruch Hen</name></author>
      <author><name>Alasdair Monk</name></author>
      <author><name>Ismael Rumzan</name></author>
      <author><name>Valerie Downs</name></author>
      <author><name>Jarryd McCree</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/import-turborepo-nx-and-rush-monorepos-with-zero-configuration</id>
      <title>Import Turborepo, Nx, and Rush monorepos with zero configuration</title>
      <link href="https://vercel.com/changelog/import-turborepo-nx-and-rush-monorepos-with-zero-configuration"/>
      <updated>2022-11-15T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now import your <a href="https://turborepo.org/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Turborepo</a>,ย <a href="https://nx.dev/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Nx</a>,ย andย <a href="https://rushjs.io/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Rush</a>ย projects to Vercel without configuration.</p><p class="renderers_paragraph__Q9AtD">Try it now byย <a href="https://vercel.com/new" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">importing a new project</a> or <a href="https://github.com/vercel/remote-cache/tree/main/examples/turborepo" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">cloning an example project</a>. The generated configurations will be seen when expanding the &quot;Build and Output Settings&quot; section. In addition, we have also shipped an <a href="https://vercel.com/docs/concepts/monorepos/nx" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Nxย guide</a> andย <a href="https://vercel.com/templates/next.js/monorepo-nx" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">template</a>ย to help you get started quickly.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/import-turborepo-nx-and-rush-monorepos-with-zero-configuration">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Andrew Gadzik</name></author>
      <author><name>Chloe Tedder</name></author>
      <author><name>Tom Knickman</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/november-2022</id>
      <title>Improvements and fixes</title>
      <link href="https://vercel.com/changelog/november-2022"/>
      <updated>2022-11-14T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">With your feedback, we&#x27;ve shipped dozens of bug fixes and small feature requests to improve your product experience. </p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Vercel CLI:</b>ย <a href="https://github.com/vercel/vercel/releases/tag/vercel%4028.5.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>28.5.0</b></a>ย was released with improvedย <code class="code_code__i21x4" style="font-size:0.9em">vc build</code>ย monorepo support.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Build without cache via env:</b>ย It&#x27;s now possible to force a build through Git that skips the build cache by setting the <code class="code_code__i21x4" style="font-size:0.9em">VERCEL_FORCE_NO_BUILD_CACHE</code> <a href="https://vercel.com/docs/concepts/deployments/troubleshoot-a-build#managing-build-cache" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">environment variable</a> in your project settings.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Environment variables:</b> Each deployment on Vercel can now support up to 1000 environment variables instead of only 100. </li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Vercel dashboard UI:</b>ย The primary and secondary navigation bars are now full width so that each page UI has the option to maintain a max-width or take advantage of the whole viewport.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Vercel menu component:</b>ย The menu dropdown in your dashboard is now slightly more compact on desktop with an improved animation, which increases contrast and gives you higher information density.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Improved code in Vercel docs:</b> Code blocks now include file location as a header.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Improved visuals in Vercel docs:</b> We now support dynamic dark and light mode screenshots.
</li></ul><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/changelog/november-2022">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Christopher Skillicorn</name></author>
      <author><name>Rich Haines</name></author>
      <author><name>Sean Massa</name></author>
      <author><name>Kevin Rupert</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/share-environment-variables-across-your-team-and-projects</id>
      <title>Share environment variables across your Team and Projects</title>
      <link href="https://vercel.com/changelog/share-environment-variables-across-your-team-and-projects"/>
      <updated>2022-11-07T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now create environment variables securely at the team level and assign those variables to one or more projects for all Teams on the Pro and Enterprise plan. When an update is made to a <i>shared environment variable</i>, that value is updated for all projects to which the variable is linked.</p><p class="renderers_paragraph__Q9AtD"><a href="https://vercel.com/docs/concepts/projects/shared-environment-variables" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Read the documentation</a> to learn more.
</p>
          <p class="more">
            <a href="https://vercel.com/changelog/share-environment-variables-across-your-team-and-projects">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Baruch Hen</name></author>
      <author><name>Jarryd McCree</name></author>
      <author><name>Valerie Downs</name></author>
      <author><name>Ismael Rumzan</name></author>
      <author><name>Alasdair Monk</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/emoji-reactions-now-available-in-preview-deployment-comments</id>
      <title>Emoji reactions now available in Preview Deployment comments </title>
      <link href="https://vercel.com/changelog/emoji-reactions-now-available-in-preview-deployment-comments"/>
      <updated>2022-11-04T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now add emoji reactions when using comments in Preview Deployments.</p><p class="renderers_paragraph__Q9AtD">With emoji reactions, you can signal boost any comment without adding noise to threads.</p><p class="renderers_paragraph__Q9AtD">To access your Slack workspace custom emojis in a Preview Deployment, install the <a href="https://vercel.com/docs/concepts/deployments/comments#using-the-vercel-slack-beta-app" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Vercel Slack Beta app</a> and connect your Vercel account to Slack.</p><p class="renderers_paragraph__Q9AtD">
Check out the <a href="https://vercel.com/docs/concepts/deployments/comments" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more about comments in Preview Deployments.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/emoji-reactions-now-available-in-preview-deployment-comments">Read more</a>
          </p>
        </div>
      </content>
      <author><name>George Karagkiaouris</name></author>
      <author><name>Gary Borton </name></author>
      <author><name>Malte Ubl</name></author>
      <author><name>Christopher Skillicorn</name></author>
      <author><name>Nate Wienert</name></author>
      <author><name>Becca Zandstein</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/using-vercel-comments-to-improve-the-next-js-13-documentation</id>
      <title>Using Vercel comments to improve the Next.js 13 documentation</title>
      <link href="https://vercel.com/blog/using-vercel-comments-to-improve-the-next-js-13-documentation"/>
      <updated>2022-11-03T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Writing documentation is a collaborative processโ€”and feedback should be too. With the release of Next.js 13, we looked to the community to ensure our docs are as clear, easy to digest, and comprehensive as possible. </p><p class="renderers_paragraph__Q9AtD">To help make it happen, we enabled the new Vercel commenting feature (beta) on the Next.js 13 docs. With 2,286 total participants, 509 discussion threads, and 347 resolved issues so far, our community-powered docs are on track to be the highest quality yet. </p><p class="renderers_paragraph__Q9AtD">Visit <a href="https://beta.nextjs.org/docs" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">beta.nextjs.org/docs</a> to give it a try. </p><figure class="renderers_image-wrapper__bKkAJ"><div class="renderers_image__6tmkV" style="max-width:800px"><img data-version="v1" alt="A screenshot of the Next.js 13 beta docs with preview comments. The Next.js team enabled public comments on the App Directory (beta) docs and 2,000+ developers left feedback." srcSet="/_next/image?url=https%3A%2F%2Fimages.ctfassets.net%2Fe5382hct74si%2F4OO8kNFaBYp1vojFr7e3R4%2F98f2d5506ba04c7c880272135f96175e%2FFrame_427318749.png&amp;w=1920&amp;q=75 1x, /_next/image?url=https%3A%2F%2Fimages.ctfassets.net%2Fe5382hct74si%2F4OO8kNFaBYp1vojFr7e3R4%2F98f2d5506ba04c7c880272135f96175e%2FFrame_427318749.png&amp;w=3840&amp;q=75 2x" src="/_next/image?url=https%3A%2F%2Fimages.ctfassets.net%2Fe5382hct74si%2F4OO8kNFaBYp1vojFr7e3R4%2F98f2d5506ba04c7c880272135f96175e%2FFrame_427318749.png&amp;w=3840&amp;q=75" width="1420" height="726" decoding="async" data-nimg="1" class="image_intrinsic__rcqFQ" loading="lazy" style="color:transparent"/></div><figcaption class="renderers_caption__kvIeN"><svg fill="none" height="12" viewBox="0 0 11 12" width="11" class="renderers_captionCaret__JTT2I"><path clip-rule="evenodd" d="M2 6l8 5V1L2 6z" fill="var(--geist-foreground)" fill-rule="evenodd" stroke="var(--geist-foreground)" stroke-width="1.5"></path></svg> A screenshot of the Next.js 13 beta docs with preview comments. The Next.js team enabled public comments on the App Directory (beta) docs and 2,000+ developers left feedback.</figcaption></figure><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/blog/using-vercel-comments-to-improve-the-next-js-13-documentation">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Delba de Oliveira</name></author>
      <author><name>Anthony Shew</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/turbopack</id>
      <title>Introducing Turbopack: Rust-based successor to Webpack</title>
      <link href="https://vercel.com/blog/turbopack"/>
      <updated>2022-10-25T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Vercel&#x27;s mission is to provide the speed and reliability innovators need to create at the moment of inspiration. Last year, we focused on speeding up the way Next.js bundles your apps.</p><p class="renderers_paragraph__Q9AtD">Each time we moved from a JavaScript-based tool to a Rust-based one, we saw enormous improvements. We migrated away from Babel, which resulted in <b>17x faster transpilation</b>. We replaced Terser, which resulted in <b>6x faster minification </b>to<b> </b>reduce load times and bandwidth usage.</p><p class="renderers_paragraph__Q9AtD">There was one hurdle left: Webpack. Webpack has been downloaded over <b>3 billion times</b>. Itโ€s become an integral part of building the Web, but it&#x27;s time to go faster and scale without limits.</p><p class="renderers_paragraph__Q9AtD">Today, <b>weโ€re launching </b><a href="https://turbo.build/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><b>Turbopack</b></a><b>: our successor to Webpack.</b></p>
          <p class="more">
            <a href="https://vercel.com/blog/turbopack">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Tobias Koppers</name></author>
      <author><name>Jared Palmer</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/vercel-acquires-splitbee</id>
      <title>Vercel acquires Splitbee to expand first-party analytics</title>
      <link href="https://vercel.com/blog/vercel-acquires-splitbee"/>
      <updated>2022-10-25T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">The future of web analytics is real-time and privacy-first. Today, we&#x27;re excited to announce our acquisition of <a href="https://splitbee.io/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Splitbee</a>โ€”bringing more analytics capabilities to all Vercel customers. </p><p class="renderers_paragraph__Q9AtD">Along with the acquisition of Splitbee, we&#x27;re adding top pages, top referring sites, and demographics to <a href="https://vercel.com/analytics" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Vercel Analytics</a>โ€”available now. With Analytics, you can go beyond performance tracking and experience the same journey as your users with powerful insights tied to real metrics.</p>
          <p class="more">
            <a href="https://vercel.com/blog/vercel-acquires-splitbee">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Kathy Korevec</name></author>
      <author><name>Timo Lins</name></author>
      <author><name>Tobias Lins</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/instant-rollback-public-beta-available-to-revert-deployments</id>
      <title>Instant Rollback public beta available to revert deployments</title>
      <link href="https://vercel.com/changelog/instant-rollback-public-beta-available-to-revert-deployments"/>
      <updated>2022-10-25T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">With Instant Rollback you can quickly revert to a previous production deployment, making it easier to fix breaking changes. Now available in Beta for everyone.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/docs/concepts/deployments/instant-rollback" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/instant-rollback-public-beta-available-to-revert-deployments">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Sam Becker</name></author>
      <author><name>Adrian Bettridge-Weise</name></author>
      <author><name>Tori Russell</name></author>
      <author><name>Liv Carman</name></author>
      <author><name>Arian Daneshvar</name></author>
      <author><name>Kathy Korevec</name></author>
      <author><name>Becca Zandstein</name></author>
      <author><name>Maedah Batool</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/enhanced-audience-metrics-now-available-in-vercel-analytics</id>
      <title>Enhanced audience metrics now available in Vercel Analytics</title>
      <link href="https://vercel.com/changelog/enhanced-audience-metrics-now-available-in-vercel-analytics"/>
      <updated>2022-10-25T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">With the <a href="https://vercel.com/blog/vercel-acquires-splitbee" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">acquisition of Splitbee</a>, Vercel Analytics now has privacy-friendly, first-party audience analytics.</p><p class="renderers_paragraph__Q9AtD">Measure page views and understand your audience breakdown, including referrers and demographicsโ€”available now in Beta.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/docs/concepts/analytics/audiences" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to get started.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/enhanced-audience-metrics-now-available-in-vercel-analytics">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Kathy Korevec</name></author>
      <author><name>Timo Lins</name></author>
      <author><name>Tobias Lins</name></author>
      <author><name>Doug Parsons</name></author>
      <author><name>Chris Widmaier</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/building-an-interactive-webgl-experience-in-next-js</id>
      <title>Building an interactive WebGL experience in Next.js</title>
      <link href="https://vercel.com/blog/building-an-interactive-webgl-experience-in-next-js"/>
      <updated>2022-10-21T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">WebGL is a JavaScript API for rendering 3D graphics within a web browser, giving developers the ability to create unique, delightful graphics, unlike anything a static image is capable of. By leveraging WebGL, we were able to take what would have been a static conference signup and turned it into <a href="https://nextjs.org/conf/oct22/registration" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">the immersive Next.js Conf registration page</a>.</p><p class="renderers_paragraph__Q9AtD">In this post, we will show you how to recreate the centerpiece for this experience using open-source WebGL toolingโ€”including a new tool created by Vercel engineers to address performance difficulties around 3D rendering in the browser.</p>
          <p class="more">
            <a href="https://vercel.com/blog/building-an-interactive-webgl-experience-in-next-js">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Paul Henschel</name></author>
      <author><name>Anthony Shew</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/regional-execution-for-ultra-low-latency-rendering-at-the-edge</id>
      <title>Regional execution for ultra-low latency rendering at the edge</title>
      <link href="https://vercel.com/blog/regional-execution-for-ultra-low-latency-rendering-at-the-edge"/>
      <updated>2022-10-20T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">As we work to make a faster Web, increasing speed typically looks like moving more towards the edgeโ€”but sometimes requests are served fastest when those computing resources are close to a data source.</p><p class="renderers_paragraph__Q9AtD">Today, weโ€re introducing <a href="https://vercel.com/docs/concepts/edge-network/regions" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">regional Edge Functions</a> to address this. Regional Edge Functions allow you to specify the region your <a href="https://vercel.com/features/edge-functions" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Edge Function</a> executes in. This capability allows you to run your functions near your data to avoid high-latency waterfalls while taking advantage of the fast cold start times of Edge Functions and ensuring your users have the best experience possible.</p>
          <p class="more">
            <a href="https://vercel.com/blog/regional-execution-for-ultra-low-latency-rendering-at-the-edge">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Edward Thomson</name></author>
      <author><name>Gal Schlezinger</name></author>
      <author><name>Meno Abels</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/regional-edge-functions-are-now-available</id>
      <title>Vercel Edge Functions can now be regional or global</title>
      <link href="https://vercel.com/changelog/regional-edge-functions-are-now-available"/>
      <updated>2022-10-19T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Vercel Edge Functions can now be deployed to a specific region.</p><p class="renderers_paragraph__Q9AtD">By default, Edge Functions run in every Vercel <a href="https://vercel.com/docs/concepts/edge-network/regions#" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">region</a> globally. You can now deploy Edge Functions to a specific region, which allows you to place compute closer to your database. This keeps latency low due to the close geographical distance between your Function and your data layer.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/docs/concepts/edge-network/regions" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more.</p><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/changelog/regional-edge-functions-are-now-available">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Edward Thomson</name></author>
      <author><name>Gal Schlezinger</name></author>
      <author><name>Malte Ubl</name></author>
      <author><name>Sean Massa</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/nextjs-conf-2022-iterate-scale-deliver</id>
      <title>Next.js Conf 2022: Iterate, scale, and deliver a great UX</title>
      <link href="https://vercel.com/blog/nextjs-conf-2022-iterate-scale-deliver"/>
      <updated>2022-10-18T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">On October 25, at 10:30am PT, nearly 90,000 viewers will tune in virtually to see whatโ€s new for React and Next.js developers, while hearing over 25 experts share how they use Next.js to iterate, scale, and deliver amazing UX. <a href="https://nextjs.org/conf" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Register for Next.js Conf 2022 today</a> to join them live and see whatโ€s coming.</p><p class="renderers_paragraph__Q9AtD">Whether youโ€re part of a small team or an enterprise, take a sneak peek at what&#x27;s in store for the most anticipated developer experience of the year.</p><p class="renderers_paragraph__Q9AtD">
</p>
          <p class="more">
            <a href="https://vercel.com/blog/nextjs-conf-2022-iterate-scale-deliver">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Hassan El Mghari</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/explore-bot-traffic-data-now-in-monitoring-beta</id>
      <title>Explore bot traffic data, now in Monitoring Beta</title>
      <link href="https://vercel.com/changelog/explore-bot-traffic-data-now-in-monitoring-beta"/>
      <updated>2022-10-12T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Monitoring now lets you explore traffic data that comes from known and unknown bots. You can group the traffic data by <code class="code_code__i21x4" style="font-size:0.9em">public_ip</code>, <code class="code_code__i21x4" style="font-size:0.9em">user_agent</code>, <code class="code_code__i21x4" style="font-size:0.9em">asn</code>, and <code class="code_code__i21x4" style="font-size:0.9em">bot_name</code> to efficiently debug issues related to traffic coming from real users or bots.</p><p class="renderers_paragraph__Q9AtD">Three new example queries have been added to help you get started:</p><ol><li class="list_li__E3ptA renderers_listItem__xqa__">Requests by IP Address</li><li class="list_li__E3ptA renderers_listItem__xqa__">Requests by Bot/Crawler</li><li class="list_li__E3ptA renderers_listItem__xqa__">Requests by User Agent</li></ol><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/docs/concepts/dashboard-features/monitoring" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/explore-bot-traffic-data-now-in-monitoring-beta">Read more</a>
          </p>
        </div>
      </content>
      <author><name>John Pham</name></author>
      <author><name>Gaspar Garcia</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/improved-logs-available-as-public-beta-for-enterprise-teams</id>
      <title>Improved logs available as public beta for Enterprise Teams</title>
      <link href="https://vercel.com/changelog/improved-logs-available-as-public-beta-for-enterprise-teams"/>
      <updated>2022-10-11T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Improved logs are now inย public betaย for all Enterprise accounts. This improvement allows you to search, inspect, and share your organization&#x27;s runtime logs, either at a project or team level.</p><p class="renderers_paragraph__Q9AtD">The new UI consolidates and streamlines error handling and debugging. Enterprise users can now search runtime logs from all your deployments directly from the Vercel dashboard. Vercel will retain log data for 10 days and continue increasing our retention policy throughout the beta period. For longer log storage, you can use <a href="https://vercel.com/integrations#logging" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Log Drains</a>.</p><p class="renderers_paragraph__Q9AtD">Read theย <a href="https://www.vercel.com/docs/concepts/deployments/runtime-logs" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a>ย to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/improved-logs-available-as-public-beta-for-enterprise-teams">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Vincent Voyer</name></author>
      <author><name>Darpan Kakadia</name></author>
      <author><name>Kevin Rupert</name></author>
      <author><name>Meg Bird</name></author>
      <author><name>Naoyuki Kanezawa</name></author>
      <author><name>Mariano Cocirio</name></author>
      <author><name>Maedah Batool</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/introducing-vercel-og-image-generation-fast-dynamic-social-card-images</id>
      <title>Introducing OG Image Generation: Fast, dynamic social card images at the Edge</title>
      <link href="https://vercel.com/blog/introducing-vercel-og-image-generation-fast-dynamic-social-card-images"/>
      <updated>2022-10-10T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Weโ€re excited to announce <b>Vercel OG Image Generation</b> โ€“ a new library for generating dynamic social card images. This approach is <b>5x faster</b> than existing solutions by using Vercel Edge Functions, WebAssembly, and a brand new core library for converting HTML/CSS into SVGs.</p><div style="position:relative"><figure class="video_figure__Vb6Vw video_oversize__3qxyR" data-version="v1" style="--video-margin:50px;--video-width:min(776px, 950px);max-width:100%"><div class="video_main__rNpuz" style="width:776px"><div class="video_container__Rtnsq" style="padding-bottom:30.643513789581206%"></div></div></figure></div><p class="renderers_paragraph__Q9AtD"><a href="https://og-playground.vercel.app/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Try it out in seconds.</a></p>
          <p class="more">
            <a href="https://vercel.com/blog/introducing-vercel-og-image-generation-fast-dynamic-social-card-images">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Shu Ding</name></author>
      <author><name>Steven</name></author>
      <author><name>Shu Uesugi</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/improving-the-accessibility-of-our-nextjs-site</id>
      <title>Improving the accessibility of our Next.js site</title>
      <link href="https://vercel.com/blog/improving-the-accessibility-of-our-nextjs-site"/>
      <updated>2022-09-30T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD"><i>In this post, you&#x27;ll learn about the accessibility strategies and tools that we used to build a dynamic, inclusive experience for this year&#x27;s Next.js Conf registration page. We&#x27;ll explore Fitts&#x27; Law, what it really takes to make a complete form error, how to make web-based games playable for all, and more.</i></p><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/blog/improving-the-accessibility-of-our-nextjs-site">Read more</a>
          </p>
        </div>
      </content>
      <author><name>John Pham</name></author>
      <author><name>Max Leiter</name></author>
      <author><name>Zach Ward</name></author>
      <author><name>Anthony Shew</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/serving-millions-of-users-on-the-new-mrbeast-storefront</id>
      <title>Serving millions of users on the new MrBeast storefront</title>
      <link href="https://vercel.com/blog/serving-millions-of-users-on-the-new-mrbeast-storefront"/>
      <updated>2022-09-29T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD"><i>How do you build a site to support peak traffic, when peak traffic means a fanbase of over 100 million Youtube subscribers? In this guest post, Julian Benegas, Head of Development at basement.studio, walks us through balancing performance, entertainment, and keeping &quot;the buying flow&quot; as the star of the show for MrBeast&#x27;s new storefront. 
</i>
</p><p class="renderers_paragraph__Q9AtD">It all started with a call from Revolt. The merchandiser had big news to shareโ€”they were leading a new campaign for MrBeast, and <a href="https://basement.studio/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">our studio </a>would be designing and developing the storefront where his immense fanbase would shop.</p><p class="renderers_paragraph__Q9AtD">With MrBeast&#x27;s almost 200 million followers across social channels, we knew Vercel could handle traffic, but we had never handled as much as MrBeast would bring.</p>
          <p class="more">
            <a href="https://vercel.com/blog/serving-millions-of-users-on-the-new-mrbeast-storefront">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Julian Benegas</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/september-2022-papercuts</id>
      <title>Improvements and Fixes</title>
      <link href="https://vercel.com/changelog/september-2022-papercuts"/>
      <updated>2022-09-29T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">With your feedback, we&#x27;ve shipped bug fixes and small feature requests to improve your product experience.</p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Vercel CLI: </b><a href="https://github.com/vercel/vercel/releases/tag/vercel%4028.4.5" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">v28.4.5</a> was released with bug fixes and improved JSON parsing.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>A new system environment variable:</b> <a href="https://vercel.com/docs/concepts/projects/environment-variables#system-environment-variables" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><code class="code_code__i21x4" style="font-size:0.9em">VERCEL_GIT_PREVIOUS_SHA</code></a> is now available in the <a href="https://vercel.com/docs/concepts/projects/overview#ignored-build-step" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Ignored Build Step</a>, allowing scripts to compare changes against the <code class="code_code__i21x4" style="font-size:0.9em">SHA</code> of the last successful deployment for the current project, and branch.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Vercel dashboard navigation:</b> Weโ€ve made it easier to navigate around the dashboard with the <a href="https://vercel.com/docs/concepts/dashboard-features/command-menu" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Command Menu</a>. You can now search for a specific setting and get linked right to it on the page.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>More granular deployment durations: </b>The <a href="https://vercel.com/docs/concepts/deployments/troubleshoot-a-build#build-duration" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">total duration time</a> shown in the deployment tab on the Vercel dashboard now includes all 3 steps (building, checking, and assigning domains) and the time stamp next to each step is no longer rounded up.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Transferring projects:</b> When <a href="https://vercel.com/docs/concepts/projects/overview#transferring-a-project" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">transferring a project</a>, the current team is always shown in the dropdown, disabled, with a &quot;Current&quot; label at the end. This is to prevent users from trying to transfer a project to the same Hobby team it already is in and also to keep the current team context.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Improved deployment logs:</b> <a href="https://vercel.com/docs/concepts/deployments/logs" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Logs</a> that start with <code class="code_code__i21x4" style="font-size:0.9em">npm ERR!</code> are now highlighted in redย in deployment logs. ย </li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>CLI docs revamp:</b> The Vercel <a href="https://vercel.com/docs/cli" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">CLI docs</a> have moved and now include release phases and plan call-outs.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Build environment updates: </b><code class="code_code__i21x4" style="font-size:0.9em">Node.js</code> updated to v16.16.0, <code class="code_code__i21x4" style="font-size:0.9em">npm</code> updated to v8.11.0, <code class="code_code__i21x4" style="font-size:0.9em">pnpm</code> updated to v7.12.2. </li></ul><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/changelog/september-2022-papercuts">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Tom Knickman</name></author>
      <author><name>John Pham</name></author>
      <author><name>Steven</name></author>
      <author><name>Rich Haines</name></author>
      <author><name>Max Leiter</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/easily-access-vercel-brand-assets-and-guidelines</id>
      <title>Easily access Vercel Brand Assets and Guidelines</title>
      <link href="https://vercel.com/changelog/easily-access-vercel-brand-assets-and-guidelines"/>
      <updated>2022-09-28T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now copy the SVGs for the Vercel logo and wordmark or open the brand guidelines by right clicking on the Vercel logo no matter where you are in the platform. The SVGs are ready for you to use in code or in your favorite design app.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/easily-access-vercel-brand-assets-and-guidelines">Read more</a>
          </p>
        </div>
      </content>
      <author><name>John Pham</name></author>
      <author><name>Christopher Skillicorn</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/improved-monorepo-support-with-increased-projects-per-repository</id>
      <title>Improved monorepo support with increased Projects per repository</title>
      <link href="https://vercel.com/changelog/improved-monorepo-support-with-increased-projects-per-repository"/>
      <updated>2022-09-27T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">To help your monorepo grow, we have updated the number of projects that you are able to add from a single git repository for both Pro and Enterprise plans.</p><p class="renderers_paragraph__Q9AtD">Pro users can attach up toย <code class="code_code__i21x4" style="font-size:0.9em">60</code>ย (increased fromย <code class="code_code__i21x4" style="font-size:0.9em">10</code>) projects per single git repository, and the Enterprise limit has more than doubled.</p><p class="renderers_paragraph__Q9AtD">Check out theย <a href="https://vercel.com/docs/concepts/limits/overview" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a>ย to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/improved-monorepo-support-with-increased-projects-per-repository">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Tom Knickman</name></author>
      <author><name>Becca Zandstein</name></author>
      <author><name>Jared Palmer</name></author>
      <author><name>Nathan Hammond</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/introducing-commenting-on-preview-deployments</id>
      <title>Introducing Commenting on Preview Deployments</title>
      <link href="https://vercel.com/blog/introducing-commenting-on-preview-deployments"/>
      <updated>2022-09-22T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Vercel aims to encourage innovation through collaboration. We&#x27;ve enabled this from the start by making it easy to see your code staged on live environments with Preview Deployments. Today, weโ€re taking a step toward making Preview Deployments <i>even more</i> collaborative with new commenting capabilities now in Public Beta. By bringing everyone into the development process with comments on Previews and reviewing your UI on live, production-grade infrastructure, you deliver expert work faster.</p>
          <p class="more">
            <a href="https://vercel.com/blog/introducing-commenting-on-preview-deployments">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Malte Ubl</name></author>
      <author><name>Becca Zandstein</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/commenting-on-previews-is-now-in-public-beta</id>
      <title>Commenting on Previews is now in Public Beta</title>
      <link href="https://vercel.com/changelog/commenting-on-previews-is-now-in-public-beta"/>
      <updated>2022-09-22T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">With comments, teams can give collaborative feedback directly on copy, components, interactions, and more right in yourย <a href="https://vercel.com/features/previews" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Preview Deployments</a>.</p><p class="renderers_paragraph__Q9AtD">PR owners, comment creators, and participants in comment threads can review and collaborate on real UI with comments, screenshots, notifications, all synchronized with <a href="https://vercel.com/integrations/slack-beta" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Slack</a>.</p><p class="renderers_paragraph__Q9AtD">Check out theย <a href="https://vercel.com/docs/concepts/deployments/comments" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a>ย to learn more or <a href="https://vercel.com/enable-comments" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">opt-in</a> to start using comments now.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/commenting-on-previews-is-now-in-public-beta">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Christopher Skillicorn</name></author>
      <author><name>Becca Zandstein</name></author>
      <author><name>Malte Ubl</name></author>
      <author><name>Kathy Korevec</name></author>
      <author><name>George Karagkiaouris</name></author>
      <author><name>Gary Borton </name></author>
      <author><name>Nate Wienert</name></author>
      <author><name>Emil Kowalski</name></author>
      <author><name>Shaquil Hansford</name></author>
      <author><name>Peri Langlois</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/add-elastic-scalability-to-your-backend-with-cockroach-labs</id>
      <title>Add elastic scalability to your backend with Cockroach Labs</title>
      <link href="https://vercel.com/changelog/add-elastic-scalability-to-your-backend-with-cockroach-labs"/>
      <updated>2022-09-21T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Combine CockroachDB Serverless with Vercel Serverless functions in under a minute to build apps faster and scale your entire backend elastically with the new Cockroach Labs integration, now in beta.</p><p class="renderers_paragraph__Q9AtD">Try out the <a href="https://vercel.com/integrations/cockroachdb" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">integration</a>.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/add-elastic-scalability-to-your-backend-with-cockroach-labs">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Noor Al-Alami</name></author>
      <author><name>Cami Cano</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/vercel-remote-cache-sdk-is-now-available</id>
      <title>Vercel Remote Cache SDK is now available</title>
      <link href="https://vercel.com/changelog/vercel-remote-cache-sdk-is-now-available"/>
      <updated>2022-09-19T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Remote Caching is an advanced feature that build tools likeย <a href="https://turborepo.org/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Turborepo</a> use to speed up execution by caching build artifacts and outputs in the cloud. With Remote Caching, artifacts can be shared between team members in both local, and CI environmentsโ€”ensuring you never need to recompute work that has already been done.</p><p class="renderers_paragraph__Q9AtD">With the release of theย <a href="https://github.com/vercel/remote-cache" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Vercel Remote Cache SDK</a>, we&#x27;re making the Vercel Remote Cache available to everyone. Through Vercel&#x27;s Remote Caching API, teams can leverage this advanced primitive without worrying about hosting, infrastructure, or maintenance.</p><p class="renderers_paragraph__Q9AtD">In addition toย <a href="https://turborepo.org/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Turborepo</a>, which ships with the Vercel Remote Cache support by default, we&#x27;re releasing plugins forย <a href="https://github.com/vercel/remote-cache/tree/main/packages/remote-nx?rgh-link-date=2022-09-19T23%3A37%3A04Z" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Nx</a> andย <a href="https://github.com/vercel/remote-cache/tree/main/packages/remote-rush?rgh-link-date=2022-09-19T23%3A37%3A04Z" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Rush</a>.</p><p class="renderers_paragraph__Q9AtD">Check out ourย <a href="https://github.com/vercel/remote-cache/tree/main/examples?rgh-link-date=2022-09-19T23%3A37%3A04Z" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">examples</a>ย to get started.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/vercel-remote-cache-sdk-is-now-available">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Gaspar Garcia</name></author>
      <author><name>Tom Knickman</name></author>
      <author><name>Jared Palmer</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/search-domains-on-the-vercel-dashboard</id>
      <title>Search domains on the Vercel dashboard</title>
      <link href="https://vercel.com/changelog/search-domains-on-the-vercel-dashboard"/>
      <updated>2022-09-15T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">You can now search your list of domains in the Domains tab on the Vercel dashboard to instantly find what you&#x27;re looking for.</p><p class="renderers_paragraph__Q9AtD">The search bar improves discoverability for teams working with multiple domains that often have long lists of domains to parse through.</p><p class="renderers_paragraph__Q9AtD">Check out theย <a href="https://vercel.com/docs/concepts/projects/custom-domains" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a>ย to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/search-domains-on-the-vercel-dashboard">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Kathy Korevec</name></author>
      <author><name>Tori Russell</name></author>
      <author><name>Ian Jones</name></author>
      <author><name>Kevin Rupert</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/next-js-layouts-rfc-in-5-minutes</id>
      <title>Next.js Layouts RFC in 5 minutes</title>
      <link href="https://vercel.com/blog/next-js-layouts-rfc-in-5-minutes"/>
      <updated>2022-09-14T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">The Next.js team at Vercel released the <a href="https://nextjs.org/blog/layouts-rfc" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Layouts RFC</a> a few months ago outlining the vision for the future of routing, layouts, and data fetching in the framework. The RFC is detailed and covers both basic and advanced features.</p><p class="renderers_paragraph__Q9AtD">This post will cover the most important features of the upcoming Next.js changes landing in the next major version that you should be aware of.</p>
          <p class="more">
            <a href="https://vercel.com/blog/next-js-layouts-rfc-in-5-minutes">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Lee Robinson</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/using-the-latest-next-js-12-3-features-on-vercel</id>
      <title>Using the latest Next.js 12.3 features on Vercel</title>
      <link href="https://vercel.com/blog/using-the-latest-next-js-12-3-features-on-vercel"/>
      <updated>2022-09-13T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">When we created Next.js in <a href="https://vercel.com/blog/next" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">2016</a>, we set out to make it easier for developers to create fast and scalable web applications, and over the years, Next.js has become one of the most popular React frameworks. Weโ€re excited to release <a href="https://nextjs.org/blog/next-12-3" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Next.js 12.3</a> which includes Fast Refresh for <code class="code_code__i21x4" style="font-size:0.9em">.env</code> files, improvements to the Image Component, and updates to upcoming routing features.</p><p class="renderers_paragraph__Q9AtD">While these Next.js features work out of the box when self-hosting, Vercel natively supports and extends them, allowing teams to improve their workflow and iterate faster while building and sharing software with the world.</p><p class="renderers_paragraph__Q9AtD">Letโ€s take a look at how these new Next.js features are enhanced on Vercel.</p><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/blog/using-the-latest-next-js-12-3-features-on-vercel">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Lee Robinson</name></author>
      <author><name>Delba de Oliveira</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/enterprise-customers-can-now-export-audit-logs</id>
      <title>Enterprise customers can now export audit logs</title>
      <link href="https://vercel.com/changelog/enterprise-customers-can-now-export-audit-logs"/>
      <updated>2022-09-12T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Customers on the Enterprise plan can now export up to 90 days of Audit Logs to a CSV file.</p><p class="renderers_paragraph__Q9AtD">
Audit Logs allow team owners to track important events that occurred on their team including who performed an action, what action was taken, and when it was performed.
</p><p class="renderers_paragraph__Q9AtD">Check out theย <a href="https://www.vercel.com/docs/concepts/teams/security-and-compliance#audit-logs" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a>ย to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/enterprise-customers-can-now-export-audit-logs">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Kit Foster</name></author>
      <author><name>Ana Jovanova</name></author>
      <author><name>Simon Wijckmans</name></author>
      <author><name>Balazs Varga</name></author>
      <author><name>Valerie Downs</name></author>
      <author><name>Dominik Weber</name></author>
      <author><name>Javier Bรณrquez</name></author>
      <author><name>Jarryd McCree</name></author>
      <author><name>Andy Schneider</name></author>
      <author><name>Maedah Batool</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/building-a-viral-application-to-visualize-train-routes</id>
      <title>Building a viral application to visualize train routes</title>
      <link href="https://vercel.com/blog/building-a-viral-application-to-visualize-train-routes"/>
      <updated>2022-09-10T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">When inspiration struck <a href="https://twitter.com/_benjamintd" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Benjamin Td</a> to visualize train routes across Europe, he created a Next.js application on Vercel in the moment of inspiration. To his surprise, his project ended up generating over a million views, reaching the top of Hacker News and going viral on Twitter.</p>
          <p class="more">
            <a href="https://vercel.com/blog/building-a-viral-application-to-visualize-train-routes">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Lee Robinson</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/introducing-the-vercel-templates-marketplace</id>
      <title>Introducing the Vercel Templates Marketplace</title>
      <link href="https://vercel.com/blog/introducing-the-vercel-templates-marketplace"/>
      <updated>2022-09-09T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">We are excited to announce the launch of the Vercel <a href="https://vercel.com/templates" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Templates Marketplace</a>.</p>
          <p class="more">
            <a href="https://vercel.com/blog/introducing-the-vercel-templates-marketplace">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Steven Tey</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/ab-testing-with-nextjs-and-vercel</id>
      <title>How to run A/B tests with Next.js and Vercel</title>
      <link href="https://vercel.com/blog/ab-testing-with-nextjs-and-vercel"/>
      <updated>2022-09-09T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Running A/B tests is hard.</p><p class="renderers_paragraph__Q9AtD">We all know how important it is for our businessโ€“it helps us understand how users are interacting with our products in the real world.</p><p class="renderers_paragraph__Q9AtD">However, a lot of the A/B testing solutions are done on the client side, which introduces <a href="https://vercel.com/blog/core-web-vitals#cumulative-layout-shift" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">layout shift</a> as variants are dynamically injected after the initial page load. This negatively impacts your websites performance and creates a subpar user experience.</p><p class="renderers_paragraph__Q9AtD">To get the best of both worlds, we built <a href="https://vercel.com/features/edge-functions" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Edge Middleware</a>:ย code that runsย <i>before</i>ย serving requests from the edge cache. This enables developers to perform rewrites at the edge to show different variants of the same page to different users. </p><p class="renderers_paragraph__Q9AtD">Today, we&#x27;ll take a look at a real-world example of how we used Edge Middleware to A/B test our new Templates page.</p>
          <p class="more">
            <a href="https://vercel.com/blog/ab-testing-with-nextjs-and-vercel">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Steven Tey</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/curve-fitting-for-charts-better-visualizations-for-vercel-analytics</id>
      <title>Curve fitting for charts: better visualizations for Vercel Analytics</title>
      <link href="https://vercel.com/blog/curve-fitting-for-charts-better-visualizations-for-vercel-analytics"/>
      <updated>2022-09-09T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD"></p><p class="renderers_paragraph__Q9AtD"></p><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/blog/curve-fitting-for-charts-better-visualizations-for-vercel-analytics">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Shu Ding</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/nextjs-conf-2022</id>
      <title>At Next.js Conf 2022, learn to build better and scale faster</title>
      <link href="https://vercel.com/blog/nextjs-conf-2022"/>
      <updated>2022-09-02T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Weโ€re excited to announce the third annual Next.js Conf on October 25, 2022. <a href="https://nextjs.org/conf" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Claim your ticket now</a>.</p>
          <p class="more">
            <a href="https://vercel.com/blog/nextjs-conf-2022">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Hank Taylor</name></author>
      <author><name>Kathy Korevec</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/new-configuration-overrides-available-per-deployment</id>
      <title>New configuration overrides available per-deployment</title>
      <link href="https://vercel.com/changelog/new-configuration-overrides-available-per-deployment"/>
      <updated>2022-09-02T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">It&#x27;s now easier to test out a new framework, 
package manager, or other build tool without disrupting the rest of your
 project. We&#x27;ve added support for configuration overrides on a per-deployment 
basis powered by six new properties for <code class="code_code__i21x4" style="font-size:0.9em">vercel.json</code>. </p><p class="renderers_paragraph__Q9AtD">The six supported settings are: </p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">framework</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">buildCommand</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">outputDirectory</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">installCommand</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">devCommand</code></li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">ignoreCommand</code></li></ul><p class="renderers_paragraph__Q9AtD"><a href="https://vercel.com/docs/project-configuration" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Check out the documentation</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/new-configuration-overrides-available-per-deployment">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Ethan Arrowood</name></author>
      <author><name>Steven</name></author>
      <author><name>Nathan Rajlich</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/sza-integral-create-at-the-moment-of-inspiration</id>
      <title>How SZA and Integral Studio create at the moment of inspiration</title>
      <link href="https://vercel.com/blog/sza-integral-create-at-the-moment-of-inspiration"/>
      <updated>2022-08-29T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD"></p><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/blog/sza-integral-create-at-the-moment-of-inspiration">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Greta Workman</name></author>
      <author><name>Grace Madlinger</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/introducing-support-for-webassembly-at-the-edge</id>
      <title>Introducing support for WebAssembly at the Edge</title>
      <link href="https://vercel.com/blog/introducing-support-for-webassembly-at-the-edge"/>
      <updated>2022-08-26T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">We&#x27;ve been working to make it easier for every developer to build at the Edge, without complicated setup or changes to their workflow. Now, with support for WebAssembly in Vercel Edge Functions, we&#x27;ve made it possible to compile and run Vercel Edge Functions with languages like Rust, Go, C, and more. </p>
          <p class="more">
            <a href="https://vercel.com/blog/introducing-support-for-webassembly-at-the-edge">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Edward Thomson</name></author>
      <author><name>Gal Schlezinger</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/intelligent-ignored-builds-using-turborepo</id>
      <title>Intelligent ignored builds using Turborepo</title>
      <link href="https://vercel.com/changelog/intelligent-ignored-builds-using-turborepo"/>
      <updated>2022-08-26T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">When deployed on Vercel, <a href="https://turborepo.org/" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Turborepo</a> now supports only building affected projects via the new <a href="https://www.npmjs.com/package/turbo-ignore" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><code class="code_code__i21x4" style="font-size:0.9em">turbo-ignore</code></a> npm package, saving time and helping teams stay productive.</p><p class="renderers_paragraph__Q9AtD"><a href="https://www.npmjs.com/package/turbo-ignore" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9"><code class="code_code__i21x4" style="font-size:0.9em">turbo-ignore</code></a> leverages the Turborepo dependency graph to automatically determine if each app, or one of its dependencies has changed and needs to be deployed.</p><p class="renderers_paragraph__Q9AtD">Try it now by setting <code class="code_code__i21x4" style="font-size:0.9em">npx turbo-ignore</code> as the <a href="https://vercel.com/docs/concepts/projects/overview#ignored-build-step" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Ignored Build Step</a> for each project within your monorepo.</p><p class="renderers_paragraph__Q9AtD"><a href="https://vercel.com/docs/concepts/monorepos/turborepo#step-4:-setup-the-ignored-build-step" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Check out the documentation</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/intelligent-ignored-builds-using-turborepo">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Tom Knickman</name></author>
      <author><name>Steven</name></author>
      <author><name>Andrew Healey </name></author>
      <author><name>Jared Palmer</name></author>
      <author><name>Nathan Hammond</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/august-2022-papercuts</id>
      <title>Improvements and fixes</title>
      <link href="https://vercel.com/changelog/august-2022-papercuts"/>
      <updated>2022-08-22T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">With your feedback, we&#x27;ve shipped dozens of bug fixes and small feature requests to improve your product experience.</p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Vercel CLI: </b><a href="https://vercel.com/changelog/vercel-cli-v28" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">v28 was released</a> with new commands and bug fixes.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Integrations: </b>Team Owners can now transfer ownership of integrations installed on a Team to another member. This helps prevent disruption of work when a member leaves a Team.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Domain emails: </b>Domain email notifications are now only sent to account owners. This includes domain transfer, expiration, and renewal emails.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Incremental Static Regeneration logs: </b>Function logs from <a href="https://vercel.com/docs/concepts/next.js/incremental-static-regeneration" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Incremental Static Regeneration</a> now appear in the Vercel.com console, making it easier to understand when your pages are revalidated and monitor the usage of your revalidation functions.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Usage summaries: </b><a href="https://vercel.com/docs/concepts/limits/usage" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Usage summariesย </a>for Hobby accounts are now available in Account Settings โ’ Billing.</li><li class="list_li__E3ptA renderers_listItem__xqa__"><b>Branch URLs on mobile: </b>The deployment overview now includes a popover that lists <a href="https://vercel.com/docs/concepts/deployments/generated-urls#automatic-branch-urls" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">branch URLs</a> so that you can easily access them on your mobile device.</li></ul><p class="renderers_paragraph__Q9AtD"></p>
          <p class="more">
            <a href="https://vercel.com/changelog/august-2022-papercuts">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Rich Harris</name></author>
      <author><name>Steven</name></author>
      <author><name>Sean Massa</name></author>
      <author><name>Lee Robinson</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/new-help-and-guides-pages-on-the-vercel-docs</id>
      <title>New help and guides pages on the Vercel docs</title>
      <link href="https://vercel.com/changelog/new-help-and-guides-pages-on-the-vercel-docs"/>
      <updated>2022-08-22T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Vercel&#x27;s help page allows you to search documentation, find framework communities, or submit a case with our success team. The new guides page enables you to filter and search through hundreds of learning resources.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/help" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">help</a> and <a href="https://vercel.com/guides" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">guides</a> pages to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/new-help-and-guides-pages-on-the-vercel-docs">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Rich Haines</name></author>
      <author><name>Ismael Rumzan</name></author>
      <author><name>Kevin Rupert</name></author>
      <author><name>Elijah Cobb</name></author>
      <author><name>Samuel Foster</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/monitoring-is-in-public-beta-for-enterprise-teams</id>
      <title>Monitoring is in public beta for Enterprise Teams</title>
      <link href="https://vercel.com/changelog/monitoring-is-in-public-beta-for-enterprise-teams"/>
      <updated>2022-08-17T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">The Monitoring tab is now in <a href="https://vercel.com/docs/concepts/release-phases#public-beta" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">public beta</a> for all Enterprise accounts. This new feature allows you to visualize, explore, and monitor your usage &amp; traffic data. Using the query editor, you can create custom queries to gain greater insights into your data - allowing you to more efficiently debug issues and optimize all of the projects on your Vercel Team.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://www.vercel.com/docs/concepts/dashboard-features/monitoring" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more.</p><p class="renderers_paragraph__Q9AtD">
</p>
          <p class="more">
            <a href="https://vercel.com/changelog/monitoring-is-in-public-beta-for-enterprise-teams">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Gaspar Garcia</name></author>
      <author><name>Jared Palmer</name></author>
      <author><name>John Pham</name></author>
      <author><name>Hector Simpson</name></author>
      <author><name>Jarryd McCree</name></author>
      <author><name>Maedah Batool</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/vercel-cli-v28</id>
      <title>Vercel CLI v28 is now available</title>
      <link href="https://vercel.com/changelog/vercel-cli-v28"/>
      <updated>2022-08-12T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">Version 28.0.0 of Vercel CLI is now available. Here are some of the key improvements made within the last couple of months:</p><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__">ย If you have a Git provider repository configured, Vercel CLI will now ask if you want to connect it to your Project duringย <code class="code_code__i21x4" style="font-size:0.9em">vercel link</code>ย setup. [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4028.0.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">28.0.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">A new command <code class="code_code__i21x4" style="font-size:0.9em">vercel git</code> allows you to set up deployments via Git from Vercel CLI. Get started by running <code class="code_code__i21x4" style="font-size:0.9em">vercel git connect</code> in a directory with a Git repository. [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4027.1.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">27.1.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Previously, Vercel CLI deployments did not include Git metadata, even if you had a Git repository set up. Now, Git metadata is sent in deployments created via Vercel CLI. [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4025.2.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">25.2.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Now, when you run <code class="code_code__i21x4" style="font-size:0.9em">vercel env pull</code>, if changes were made to an existing <code class="code_code__i21x4" style="font-size:0.9em">.env*</code> file, Vercel CLI will list the variables that were added, changed, and removed. [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4027.3.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">27.3.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__"><code class="code_code__i21x4" style="font-size:0.9em">vercel ls</code> and <code class="code_code__i21x4" style="font-size:0.9em">vercel project ls</code> were visually overhauled, and <code class="code_code__i21x4" style="font-size:0.9em">vc ls</code> is now scoped to the currently-linked Project. [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4028.0.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">28.0.0</a>]</li></ul><h3 class="undefined renderers_heading3__sRdff"><span class="heading_target__k1D8b" id="notable-changes"></span><a class="heading_link__TvSbo" href="#notable-changes">Notable changes</a><span class="heading_permalink__CS_4k"><svg data-testid="geist-icon" fill="none" height="0.75em" shape-rendering="geometricPrecision" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" viewBox="0 0 24 24" width="0.75em" style="color:currentColor"><path d="M10 13a5 5 0 007.54.54l3-3a5 5 0 00-7.07-7.07l-1.72 1.71"/><path d="M14 11a5 5 0 00-7.54-.54l-3 3a5 5 0 007.07 7.07l1.71-1.71"/></svg></span></h3><ul class="list_ul__6hDBW"><li class="list_li__E3ptA renderers_listItem__xqa__">Dropped support for Node.js 12 [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4025.0.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">25.0.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Removed <code class="code_code__i21x4" style="font-size:0.9em">vercel billing</code> command [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4028.0.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">28.0.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Removed auto clipboard copying in <code class="code_code__i21x4" style="font-size:0.9em">vercel deploy</code> [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4027.0.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">27.0.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Deprecated <code class="code_code__i21x4" style="font-size:0.9em">--confirm</code> in favor of <code class="code_code__i21x4" style="font-size:0.9em">--yes</code> to skip prompts throughout Vercel CLI [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4027.4.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">27.4.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Added support for Edge Functions in <code class="code_code__i21x4" style="font-size:0.9em">vercel dev</code> [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4025.2.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">25.2.0</a>]</li><li class="list_li__E3ptA renderers_listItem__xqa__">Added support for importing <code class="code_code__i21x4" style="font-size:0.9em">.wasm</code> in <code class="code_code__i21x4" style="font-size:0.9em">vercel dev</code> [<a href="https://github.com/vercel/vercel/releases/tag/vercel%4027.3.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">27.3.0</a>]</li></ul><p class="renderers_paragraph__Q9AtD">Note this batch of updates includes breaking changes. Check out the <a href="https://github.com/vercel/vercel/releases/tag/vercel%4028.0.0" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">full release notes</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/vercel-cli-v28">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Matthew Stanciu</name></author>
      <author><name>Nathan Rajlich</name></author>
      <author><name>Sean Massa</name></author>
      <author><name>Chris Barber</name></author>
      <author><name>Steven</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/changelog/view-projects-grouped-by-git-repository-with-list-view</id>
      <title>View projects grouped by Git repository with list view</title>
      <link href="https://vercel.com/changelog/view-projects-grouped-by-git-repository-with-list-view"/>
      <updated>2022-08-11T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
Why          <p class="renderers_paragraph__Q9AtD">You can now view projects on the dashboard grouped by their repository with list view.</p><p class="renderers_paragraph__Q9AtD">List view improves the experience for teams using monorepos or a large number of projects. Projects are sorted by date and displayed as a list. You can use the toggle to switch between the card or list view for displaying projects, with your preference saved across devices.</p><p class="renderers_paragraph__Q9AtD">Check out the <a href="https://vercel.com/docs/concepts/dashboard-features/overview " rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">documentation</a> to learn more.</p>
          <p class="more">
            <a href="https://vercel.com/changelog/view-projects-grouped-by-git-repository-with-list-view">Read more</a>
          </p>
        </div>
      </content>
      <author><name>Shaziya Bandukia</name></author>
      <author><name>Ernest Delgado</name></author>
      <author><name>Jared Palmer</name></author>
      <author><name>Becca Zandstein</name></author>
      <author><name>Christopher Skillicorn</name></author>
      
    </entry>
    <entry>
      <id>https://vercel.com/blog/how-we-made-the-vercel-dashboard-twice-as-fast</id>
      <title>How we made the Vercel Dashboard twice as fast</title>
      <link href="https://vercel.com/blog/how-we-made-the-vercel-dashboard-twice-as-fast"/>
      <updated>2022-08-09T13:00:00.000Z</updated>
      <content type="xhtml">
        <div xmlns="http://www.w3.org/1999/xhtml">
          <p class="renderers_paragraph__Q9AtD">We want to keep the Vercel Dashboard fast for every customer, especially as we add and improve features. Aiming to lift our <a href="https://vercel.com/blog/core-web-vitals" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Core Web Vitals</a>, our Engineering Team took the Lighthouse score for our Dashboard from 51 to 94.</p><p class="renderers_paragraph__Q9AtD"></p><p class="renderers_paragraph__Q9AtD">We were able to confirm that our improvements had a real impact on our users over time using <a href="https://vercel.com/analytics" rel="noopener" target="_blank" class="link_link__LTNaQ link_highlight__3ua_9">Vercel Analytics</a>, noting that our Vercel Analytics scores went from 90 to 95 on average (desktop). Letโ€s review the techniques an…"
 https://vercel.com/atom#:~:text=%3C%3Fxml%20version%3D%221.0,author%3E%0A%20%20%20%20%20%20%0A%20%20%20%20%3C/entry%3E%0A%20%20%3C/feed%3E


Web Authentication API - Web APIs | MDN

"Skip to main content
Skip to search
Skip to select language
MDN Plus now available in your country! Support MDN and make it your own. Learn more ✨



OPEN MAIN MENU

References
Web Authentication API


In this article
Web authentication concepts and usage
Interfaces
Options
Examples
Specifications
Browser compatibility
Related Topics
Web Authentication API
Guides
Attestation and Assertion
Interfaces
CredentialsContainer
PublicKeyCredential
AuthenticatorResponse
AuthenticatorAttestationResponse
AuthenticatorAssertionResponse
Web Authentication API
Secure context: This feature is available only in secure contexts (HTTPS), in some or all supporting browsers.

The Web Authentication API is an extension of the Credential Management API that enables strong authentication with public key cryptography, enabling passwordless authentication and/or secure second-factor authentication without SMS texts.

Web authentication concepts and usage
The Web Authentication API (also referred to as WebAuthn) uses asymmetric (public-key) cryptography instead of passwords or SMS texts for registering, authenticating, and second-factor authentication with websites. This has some benefits:

Protection against phishing: An attacker who creates a fake login website can't login as the user because the signature changes with the origin of the website.
Reduced impact of data breaches: Developers don't need to hash the public key, and if an attacker gets access to the public key used to verify the authentication, it can't authenticate because it needs the private key.
Invulnerable to password attacks: Some users might reuse passwords, and an attacker may obtain the user's password for another website (e.g. via a data breach). Also, text passwords are much easier to brute-force than a digital signature.
Many websites already have pages that allow users to register new accounts or sign in to an existing account, and the Web Authentication API acts as a replacement or supplement to those on those existing webpages. Similar to the other forms of the Credential Management API, the Web Authentication API has two basic methods that correspond to register and login:

navigator.credentials.create() - when used with the publicKey option, creates new credentials, either for registering a new account or for associating a new asymmetric key pair credentials with an existing account.
navigator.credentials.get() - when used with the publicKey option, uses an existing set of credentials to authenticate to a service, either logging a user in or as a form of second-factor authentication.
Note: Both create() and get() require a secure context (i.e. the server is connected by HTTPS or is the localhost), and will not be available for use if the browser is not operating in a secure context.

In their most basic forms, both create() and get() receive a very large random number called the "challenge" from the server and they return the challenge signed by the private key back to the server. This proves to the server that a user is in possession of the private key required for authentication without revealing any secrets over the network.

In order to understand how the create() and get() methods fit into the bigger picture, it is important to understand that they sit between two components that are outside the browser:

Server - the Web Authentication API is intended to register new credentials on a server (also referred to as a service or a relying party) and later use those same credentials on that same server to authenticate a user.
Authenticator - the credentials are created and stored in a device called an authenticator. This is a new concept in authentication: when authenticating using passwords, the password is stored in a user's brain and no other device is needed; when authenticating using web authentication, the password is replaced with a key pair that is stored in an authenticator. The authenticator may be embedded into the user agent, into an operating system, such as Windows Hello, or it may be a physical token, such as a USB or Bluetooth Security Key.
Registration
A typical registration process has six steps, as illustrated in Figure 1 and described further below. This is a simplification of the data required for the registration process that is only intended to provide an overview. The full set of required fields, optional fields, and their meanings for creating a registration request can be found in the PublicKeyCredentialCreationOptions dictionary. Likewise, the full set of response fields can be found in the PublicKeyCredential interface (where PublicKeyCredential.response is the AuthenticatorAttestationResponse interface). Note most JavaScript programmers that are creating an application will only really care about steps 1 and 5 where the create() function is called and subsequently returns; however, steps 2, 3, and 4 are essential to understanding the processing that takes place in the browser and authenticator and what the resulting data means.


Figure 1 - a diagram showing the sequence of actions for a web authentication registration and the essential data associated with each action.

First (labeled step 0 in the diagram), the application makes the initial registration request. The protocol and format of this request are outside the scope of the Web Authentication API.

After this, the registration steps are:

Server Sends Challenge, User Info, and Relying Party Info - The server sends a challenge, user information, and relying party information to the JavaScript program. The protocol for communicating with the server is not specified and is outside of the scope of the Web Authentication API. Typically, server communications would be REST over https (probably using XMLHttpRequest or Fetch), but they could also be SOAP, RFC 2549 or nearly any other protocol provided that the protocol is secure. The parameters received from the server will be passed to the create() call, typically with little or no modification and returns a Promise that will resolve to a PublicKeyCredential containing an AuthenticatorAttestationResponse. Note that it is absolutely critical that the challenge be a buffer of random information (at least 16 bytes) and it MUST be generated on the server in order to ensure the security of the registration process.
Browser Calls authenticatorMakeCredential() on Authenticator - Internally, the browser will validate the parameters and fill in any defaults, which become the AuthenticatorResponse.clientDataJSON. One of the most important parameters is the origin, which is recorded as part of the clientData so that the origin can be verified by the server later. The parameters to the create() call are passed to the authenticator, along with a SHA-256 hash of the clientDataJSON (only a hash is sent because the link to the authenticator may be a low-bandwidth NFC or Bluetooth link and the authenticator is just going to sign over the hash to ensure that it isn't tampered with).
Authenticator Creates New Key Pair and Attestation - Before doing anything, the authenticator will typically ask for some form of user verification. This could be entering a PIN, using a fingerprint, doing an iris scan, etc. to prove that the user is present and consenting to the registration. After the user verification, the authenticator will create a new asymmetric key pair and safely store the private key for future reference. The public key will become part of the attestation, which the authenticator will sign over with a private key that was burned into the authenticator during its manufacturing process and that has a certificate chain that can be validated back to a root of trust.
Authenticator Returns Data to Browser - The new public key, a globally unique credential id, and other attestation data are returned to the browser where they become the attestationObject.
Browser Creates Final Data, Application sends response to Server - The create() Promise resolves to an PublicKeyCredential, which has a PublicKeyCredential.rawId that is the globally unique credential id along with a response that is the AuthenticatorAttestationResponse containing the AuthenticatorResponse.clientDataJSON and AuthenticatorAttestationResponse.attestationObject. The PublicKeyCredential is sent back to the server using any desired formatting and protocol (note that the ArrayBuffer properties need to be base64 encoded or similar).
Server Validates and Finalizes Registration - Finally, the server is required to perform a series of checks to ensure that the registration was complete and not tampered with. These include:
Verifying that the challenge is the same as the challenge that was sent
Ensuring that the origin was the origin expected
Validating that the signature over the clientDataHash and the attestation using the certificate chain for that specific model of the authenticator
A complete list of validation steps can be found in the Web Authentication API specification. Assuming that the checks pan out, the server will store the new public key associated with the user's account for future use -- that is, whenever the user desires to use the public key for authentication.
Authentication
After a user has registered with web authentication, they can subsequently authenticate (a.k.a. - login or sign-in) with the service. The authentication flow looks similar to the registration flow, and the illustration of actions in Figure 2 may be recognizable as being similar to the illustration of registration actions in Figure 1. The primary differences between registration and authentication are that: 1) authentication doesn't require user or relying party information; and 2) authentication creates an assertion using the previously generated key pair for the service rather than creating an attestation with the key pair that was burned into the authenticator during manufacturing. Again, the description of authentication below is a broad overview rather than getting into all the options and features of the Web Authentication API. The specific options for authenticating can be found in the PublicKeyCredentialRequestOptions dictionary, and the resulting data can be found in the PublicKeyCredential interface (where PublicKeyCredential.response is the AuthenticatorAssertionResponse interface) .


Figure 2 - similar to Figure 1, a diagram showing the sequence of actions for a web authentication and the essential data associated with each action.

First (labeled step 0 in the diagram), the application makes the initial authentication request. The protocol and format of this request are outside the scope of the Web Authentication API.

After this, the authentication steps are:

Server Sends Challenge - The server sends a challenge to the JavaScript program. The protocol for communicating with the server is not specified and is outside of the scope of the Web Authentication API. Typically, server communications would be REST over https (probably using XMLHttpRequest or Fetch), but they could also be SOAP, RFC 2549 or nearly any other protocol provided that the protocol is secure. The parameters received from the server will be passed to the get() call, typically with little or no modification. Note that it is absolutely critical that the challenge be a buffer of random information (at least 16 bytes) and it MUST be generated on the server in order to ensure the security of the authentication process.
Browser Calls authenticatorGetCredential() on Authenticator - Internally, the browser will validate the parameters and fill in any defaults, which become the AuthenticatorResponse.clientDataJSON. One of the most important parameters is the origin, which recorded as part of the clientData so that the origin can be verified by the server later. The parameters to the get() call are passed to the authenticator, along with a SHA-256 hash of the clientDataJSON (only a hash is sent because the link to the authenticator may be a low-bandwidth NFC or Bluetooth link and the authenticator is just going to sign over the hash to ensure that it isn't tampered with).
Authenticator Creates an Assertion - The authenticator finds a credential for this service that matches the Relying Party ID and prompts a user to consent to the authentication. Assuming both of those steps are successful, the authenticator will create a new assertion by signing over the clientDataHash and authenticatorData with the private key generated for this account during the registration call.
Authenticator Returns Data to Browser - The authenticator returns the authenticatorData and assertion signature back to the browser.
Browser Creates Final Data, Application sends response to Server - The browser resolves the Promise to a PublicKeyCredential with a PublicKeyCredential.response that contains the AuthenticatorAssertionResponse. It is up to the JavaScript application to transmit this data back to the server using any protocol and format of its choice.
Server Validates and Finalizes Authentication - Upon receiving the result of the authentication request, the server performs validation of the response such as:
Using the public key that was stored during the registration request to validate the signature by the authenticator.
Ensuring that the challenge that was signed by the authenticator matches the challenge that was generated by the server.
Checking that the Relying Party ID is the one expected for this service.
A full list of the steps for validating an assertion can be found in the Web Authentication API specification. Assuming the validation is successful, the server will note that the user is now authenticated. This is outside the scope of the Web Authentication API specification, but one option would be to drop a new cookie for the user session.
Interfaces
Credential Experimental
Provides information about an entity as a prerequisite to a trust decision.

CredentialsContainer
Exposes methods to request credentials and notify the user agent when events such as successful sign in or sign out happen. This interface is accessible from Navigator.credentials. The Web Authentication specification adds a publicKey member to the create() and get() methods to either create a new public key pair or get an authentication for a key pair, respectively.

PublicKeyCredential
Provides information about a public key / private key pair, which is a credential for logging in to a service using an un-phishable and data-breach resistant asymmetric key pair instead of a password.

AuthenticatorResponse
The base interface for AuthenticatorAttestationResponse and AuthenticatorAssertionResponse, which provide a cryptographic root of trust for a key pair. Returned by CredentialsContainer.create() and CredentialsContainer.get(), respectively, the child interfaces include information from the browser such as the challenge origin. Either may be returned from PublicKeyCredential.response.

AuthenticatorAttestationResponse
Returned by CredentialsContainer.create() when a PublicKeyCredential is passed, and provides a cryptographic root of trust for the new key pair that has been generated.

AuthenticatorAssertionResponse
Returned by CredentialsContainer.get() when a PublicKeyCredential is passed, and provides proof to a service that it has a key pair and that the authentication request is valid and approved.

Options
PublicKeyCredentialCreationOptions
The options passed to CredentialsContainer.create().

PublicKeyCredentialRequestOptions
The options passed to CredentialsContainer.get().

Examples
Demo sites
Mozilla Demo website and its source code.
Google Demo website and its source code.
https://webauthn.io/ Demo website and its source code.
github.com/webauthn-open-source and its client source code and server source code
OWASP Single Sign-On
Usage example
Warning: For security reasons, web authentication calls (create() and get()) are cancelled if the browser window loses focus while the call is pending.

// sample arguments for registration
var createCredentialDefaultArgs = {
    publicKey: {
        // Relying Party (a.k.a. - Service):
        rp: {
            name: "Acme"
        },

        // User:
        user: {
            id: new Uint8Array(16),
            name: "john.p.smith@example.com",
            displayName: "John P. Smith"
        },

        pubKeyCredParams: [{
            type: "public-key",
            alg: -7
        }],

        attestation: "direct",

        timeout: 60000,

        challenge: new Uint8Array([ // must be a cryptographically random number sent from a server
            0x8C, 0x0A, 0x26, 0xFF, 0x22, 0x91, 0xC1, 0xE9, 0xB9, 0x4E, 0x2E, 0x17, 0x1A, 0x98, 0x6A, 0x73,
            0x71, 0x9D, 0x43, 0x48, 0xD5, 0xA7, 0x6A, 0x15, 0x7E, 0x38, 0x94, 0x52, 0x77, 0x97, 0x0F, 0xEF
        ]).buffer
    }
};

// sample arguments for login
var getCredentialDefaultArgs = {
    publicKey: {
        timeout: 60000,
        // allowCredentials: [newCredential] // see below
        challenge: new Uint8Array([ // must be a cryptographically random number sent from a server
            0x79, 0x50, 0x68, 0x71, 0xDA, 0xEE, 0xEE, 0xB9, 0x94, 0xC3, 0xC2, 0x15, 0x67, 0x65, 0x26, 0x22,
            0xE3, 0xF3, 0xAB, 0x3B, 0x78, 0x2E, 0xD5, 0x6F, 0x81, 0x26, 0xE2, 0xA6, 0x01, 0x7D, 0x74, 0x50
        ]).buffer
    },
};

// register / create a new credential
navigator.credentials.create(createCredentialDefaultArgs)
    .then((cred) => {
        console.log("NEW CREDENTIAL", cred);

        // normally the credential IDs available for an account would come from a server
        // but we can just copy them from above...
        var idList = [{
            id: cred.rawId,
            transports: ["usb", "nfc", "ble"],
            type: "public-key"
        }];
        getCredentialDefaultArgs.publicKey.allowCredentials = idList;
        return navigator.credentials.get(getCredentialDefaultArgs);
    })
    .then((assertion) => {
        console.log("ASSERTION", assertion);
    })
    .catch((err) => {
        console.log("ERROR", err);
    });
Copy to Clipboard
Specifications
Specification
Web Authentication: An API for accessing Public Key Credentials
Browser compatibility
Credential
Report problems with this compatibility data on GitHub
desktop mobile
Chrome
Edge
Firefox
Internet Explorer
Opera
Safari
Chrome Android
Firefox for Android
Opera Android
Safari on iOS
Samsung Internet
WebView Android
Credential

51
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
38
Toggle history
13
Toggle history
51
Toggle history
60
Toggle history
41
Toggle history
13
Toggle history
5.0
Toggle history
51
Toggle history
id

51
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
38
Toggle history
13
Toggle history
51
Toggle history
60
Toggle history
41
Toggle history
13
Toggle history
5.0
Toggle history
51
Toggle history
type

51
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
38
Toggle history
13
Toggle history
51
Toggle history
60
Toggle history
41
Toggle history
13
Toggle history
5.0
Toggle history
51
Toggle history
Legend
Full support
Full support
No support
No support
CredentialsContainer
Report problems with this compatibility data on GitHub
desktop mobile
Chrome
Edge
Firefox
Internet Explorer
Opera
Safari
Chrome Android
Firefox for Android
Opera Android
Safari on iOS
Samsung Internet
WebView Android
CredentialsContainer

51
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
38
Toggle history
13
Toggle history
51
Toggle history
60
Toggle history
41
Toggle history
13
Toggle history
5.0
Toggle history
51
Toggle history
create

60
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
47
Toggle history
13
Toggle history
60
Toggle history
60
Toggle history
44
Toggle history
13
Toggle history
8.0
Toggle history
60
Toggle history
get

51
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
38
Toggle history
13
Toggle history
51
Toggle history
60
Toggle history
41
Toggle history
13
Toggle history
5.0
Toggle history
51
Toggle history
preventSilentAccess

60
Toggle history
18
Toggle history
60
Toggle history
No
Toggle history
47
Toggle history
13
Toggle history
60
Toggle history
60
Toggle history
44
Toggle history
13
Toggle history
8.0
Toggle history
60
Toggle history
store

51
Toggle history
79
Toggle history
60
Toggle history
No
Toggle history
38
Toggle history
13
Toggle history
51
Toggle history
60
Toggle history
41
Toggle history
13
Toggle history
5.0
Toggle history
51
Toggle history
Legend
Full support
Full support
No support
No support
Uses a non-standard name.
PublicKeyCredential
Report problems with this compatibility data on GitHub
desktop mobile
Chrome
Edge
Firefox
Internet Explorer
Opera
Safari
Chrome Android
Firefox for Android
Opera Android
Safari on iOS
Samsung Internet
WebView Android
PublicKeyCredential

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
49
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
getClientExtensionResults

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
49
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
isUserVerifyingPlatformAuthenticatorAvailable

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
49
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
rawId

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
49
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
response

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
49
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
Legend
Full support
Full support
Partial support
Partial support
No support
No support
See implementation notes.
AuthenticatorResponse
Report problems with this compatibility data on GitHub
desktop mobile
Chrome
Edge
Firefox
Internet Explorer
Opera
Safari
Chrome Android
Firefox for Android
Opera Android
Safari on iOS
Samsung Internet
WebView Android
AuthenticatorResponse

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
clientDataJSON

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
Legend
Full support
Full support
Partial support
Partial support
No support
No support
See implementation notes.
AuthenticatorAttestationResponse
Report problems with this compatibility data on GitHub
desktop mobile
Chrome
Edge
Firefox
Internet Explorer
Opera
Safari
Chrome Android
Firefox for Android
Opera Android
Safari on iOS
Samsung Internet
WebView Android
AuthenticatorAttestationResponse

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
attestationObject

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
getAuthenticatorData
Experimental

85
Toggle history
85
Toggle history
No
Toggle history
No
Toggle history
71
Toggle history
No
Toggle history
85
Toggle history
No
Toggle history
60
Toggle history
No
Toggle history
No
Toggle history
No
Toggle history
getPublicKey
Experimental

85
Toggle history
85
Toggle history
No
Toggle history
No
Toggle history
71
Toggle history
No
Toggle history
85
Toggle history
No
Toggle history
60
Toggle history
No
Toggle history
No
Toggle history
No
Toggle history
getPublicKeyAlgorithm
Experimental

85
Toggle history
85
Toggle history
No
Toggle history
No
Toggle history
71
Toggle history
No
Toggle history
85
Toggle history
No
Toggle history
60
Toggle history
No
Toggle history
No
Toggle history
No
Toggle history
getTransports
Experimental

74
Toggle history
79
Toggle history
No
Toggle history
No
Toggle history
62
Toggle history
No
Toggle history
74
Toggle history
No
Toggle history
53
Toggle history
No
Toggle history
No
Toggle history
No
Toggle history
Legend
Full support
Full support
Partial support
Partial support
No support
No support
Experimental. Expect behavior to change in the future.
See implementation notes.
AuthenticatorAssertionResponse
Report problems with this compatibility data on GitHub
desktop mobile
Chrome
Edge
Firefox
Internet Explorer
Opera
Safari
Chrome Android
Firefox for Android
Opera Android
Safari on iOS
Samsung Internet
WebView Android
AuthenticatorAssertionResponse

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
authenticatorData

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
signature

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
userHandle

67
Toggle history
18
Toggle history
60
footnote
Toggle history
No
Toggle history
54
Toggle history
13
Toggle history
70
Toggle history
92
Toggle history
48
Toggle history
13
Toggle history
No
Toggle history
No
Toggle history
Legend
Full support
Full support
Partial support
Partial support
No support
No support
See implementation notes.
Found a problem with this page?
Edit on GitHub
Source on GitHub
Report a problem with this content on GitHub
Want to fix the problem yourself? See our Contribution guide.
Last modified: Jun 10, 2022, by MDN contributors

Your blueprint for a better internet.

MDN on Twitter
MDN on GitHub
MDN
About
Hacks Blog
Careers
Support
Product help
Report a page issue
Report a site issue
Our communities
MDN Community
MDN Forum
MDN Chat
Developers
Web Technologies
Learn Web Development
MDN Plus
Website Privacy Notice
Cookies
Legal
Community Participation Guidelines
Visit Mozilla Corporation’s not-for-profit parent, the Mozilla Foundation.
Portions of this content are ©1998–2022 by individual mozilla.org contributors. Content available under a Creative Commons license.

"
 https://developer.mozilla.org/en-US/docs/Web/API/Web_Authentication_API#:~:text=Skip%20to%20main,Creative%20Commons%20license.

JetBrains Terms of Service (Space Cloud)
Version 1.0, effective as of December 9, 2020
Welcome to JetBrains Space!
This is a legal document and it is important that You read it carefully.
You understand that by accepting these Terms of Service (You do that by clicking the “I agree” or a similar button, or by accessing or using JetBrains Space) You are entering into a legal agreement and agree to certain legal consequences for Yourself or for Your Organization.
By accepting these Terms of Service, You confirm that You understand them, You agree with them, and You are at least 18 years of age.
1. Introduction
These JetBrains Terms of Service ("Terms"), describe how You can access, purchase, and use the in-cloud version of JetBrains Space.
Accepting these Terms creates a legal agreement between (i) JetBrains s.r.o., a company registered in the Commercial Register of the Prague Municipal Court, Section C, File 86211, ID No. 265 02 275 with its registered office at Na Hřebenech II 1718/10, Prague, 14000, Czech Republic ("JetBrains“, ”Us“, or ”We“) and (ii) yourself, that is either an organization, including a sole trader, one person organization or similar (”Organization“, or ”You"), or a physical person (“You”) for the Free Subscription only (as defined below).
If You are accepting these Terms on behalf of an Organization, such as (’including, but not limited to’) a company, organization, school, or charity, You confirm (’represent and warrant’) that You are authorized to enter into agreements on behalf of that Organization. If these Terms are accepted using an email address provided by a legal entity, We will regard (’deem’) You as authorized to represent that Organization. You must be able to enter into contracts (’have capacity’).
Summary: Accepting these Terms creates an important legal agreement between You and JetBrains. There are legal consequences to accepting these Terms, either for Yourself or the Organization You represent.
2. Definitions
a) Special legal phrases
There are certain phrases that have an accepted meaning for lawyers. To ensure these Terms are clear and accessible, We have included the accepted ‘legal’ phrase in parentheses after the word to show that We intend it to have the accepted ‘legal’ meaning. We do this every time a certain phrase is used for the first time in these Terms.
b) Definitions
There are words or phrases in these Terms that have a particular meaning. When the word or a phrase is used for the first time, it is defined and capitalized. These Terms also use these definitions:
“Applications” are either JetBrains or third-party software applications designed to be used in Space and available on the JetBrains Plugin Marketplace or from third parties.
“Content” refers to content that is featured, displayed, stored, or otherwise available in Space, such as (’including, but not limited to’) code, repositories, text, data, articles, images, photographs, graphics, software, Third Party Software, Applications, packages, designs, features, the Organization URL, and other materials.
“Confirmation” means an email confirming Your rights to use Space and describing important information about Your Subscription Plan, such as (’including, but not limited to’) the Subscription Period, number of Members, Resources that You are entitled to, as well as important payment information and the number of searchable messages and application integrations You can use.
“Documentation” means the latest versions of all online Space technical documentation, the ’JetBrains Team Tools Acceptable Use Policy’ (which outlines what You and Your Members can and cannot do in Space), and any other relevant Space policy available on the JetBrains Website which applies to You and the Members when using Space.
“Inactive Member” means a Member who, during any consecutive 14 day period, has not completed at least one intentional action in Space in any Client App (“Inactive”). If an Inactive Member performs an intentional action, they become an “Active Member” again. Creating or editing Content in Space, pushing to a Git repository, editing a profile, and reading chat messages are examples of intentional actions. However, auto-login, minimizing an application without navigation, and closing a Client App are not.
“JetBrains Website” means the Space product website and any other website operated by JetBrains including (but not limited to) websites listed on the JetBrains Legal Information page.
“Resources” means CI Credits, Data Transfer, Storage, and any other resource made available to You by JetBrains in Space.
“Space” means the JetBrains product offering known as “JetBrains Space”, offered in-cloud, comprising the JetBrains software program known as ‘Space’, which includes all downloadable parts of Space that are provided by JetBrains in binary form (if any), access to Space, the Documentation, updates of Space, and any incorporated Third-Party Software, as well as the Content and Resources.
“Subscription” means Your right to use Space according to these Terms and the Documentation, and within the limits set out in Your Subscription Plan and described in Your Confirmation.
“Subscription Period” means either a monthly or yearly period as described in Your Subscription Plan.
“Subscription Plan” means subscription plans described in Your Confirmation and the specific features for each type of plan described on the JetBrains Website and/or in the Documentation. If the description in Your Confirmation is different from the description on the JetBrains Website or in the Documentation, the description in Your Confirmation will prevail.
“Third-Party Software” means third party software programs that are owned or licensed by someone other than Us and described on the JetBrains Website.
“Member” means a person who is authorized by You to use Space and who has Your permission to access and use Space under Your Subscription.
“Your Content” means Content that You (any of Your Members) create, own, or have the right to use.
Summary: Words starting with capital letters have a special meaning. These are defined in this Section or wherever they are used for the first time in these Terms.
3. Rights and responsibilities
a) Right to use Space
You can use Space as long as You comply with these Terms, the Documentation and the limits of Your Subscription. You can change Your Subscription at any time, including by purchasing additional Resources or Member access rights. Any changes to Your Subscription will be effective as soon as We confirm those changes.
We will use commercially reasonable efforts to make Space available to You. Space may be unavailable to You during (i) planned downtime, (ii) failures of Space, including failures or delays contributed to by an internet service provider, or (iii) any unavailability caused by circumstances beyond JetBrains’ reasonable control (see the ‘Force Majeure’ Section).
b) Your responsibilities when using Space
You are responsible for:
i) Organization Accounts - creating and maintaining an electronic record in Space for Yourself ("Organization Account“), which allows a representative of Your Organization to manage Your Organization’s Subscription as system administrator (”Admin Member"), manage Resources, and create, manage, deactivate and/or delete Member Accounts. You are responsible for making sure Your Organization’s representative is authorized to use the Organization Account;
ii) Member Accounts - creating and maintaining electronic records for one or more Members ("Member Account“), which allow Members to join and use Space. The number of Members You can invite depends on Your Subscription Plan and whether the ”Overdraft" feature is enabled (see the “Overdraft” Section);
iii) External Account - creating and maintaining electronic records for Members who may not usually be affiliated with Your Organization ("External Member“), but to whom You give limited access to use Space (”External Account"). You are responsible for the permission You give an External Member, their activities in Space, and the Content to which they will have access;
iv) Members - Your behavior, Your Members’ and External Members’ behavior, and making sure that You, Your Members, and any External Members do not breach these Terms. If You become aware that a Member or External Member is breaching these Terms, You must immediately cancel that Member’s rights to use Space by suspending their Member Account;
v) Confidentiality - keeping Your Organization Account, Your Content, passwords, Members’ usernames, and access tokens confidential and secure, and making sure that Your Members do the same;
vi) Permitted use - configuring and using Space according to the Documentation and Your Subscription Plan;
vii) Internet and software - making sure that You have a suitable internet connection and any equipment that You need for that internet connection. It is also Your responsibility to have access to appropriate hardware and any third-party software needed to run Space, such as a browser with compatible data security protocols;
viii) Your Content - all Content that You or Your Members submit or allow to submit to Space, and all of Your Content that is stored by JetBrains on Space, including any permission You need to use Your Content. You are also responsible for the way in which You acquired Your Content, and for ensuring that it is legal for You to use Your Content and for You to submit it, or allow it to be submitted, to Space. If You become aware that any part of Your Content breaches these Terms or any other person’s (“third-party’s”) rights, You must immediately remove this part of Your Content from Space;
ix) Legal use - making sure that Your use of Space does not breach applicable law or government regulations.
c) Restrictions on using Space
You must not:
i) Interfere - reverse-engineer, disassemble, or decompile Space or try to derive the source code of Space in any way, unless applicable law allows this;
ii) Steal - modify, alter, tamper with, repair, or otherwise create derivative works of Space, except if We give You a separate license that expressly allows You to create derivative works of all or part of Space;
iii) Cheat - use, or try to use, Space in a way that avoids incurring fees or exceeding the limits for Your Subscription Plan. You also must not obtain Resources in a way that breaches these Terms;
iv) Transmit illegal Content - use Space to upload, store or share, or allow others to upload, store, or share (“transmit”) any material that is criminal, offensive, defamatory, or otherwise unlawful or a tort, or breach the privacy or intellectual property rights of anyone else (“third-party”); and
v) Gain unauthorized access - try to gain unauthorized access to Space, or allow anyone else to access Your or anyone else’s Member Account, or allow anyone outside of Your Organization to use Space other than through an External Account.
vi) Resell Space or access to Space to any third party.
You also must make sure that each Member does not do any of these things.
Summary: You can use Space according to these Terms. Do not breach the restrictions outlined above, as they are an important part of Our mutual agreement.
4. Subscriptions
a) Free Subscriptions - when You sign up for Space, You can use Space for free (i.e. on the ‘Free’ or Subscription Plan with a similar name) ("Free Subscription"). With a Free Subscription, You can have as many Members as You need (subject to our Acceptable Use Policy) and use the maximum number of Resources described on the JetBrains Website at the time You sign up.
b) Upgrading Subscriptions - You can change Your Subscription from a Free Subscription to a paid Subscription (ie. a Subscription You have to pay for) ("Paid Subscription“) or change one type of Paid Subscription to another type of Paid Subscription with more Resources (”Upgrade“) at any time. Depending on the Subscription Plan that You select, Your Subscription will be either for one month or one year (”Subscription Period"). Your Subscription starts on the date in Your Confirmation.
When Upgrading, Your Paid Subscription will be set to an annual Subscription Period by default and will include a starting number of Members. You can manually set Your Paid Subscription from an annual to a monthly Subscription Period. The number of Members that You start Your Upgraded Subscription with is based on the number of Members who are not Inactive. For the purposes of calculating the number of Active Members, an External Member who is not Inactive will be counted as an Active Member.
c) Downgrading Subscriptions - You can downgrade from any Paid Subscription to a Free Subscription or change from one type of Paid Subscription to another type of Paid Subscription with fewer Resources ("Downgrade") at any time. If You Downgrade from a Paid Subscription to a Free Subscription, We will refund You the unused portion of Your Paid Subscription, including any additional Members or Resources, as General Credits (defined below). If You Downgrade from one type of Subscription to another, We will refund You the difference between the types of Paid Subscriptions as General Credits.
d) Automatic Renewals - unless You explicitly opt out, Your Subscription, and the Subscription Period, renew automatically. This means that if Your monthly Subscription is about to expire, it will be automatically renewed for another month. If Your annual Subscription is about to expire, it will be automatically renewed for another year. We will notify You shortly before Your Subscription is renewed. You can change Your Subscription Period or opt out of the automatic renewal of Your Subscription in Your Organization Account at any time.
e) Trial Subscriptions - You may be eligible for a trial Subscription. The details of trial Subscriptions are displayed within Space. These trials are free and must be used only to assess which Space Subscription Plan suits Your needs. You can request a trial Subscription once (unless agreed otherwise with JetBrains) and a trial doesn’t automatically increase Resources to the trial Subscription Plan Resource limits during Your trial Subscription. Once the Trial Subscription ends, You will be downgraded to the Subscription Plan that You had before the trial Subscription began, including any additional Members or Resources that You previously purchased.
Summary: Please pay attention to the time period in which You are entitled to use Space, the fact that it auto-renews, and the number of Members and other Resources You have purchased. If You need to add more Resources, please do so in the Organization configuration, or let Us know.
5. Member Content
a) Responsibility for Your Content
You can create or upload Your Content (see the ‘Definitions’ Section) while using Space. You are solely responsible for all of Your Content that You post, upload, link to or otherwise make available or allow others to make available on Space, regardless of the form of that Content.
You are also responsible for all legal consequences, such as claims, damages, losses, liabilities, costs, and expenses that result from Your Content. We are not responsible for any public display or misuse of Your Content.
b) Ownership of Your Content
You keep (’retain’) ownership, title, and interest of Your Content.
You will only submit or allow submission of Content, including Third-Party Software, that You have the right to use, display, publish and/or modify. You will fully comply with any third-party rights relating to Your Content. This means that if Your Content is licenced or copyrighted by a third-party, You must make sure You have the right to submit this Content to Space and You must include any notices as required by the copyright owner or licensor.
Each time You post something that You did not create Yourself, or that You do not own the rights to, You confirm that You have the right to do so and understand that You are doing so at Your own risk, and are solely responsible for this Content and all consequences of its use in Space. You also indemnify Us for any liability relating to this Content (see the ‘Indemnification’ Section).
c) Removing Content
We do not review, screen, or otherwise moderate Content and are not responsible for doing so. We have the right, but not the responsibility, to refuse or remove any Content or close any Member Account that We (“in Our sole discretion”) believe breaches these Terms, any other legal agreement with JetBrains, any other JetBrains policies, or someone else’s rights.
If You believe any Content affects (“infringes”) Your rights, please let Us know by emailing Us at copyrights@jetbrains.com.
d) Permissions to handle Your Content
You need to give (“grant”) Us certain permissions (“rights”) so that We can provide Space to You and make Your Content submitted or allowed to be submitted by You accessible in Space. The exact scope of such permissions is described in the Sections below (see Sections 5(d)(i) - 5(d)(iii)).
Each of these permissions takes effect immediately when Your Content is submitted on Space. Each permission ends when Your Content is removed from Space, but for backups, these permissions will last longer as described in these Terms (see the “Data Retention” Section) and Our Data Retention Policy. You understand (“acknowledge”) that You will not receive any payment for giving Us these permissions.
If the Content You upload already comes with a standalone permission that allows Us to make it accessible within Space as described in these Terms, either the permissions as described here or the standalone permission will apply, whichever is broader.
i) Content permission that You grant to us
You give Us permission to host, store, copy, parse, display, publish and share with Your Members Content in Space, and You allow it to be shared in Space with Your Members. This permission includes the right to do things such as copy it to Our database and make backups and analyze it on Our servers. However, nothing here gives Us permission to sell or otherwise transfer ownership of Your Content to a third party, nor does anything here give Us permission to grant access to Your Content to any third party without Your explicit permission.
ii) Content Permission that You grant to Members
You understand that, depending on the specific settings You choose in Space, Your Members may be able to access and use all of Your Content. You and, if You allow them to do so, Your Members can give other Members the right to access and use Your Content. It is Your responsibility to set Your Members’ access and use rights to Your Content. These rights apply to all Members in Your Organization. These rights can be given to multiple Members (“non-exclusive”) and apply worldwide.
If You are uploading Content that You did not create or do not own, You are responsible for ensuring that You have the right to upload this Content and give Us the same right to access and use this Content as described above in Section 5(d)(i).
iii) Moral Rights
You keep all moral rights to all of Your Content that You upload, publish, or submit to Space. This includes rights of integrity and attribution. However, You waive these rights and agree not to assert them against us, but only so that We can do the things described in clause i) of this Section (“Permissions to handle Your Content”). If a court finds that these Terms are not enforceable under applicable law, You grant JetBrains the rights that We need to use Your Content without attribution and to make reasonable adaptations to Your Content, but only to the extent that is necessary to enable Us to provide Space.
Summary: Any Content created by You remains Yours. However, You provide Us with certain rights to it, so that We can display and share the Content You post. You have control over Your Content, and responsibility for it, and the rights You grant Us are limited to those necessary for Us to provide Space. We have the right to remove Content or close Member Accounts if We need to.
6. Content Access
a) Access control
Depending on the nature of Your Content and the specific Space feature that You are using, Your Content might be visible to other Members by default. It is Your responsibility to set the appropriate level of access to Your Content, as described in the Documentation.
b) Content protection & confidentiality
We will protect any Content that You upload from unauthorized use, access, or disclosure. We will exercise commercially reasonable efforts to protect it in the same way (’manner’) that We protect Our own content that is similar in nature and no less than with a reasonable standard of care.
c) Access
You give Us permission to access Your Content in the following situations:
i) For support reasons - If You request support, Your Admin Member can grant Us access with the same permissions as You or another Member if such access is necessary to carry out the support task. You can revoke these permissions at any time.
ii) For security reasons - We can access Your Content if We have a good reason to (“reasonably”) believe this access is required to maintain the ongoing confidentiality, integrity, availability, performance, and resilience of JetBrains’ systems and Space.
iii) If You give Us permission - We can access Your Content in the context of providing You with support if You give Us permission to do so. You can enable services or features in Space that give Us or other Members additional access rights. If any services or features require permissions other than those You have given us, We will provide an explanation of those permissions.
iv) If We are legally required - We have the right to access, review, and remove all or a part of Your or Your Members’ Content if We have a good reason to (’reasonably’) believe that Your Content breaches the law or these Terms. You understand that there are laws that could require Us to disclose Your Content and, if these laws apply, We are obliged to comply with them.
Summary: You are responsible for deciding who has access to Your Content in Space. External Members will see Your Content only if You allow it. We protect Your Content, and We only access it for support reasons with Your consent, or if required to for security or legal reasons.
7. Fees and Payments
a) Subscription and Other Fees
When You sign up for Space, You will have a Free Subscription and will not pay any Subscription fees. You can Upgrade Your Subscription at any time and will start paying Subscription fees depending on Your Subscription Plan, Subscription Period, number of Active Members, the pricing described on the JetBrains Website, and Your chosen method of payment.
Depending on Your Subscription Plan, You will have access to different features and Resources, and be subject to certain limits. These features, Resources and limits are described on the JetBrains Website and apply at the time You Upgrade. The most important limits include:
i) Your Subscription Period;
ii) the number of Active Members linked to Your Organization Account;


iii) the total amount of data You and Your Members can transfer in Space per month by uploading or downloading Content to and from Space ("Data Transfer");
iv) the total number of gigabytes of storage available to You and Your Members for use in Space ("Storage");
v) the number of searchable messages available across Your Organization. A searchable message includes chat messages or comments posted to an issue, code reviews, or blog posts;


vi) the number of Applications and/or integrations (see the “Space Applications and Integrations” Section below);
vii) the level of support (see the “Support” Section below); and


viii) CI Credits (see the “Credits” Section below).
You can monitor key aspects of Your Subscription using the relevant page in Space.
b) Subscription Billing
You will be billed either monthly or annually depending on Your Subscription Plan, Subscription Period and the method by which You choose to pay ("Billing Period").
i) Annual Subscriptions - if You have an annual Subscription, We will bill You at the beginning of the annual Subscription Period. You will be charged for the number of Active Members and Resources that are included in Your Subscription Plan and that are described in Your Confirmation. You can choose to pay with any major debit or credit card (“Payment Card”) or, for some of the Subscription Plans as specified on the JetBrains Website You can activate payment by electronic funds transfer within 30 days of receiving an invoice ("EFT") within Your JetBrains Account or during Your purchase of Space.
ii) Monthly Subscriptions - if You have a monthly Subscription, We will bill You at the beginning of the monthly Subscription Period. You will be charged on the basis of Your Subscription Plan, including the Resources available to You, and the number of Active Members on the billing date. You can pay for monthly Subscriptions by Payment Card. For monthly Subscriptions, the VAT supply date is the last date of the month.
iii) Subscription Renewals - when Your Subscription is automatically renewed, We will bill You based on the number of Active Members at the time of Your renewal.
iv) Refunds for Inactive Members - if You have a Paid Subscription, and any of Your Members become Inactive during the Subscription Period, We will refund part of Your Subscription fee. You will be charged for the first 14 consecutive days during which a Member is Inactive, but not for the rest of the Subscription Period applicable to that Member. This refund is only available as General Credits, which We will assign to Your Organization Account no later than at the beginning of the month immediately following the one when a Member becomes Inactive (see the “Credits” Section). The General Credits refund for Inactive Members will be assigned for a given month at the beginning of the next month as long as the Member is Inactive and has not been replaced by another Member. This means You will receive a refund of General Credits based on the duration of the Subscription during which the Member was Inactive.
If You deactivate a Member’s Account, We will regard that Member as Inactive and You will be entitled to a refund for the rest of the Subscription Period. This will be a pro-rata refund for the period beginning on the day the Member Account was deactivated and ending at the end of the Subscription Period.
If an Inactive Member is no longer Inactive, then You will be charged for the period beginning when the Member is no longer Inactive and ending at the end of the relevant Subscription Period. You will be billed General Credits if You have any available, and if You don’t have General Credits We will bill Your Payment Card or via EFT.
v) Change billing period - You can change Your Billing Period from monthly to annual at any time, and this change will be effective no later than as of the first day of the following month. If You make this change, Your first annual bill will include amounts relating to the previous monthly Billing Period, as well as the new annual Billing Period. You also can change billing from annual to monthly at any time, but the change will be only effective from the beginning of the next Billing Period.
c) Additional Members and Resources Billing
i) Additional Members & Resources - if You have a paid Subscription, You can extend Your Subscription to include additional Member access rights at any time during Your Subscription Period by activating the “Overdraft” feature (see the “Overdraft” Section) ("Additional Members"). Each Additional Member’s Subscription will begin when they become an Active Member of Your Organization and will end at the end of the Subscription Period in which You added the Additional Member (’is co-termed’).
By activating the Overdraft feature, You can buy additional Resources which will be billed after the end of the applicable Billing Period.
ii) If You add any Additional Members or Resources during Your Billing Period, We will charge You for those Additional Members or Resources not earlier than on the first day of the following month. As Subscriptions for Additional Members are co-termed, Your Subscription fee will be calculated on the basis of the remaining Subscription Period. If any additional Member becomes Inactive during the Billing Period, We will apply a pro-rata refund of General Credits for the period of Inactivity (see the “Refunds for Inactive Members” Section).
d) Overdraft
In Your Organization Account, You can choose to enable the Overdraft feature. By enabling this feature You can use more Resources or Members than You initially bought (i.e. those described in Your Confirmation) during the Subscription Period. This means that You will be able to use additional Members or Resources up to the limits allowed in the Overdraft feature ("Overdraft Limit"). Your Overdraft Limit is decided by Us based on:
the number of average Active Members in Your Organization (i.e. the base Overdraft Limit is calculated as a coefficient of Your monthly Subscription invoice amount);
the payment method that You have selected (i.e. Payment Card or EFT); and
any outstanding unpaid amounts and Your overall payment history.
If You reach the Overdraft Limit, We will issue an invoice for additional Members and Resources immediately, and, after You pay this invoice, the Overdraft Limit will be reset. You can also pay fees for additional Members and Resources at any time. All fees for any Overdraft must be paid within 30 days of the end of the calendar month in which the Overdraft was allocated. If You have a Free Subscription, You cannot enable the Overdraft feature.
e) Credits
Space allows You and Your Members to buy non-refundable credits ("General Credits“) that can be used in Space to purchase Resources or additional Members. The exact Resources that You can purchase with General Credits is described in Your Organization Account and can change at any time. The exchange rate of the Resources is available in the Organization Account. General Credits can also be used to pay for running automations in Space, and these are tracked as ”CI Credits". CI Credits cannot be purchased directly, but may be included in Your Subscription Plan or obtained via Overdraft. One CI Credit entitles You or Your Members to one minute of automation in the Container with Default Resources as defined in the Space Documentation.
You understand that General Credits can only be issued by JetBrains. General Credits are not real money (“legal tender”), currency, cryptocurrency, a voucher, or a prize, and have no cash value. They can be purchased from Us, but they cannot be sold, traded off, transferred, exchanged, or bartered with, and can only be used to purchase Resources or additional Members.
Neither General Credits nor other Resources are refundable. General Credits will expire if these Terms are terminated (see the “Term and Termination” Section). The Resources are tied to Your Subscription Plan, and change with it accordingly. That means You will get different amounts of Resources when Your Subscription Plan changes.
f) Payments
i) Payment Terms - The JetBrains Terms and Conditions of Purchase apply to all fees and other amounts that You have or might have to pay (“are payable”) relating to these Terms.
ii) Payment methods - You can pay either by Payment Card, which is available to any Subscription Plan and for any amount, or by EFT, which is available only to certain Subscription Plans as detailed on the JetBrains Website/FAQ;
iii) Interest - We can charge You interest at the rate of 1.5% per month (or the highest rate permitted by applicable laws, if that rate is less than 1.5% per month) on all late payments.
iv) Withholding - You cannot deduct or withhold (“set-off”) any amount from the fees that You have to pay to JetBrains, even if We owe You an amount or You believe We owe You an amount (“counterclaim”), unless We agree to do so in writing.
v) Taxes - Fees quoted in Your Confirmation exclude any and all applicable taxes and similar fees (other than taxes solely based on JetBrains’ income) now in force or imposed in the future on provision of the Service. You are responsible for all taxes, levies and/or duties, such as value added tax (’VAT’), sales tax, and withholding tax, that apply in Your country. You have to pay these on top of fees payable to JetBrains.
g) Resolution of late payments
To continue using Space without interruption, You must make sure that You pay all the relevant fees on time. Payment dates are described either in the Organization Account or an invoice. If You do not pay all fees in full and on time, We can:
i) Limit Your or a Member’s access to Space or any features in Space.
ii) Suspend or altogether end (“terminate”) Your access to Space and terminate these Terms as described in Sections 15 and 16.
You will reimburse Us for any additional costs that We incur in collecting late payments or if You breach anything in this Section.
8. Support
Your Subscription includes the support included with Your Subscription Plan and outlined on the JetBrains Website ("Support"). We will provide Support only to the extent required for You to use Space according to the Documentation and only in relation to current Client Apps.
You (or an Admin Member) can request Support by submitting a support ticket at any time. We will try to respond to Your request in a reasonable period of time. If it is needed in order for Us to provide Support, We can ask You to provide Us with access to the unique URL that was assigned to Your Organization and that allows You to use Space ("Organization URL"), by giving Us the appropriate permissions in the Support settings of Your Organization Account. To withdraw this permission, You must change these settings.
You understand that We can resolve a Support request by deciding in Our sole discretion to implement a publicly available patch, upgrade, or release in the future, or by choosing to modify certain features, functionality, or settings.
9. Space Applications & Integrations
a) Space Applications
You can access free and paid Space Applications from the JetBrains Plugin Marketplace and use them in Space. Space Applications are not included in Your Subscription Plan and You will need to acquire them from the JetBrains Plugin Marketplace, accepting the relevant terms and conditions for each individual Space Application. If a Space Application is owned by someone other than JetBrains, You may be required to accept their terms and conditions.
You are responsible for deciding whether a particular Space Application is compatible with Space and suitable for Your needs, and for assessing how it might affect Your Subscription. You are also responsible for installing and connecting a Space Application with Space. You may be able to co-term certain paid Space Applications with Your Subscription Period.
b) Integrations
You can integrate certain Space functionality with software and/or services that are not part of Your Subscription, or owned or operated by JetBrains. The software and/or services that You can integrate with are described in Your Organization Account and can change at any time.
10. Ownership
a) We own Space
We own (or have the right to use) all the proprietary and intellectual property rights to Space and to all related trade secrets, copyright, trademarks, service marks, patents, and other unregistered intellectual property. These are Our rights (“rights are reserved”). The only intellectual property rights that You have in relation to Space are those that are necessary in order for You and Your Members to access and use Space according to the Documentation.
b) You own Your Content
You keep ownership of all proprietary and intellectual property rights to Your Content. This means that We never own any of Your Content even though it is collected, passed through (“transmitted”), or created in Space.
c) Feedback
You give Us the right to use, change (“modify”), commercialize, and incorporate into Space any of Your ideas, suggestions, recommendations, proposals or other feedback relating to Space. You cannot withdraw this permission after it is given (“irrevocable”) and it is perpetual. We are not required to pay a fee for this feedback (“royalty-free”), and We can transfer and give similar rights (“sublicense”) to Your feedback to anyone else worldwide.
d) Third-Party Software and its associated Rights
You understand that the Software integrates Third-Party Software and that by using Space You might be using Third-Party Software. This Third-Party Software is provided to You on the terms and conditions of the respective Third-Party Software and You need to comply with those terms and conditions.
11. Client Apps
You can use Space in a suitable browser, in any supported desktop applications (MacOS, Windows and Linux), in a mobile environment (iOS, Android), or through a supported JetBrains IDE or API ("Client Apps").
You understand that these Terms apply to Your use of Space in any of these Client Apps. You also confirm that You have accepted the relevant terms and conditions when accessing a Client App and understand that all trademarks associated with the Client Apps are the property of their respective owners.
12. Indemnification
a) Indemnity
If there are any claims, damages, losses, liabilities, fees and similar expenses (including reasonable attorney’s fees) brought against JetBrains that arise out of, or are related to, any of the following things:
i) Access and use of Space - Your access or use of Space or any access or use of Space by any of Your Members. This includes all activities related to Your Organization URL and any actions taken by Your employees and personnel in relation to Space;
ii) Breach of these Terms - if You or any of Your Members breach these Terms or applicable laws;
iii) Your Content - Your Content or the combination of Your Content with other Applications, Content, or processes. This includes any allegation that Your Content, or its use, development, design, production, advertising, or marketing infringes someone else’s (“a third party’s”) rights, or that You have illegally or without permission claimed someone else’s rights; or
iv) Disagreements - in the event of a disagreement between You and any of Your Members or any other third party (each of these is a "Claim"),
then You agree to indemnify, defend, and hold Us and Our owners, directors, employees, agents, and representatives harmless, and to indemnify, defend and hold Our affiliates and their owners, directors, employees, agents, and representatives harmless, from any and all such Claims.
b) Indemnity claims
We will quickly (“promptly”) let You know if any of the things above happen (see the “Indemnity” Section above) and someone makes a claim. If We fail to let You know quickly, then that failure will only affect Your obligation to indemnify Us to the extent that Our failure to inform You quickly adversely affected Your ability to defend the claim. When defending such a claim You can choose Your own lawyer, with Our written permission. If You have Our written approval, You can resolve (“settle”) the claim as You decide (“at your discretion”). However, We can take control of the defence and settlement at any time.
13. IMPORTANT - YOUR RISK AND OUR DISCLAIMERS
(RISK) SPACE IS PROVIDED ON AN “AS IS” AND “AS AVAILABLE” BASIS. YOU ACCESS AND USE SPACE AT YOUR OWN RISK.
(WARRANTIES & REPRESENTATIONS) EXCEPT AS EXPRESSLY SET OUT IN THESE TERMS, WE MAKE NO REPRESENTATIONS AND GIVE NO WARRANTIES IN RELATION TO SPACE - EXPRESS, IMPLIED, STATUTORY, OR OTHERWISE. THIS INCLUDES WARRANTIES THAT SPACE WILL BE UNINTERRUPTED, ERROR-FREE, OR FREE OF HARMFUL COMPONENTS, AS WELL AS WARRANTIES THAT YOUR CONTENT WILL BE SECURE OR NOT OTHERWISE LOST OR DAMAGED.
WE ALSO DENY (“DISCLAIM”) ALL WARRANTIES. THIS INCLUDES ANY IMPLIED WARRANTIES OF MERCHANTABILITY, SATISFACTORY QUALITY, FITNESS FOR A PARTICULAR PURPOSE, OR NON-INFRINGEMENT AND ANY WARRANTIES ARISING OUT OF ANY COURSE OF DEALING OR USAGE OF TRADE.
THIS DISCLAIMER DOES NOT APPLY TO REPRESENTATIONS AND WARRANTIES THAT CANNOT BE EXCLUDED BY LAW.
14. IMPORTANT - LIMITATION OF OUR LIABILITY
(TYPES OF DAMAGES) WE WILL NOT BE LIABLE TO YOU OR A MEMBER FOR ANY DIRECT, INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR EXEMPLARY DAMAGES. THIS INCLUDES DAMAGES FOR LOSS OF PROFITS, GOODWILL OR DATA, EVEN IF WE HAVE BEEN ADVISED OF THE POSSIBILITY OF SUCH DAMAGES.
(CIRCUMSTANCES OF LOSS) WE WILL NOT BE LIABLE FOR ANY COMPENSATION, REIMBURSEMENT, OR DAMAGES ARISING IN CONNECTION WITH:
a) YOUR, OR A MEMBER’S, INABILITY TO USE SPACE, INCLUDING AS A RESULT OF A SUSPENDED SUBSCRIPTION, OR THE CANCELLATION OF YOUR SUBSCRIPTION OR THESE TERMS;
b) OUR DECISION TO NO LONGER PROVIDE SPACE FOR BUSINESS, ECONOMIC, LEGAL OR REGULATORY REASONS;
c) HAVING MADE SPACE AVAILABLE TO YOU OR A MEMBER;
d) ANY UNANTICIPATED OR UNSCHEDULED DOWNTIME OF SPACE OR A PART OF SPACE FOR ANY REASON, INCLUDING AS A RESULT OF POWER OUTAGES, SYSTEM FAILURES, OR OTHER INTERRUPTIONS;
e) THE COST OF PROVIDING A SUBSTITUTE FOR SPACE;
f) ANY INVESTMENTS, EXPENSES, OR COMMITMENTS THAT YOU OR A MEMBER MAKE RELATING TO THESE TERMS OR YOUR ACCESS TO OR USE OF SPACE; OR
g) ANY UNAUTHORIZED ACCESS TO, MODIFICATION OR DELETION, DESTRUCTION, DAMAGE, LOSS, OR FAILURE TO STORE ANY OF YOUR CONTENT.
(MAXIMUM LIABILITY) OUR MAXIMUM, OVERALL (“AGGREGATE”) LIABILITY RELATING TO THESE TERMS IS LIMITED TO THE AMOUNT THAT YOU ACTUALLY PAID TO US FOR SPACE AND RESOURCES IN THE SIX (6) MONTHS BEFORE YOU CLAIMED THAT WE WERE LIABLE. THE MAXIMUM LIABILITY APPLIES EVEN IF WE WERE ADVISED THAT LIABILITY COULD EXCEED THE MAXIMUM LIABILITY AMOUNT OR EVEN IF THE LEGAL BASIS (I.E. TORT, BREACH OF CONTRACT, EQUITY, OR A SIMILAR BASIS) FOR A REMEDY IS INVALID.
15. Temporary Suspension
We can immediately suspend Your or any of Your Members’ right to use Space, or any part of Space, as soon as We let You know (“give notice”) that We have a good reason to (“reasonably”) believe that:
a) Threats - Your use of Space might adversely impact, or pose a security, privacy or legal risk to JetBrains, Space, or any other person (“third party”). This also applies to Your Members’ use of Space;
b) Failure to pay - You do not comply with the payment obligations in these Terms (see the “Fees and Payment” Section);
c) Financial distress - You have stopped operating in the usual course of business, have transferred (“assigned”) Your assets for the benefit of creditors or made a similar arrangement, or are undergoing bankruptcy, reorganization, liquidation, dissolution or a similar proceeding; or
d) Breach of Terms - You breach these Terms or applicable law.
16. Term and Termination
a) Term
These Terms start (“have effect”) when You click the “I Accept” button or provide similar consent to (“be bound by”) these Terms. These Terms continue until they are ended (“terminated”) either by You or Us ("End Date") as described in these Terms.
b) Ending this agreement
Either You or We can terminate these Terms if the other party breaches them. You must let the other party (“give notice”) know that it has breached these Terms and, if these breaches are not resolved within 30 days, these Terms will end. If You end these Terms according to this Section, We are not required to refund You any prepaid fees for the period that would be Your Subscription Period, after the date these Terms were ended.
If We end these Terms according to this Section, You will pay Us any unpaid fees that You have to pay for the period that would be Your Subscription Period, after the date these Terms were ended.
c) Termination by us
In addition, We can immediately end these Terms, if We decide that:
i) You have materially breached or abused any part of these Terms and have not remedied this in 3 consecutive days after We let You know;
ii) We will no longer provide Space, due to any business, economic, legal, or regulatory reason; or
iii) You are the holder of a Free Subscription and all Your Members are Inactive Members for 3 consecutive months.
If We end these Terms according to the Section 16(c)(i) above, You will pay Us any unpaid fees that You have to pay for the period that would be Your Subscription Period, after the date these Terms were ended.
d) Your Content at Termination
If You or any of Your Active Members stop using Space for any reason, We will store Your Content and make it available to You for export (download) only for a certain period of time. Your Content will be available for 6 months and 2 weeks after You or any of Your Active Members stop using Space. You understand that after these time periods, Your Content will be deleted.
You understand that there is no feature in Space that allows You to export all Your Content directly from Space; You will need to do this through the application programming interface (API).
We reserve the right to remove Your Content from Space in the event that Your Content exceeds the amount of Resources associated with Your Subscription Plan.
You also understand that We will not have any responsibility to store Your Content or make it available to You and, unless We are legally prevented from doing so, We can remove Your Content from Space. We will let You know about any planned deletion of Your Content. We will use commercially reasonable efforts to keep a backup of Your Content for 1 month after it is deleted. You understand that it will not be possible to restore Your deleted Content after the backup is deleted.
e) Manual deletion
You can request the manual deletion of Your Content currently stored by Us by sending a request to the privacy@jetbrains.com email address. JetBrains will use commercially reasonable efforts to keep an automatic backup of Your deleted Content deleted for 1 month after deletion. However, We will also delete the backup of Your deleted Content if You request Us to do so.
17. Marketing
If You are a legal entity, You give Us permission to publicly identify You as a customer of JetBrains, refer to You by name, trade name, and trademarks, and describe Your business. You give Us permission to do this, but only for marketing purposes. We can use Your name, trade name, and trademarks in marketing materials, on the JetBrains Website, and in other public documents. We are not required to pay a fee for this permission (it is “royalty-free”), and it applies worldwide.
18. Notices
If You are required under these Terms to notify Us (“give notice”) of anything, You may do so:
a) by sending an email to legal@jetbrains.com. Any time period starts on the next business day after You send the email;
b) by courier delivery of a letter marked for the attention of the “Legal Department” at the physical address on the JetBrains Website. Any time period starts 5 consecutive days from when You send the letter; and
c) by registered post, marked for the attention of the “Legal Department” at the address on the JetBrains Website. Any time period starts 10 consecutive days from when You send the letter.
If We are required under these Terms to notify You (“give notice”) of anything, We may do so:
d) by posting the information on the JetBrains Website. Any time period starts on the day specified on the JetBrains Website;
e) by sending an email to the email address that Your Confirmation was sent to. Any time period starts on the next business day after We send the email.
It is Your responsibility to check the JetBrains Website for any changes and make sure that Your email address is up to date in Our records.
19. General Provisions
a) This Agreement and its Parties
The JetBrains Privacy Policy, the JetBrains Conditions of Purchase and JetBrains Team Tools Acceptable Use Policy are part of (“incorporated into”) these Terms. Together, these documents form the entire agreement and replace any previous agreement between You and Us in relation to its subject matter. Except as expressly mentioned, these Terms do not apply or give rights to anyone else (“no third-party beneficiaries”).
b) Organization User Agreements
You can require members to accept Your Organization’s user agreement. By activating the ‘User Agreements’ feature in Space, You can request Your Members to accept a user agreement between You and a Member ("Organization User Agreement"), which must comply with applicable law, be consistent with these Terms and the Documentation. Any part of Your Organization User Agreement that is illegal or inconsistent with these Terms, the Team Tools End-User Agreement, or any Documentation will not apply and You are responsible for the Content, correctness, and all other aspects of Your Organization User Agreement. You also understand that any records relating to Your Organization User Agreement are provided for convenience only and are subject to Our Data Retention Policy.
c) Governing Law and Disputes
These Terms are governed by the laws of the Czech Republic, without regard to conflict of laws principles. You agree that any litigation relating to these Terms may only be brought in, and will be subject to the jurisdiction of, any competent court of the Czech Republic. The United Nations Convention on Contracts for the International Sale of Goods does not apply to these Terms. Notwithstanding this, You agree that JetBrains shall still be allowed to apply (A) for payment orders (or otherwise enforce payment for Space provided under the Terms) in the jurisdiction in which You have Your registered seat or principal place of business, and (B) for injunctive remedies (or an equivalent type of urgent legal relief) in any jurisdiction.
If there is a disagreement (“dispute”) regarding these Terms, You and JetBrains will each do Your best (“use best efforts”) to settle the disagreement in a respectful, constructive, and non-litigious way. Should the parties fail to settle a dispute amicably, all disputes arising from the present Terms and/or in connection with it shall be finally decided with the Arbitration Court attached to the Czech Chamber of Commerce and the Agricultural Chamber of the Czech Republic according to its Rules by three arbitrators in accordance with the Rules of that Arbitration Court.
d) Force Majeure
We will not be responsible (“liable”) for any delay or failure to perform any obligation under these Terms where the delay or failure results from any cause beyond Our reasonable control. This includes any “acts of God”, labour disputes or other industrial disturbances, systemic electrical, telecommunications, or other utility failures, public health emergencies, earthquakes, storms or other elements of nature, blockages, embargoes, riots, acts or orders of government, acts of terrorism, or war.
e) Severability
If a court finds that any part of, or word in, these Terms is not enforceable, that part or word will not affect the enforceability of the rest of these Terms.
f) Interpretation
Any heading, title or paragraph summary is only for convenience and does not affect interpretation of these Terms. Any references to an inclusive word, such as ‘including’, is not comprehensive and refers to other items in that category. References to time or periods of time are determined in reference to Central European Time.
g) Waiver9
Any waiver of Our rights under these Terms must be in writing and signed by us.
h) Changes to Terms and Policies
We can update or modify these Terms at any time by posting a revised version to the JetBrains Website. The modified Terms will start (“be effective”) on the date they are posted on the JetBrains Website. By continuing to use Space after the effective date, You agree to be bound by the modified Terms. It is Your responsibility to check the JetBrains Website regularly for any changes to these Terms.
i) Relationship
Your relationship with JetBrains is that of a customer and vendor (“independent parties”). These Terms do not create a partnership, franchise, joint venture, agency, fiduciary, employment, or any other type of relationship.
20. Important notices
a) Adhesion Contracts
By agreeing to these Terms, You are confirming to Us that:
You have had sufficient opportunity to read, review, and consider these Terms.
You understand the content of each paragraph of these Terms.
You have had sufficient opportunity to seek independent professional legal advice.
This means that, to the extent permitted by applicable law, any statutory provisions relating to so-called “form” or “adhesion” contracts do not apply to these Terms.
b) Children and Minors
If You are younger than 18 years old, You cannot agree to these Terms or use Space. By agreeing to these Terms You are confirming that:
either You have legal capacity to enter into these Terms, or You have valid consent from a parent or legal guardian to do so; and
You understand the JetBrains Privacy Policy.
IF YOU DO NOT UNDERSTAND THIS SECTION, DO NOT UNDERSTAND THE JETBRAINS PRIVACY POLICY, OR DO NOT KNOW WHETHER YOU HAVE THE LEGAL CAPACITY TO ACCEPT THESE TERMS, PLEASE ASK YOUR PARENT OR LEGAL GUARDIAN FOR HELP.
If You have any questions about these Terms, please contact Us at sales@jetbrains.com.



## Reporting a bug in Node.js


Report security bugs in Node.js via [HackerOne](https://hackerone.com/nodejs).

Normally your report will be acknowledged within 5 days, and you'll receive
a more detailed response to your report within 10 days indicating the
next steps in handling your submission. These timelines may extend when
our triage volunteers are away on holiday, particularly at the end of the
year.

After the initial reply to your report, the security team will endeavor to keep
you informed of the progress being made towards a fix and full announcement,
and may ask for additional information or guidance surrounding the reported
issue.

### Node.js bug bounty program

The Node.js project engages in an official bug bounty program for security
researchers and responsible public disclosures.  The program is managed through
the HackerOne platform. See <https://hackerone.com/nodejs> for further details.

## Reporting a bug in a third party module

Security bugs in third party modules should be reported to their respective
maintainers.

## Disclosure policy

Here is the security disclosure policy for Node.js

* The security report is received and is assigned a primary handler. This
  person will coordinate the fix and release process. The problem is confirmed
  and a list of all affected versions is determined. Code is audited to find
  any potential similar problems. Fixes are prepared for all releases which are
  still under maintenance. These fixes are not committed to the public
  repository but rather held locally pending the announcement.

* A suggested embargo date for this vulnerability is chosen and a CVE (Common
  Vulnerabilities and Exposures (CVE®)) is requested for the vulnerability.

* On the embargo date, the Node.js security mailing list is sent a copy of the
  announcement. The changes are pushed to the public repository and new builds
  are deployed to nodejs.org. Within 6 hours of the mailing list being
  notified, a copy of the advisory will be published on the Node.js blog.

* Typically the embargo date will be set 72 hours from the time the CVE is
  issued. However, this may vary depending on the severity of the bug or
  difficulty in applying a fix.

* This process can take some time, especially when coordination is required
  with maintainers of other projects. Every effort will be made to handle the
  bug in as timely a manner as possible; however, it's important that we follow
  the release process above to ensure that the disclosure is handled in a
  consistent manner.

## Receiving security updates

Security notifications will be distributed via the following methods.

* <https://groups.google.com/group/nodejs-sec>
* <https://nodejs.org/en/blog/>

## Comments on this policy
cook

#https://github.com/S-pegin/cookies/blob/bot/cookie_update_appinsights/README.md#L1-L103

# Cookies 
  
 GitHub provides a great deal of transparency regarding how we use your data, how we collect your data, and with whom we share your data. To that end, we provide this page which details how we use cookies. 
  
 GitHub uses cookies to provide and secure our websites, as well as to analyze the usage of our websites, in order to offer you a great user experience. Please take a look at our [Privacy Statement](https://docs.github.com/en/github/site-policy/github-privacy-statement#our-use-of-cookies-and-tracking-technologies) if you’d like more information about cookies, and on how and why we use them and cookie-related personal data. You can change your preference about non-essential cookies at any time by following [these instructions](https://docs.github.com/en/account-and-profile/setting-up-and-managing-your-personal-account-on-github/managing-personal-account-settings/managing-your-cookie-preferences-for-githubs-enterprise-marketing-pages). 
  
 Since the number and names of cookies may change, the table below may be updated from time to time. When it is updated, the data of the repo will change. Follow [these instructions](https://github.com/privacy/resources/blob/main/subscribe.md) to subscribe to changes in this repo. 
  
 Provider of Cookie | Cookie Name | Description | Expiration* 
 -----------------|-------------|-------------|------------ 
 GitHub | `app_manifest_token` | This cookie is used during the App Manifest flow to maintain the state of the flow during the redirect to fetch a user session. | Five minutes 
 GitHub | `color_mode` | This cookie is used to indicate the user selected theme preference. | Session 
 GitHub | `_device_id` | This cookie is used to track recognized devices for security purposes. | One year 
 GitHub | `dotcom_user` | This cookie is used to signal to us that the user is already logged in. | One year 
 GitHub | `ghcc` | This cookie validates user's choice about cookies | 180 Days 
 GitHub | `_gh_ent` | This cookie is used for temporary application and framework state between pages like what step the customer is on in a multiple step form. | Two weeks 
 GitHub | `_gh_sess` | This cookie is used for temporary application and framework state between pages like what step the user is on in a multiple step form. | Session 
 GitHub | `gist_oauth_csrf` | This cookie is set by Gist to ensure the user that started the oauth flow is the same user that completes it. | Deleted when oauth state is validated 
 GitHub | `gist_user_session` | This cookie is used by Gist when running on a separate host. | Two weeks 
 GitHub | `has_recent_activity` | This cookie is used to prevent showing the security interstitial to users that have visited the app recently. | One hour 
 GitHub | `__Host-gist_user_session_same_site` | This cookie is set to ensure that browsers that support SameSite cookies can check to see if a request originates from GitHub. | Two weeks 
 GitHub | `__Host-user_session_same_site` | This cookie is set to ensure that browsers that support SameSite cookies can check to see if a request originates from GitHub. | Two weeks 
 GitHub | `logged_in` | This cookie is used to signal to us that the user is already logged in. | One year 
 GitHub | `marketplace_repository_ids` | This cookie is used for the marketplace installation flow. | One hour 
 GitHub | `marketplace_suggested_target_id` | This cookie is used for the marketplace installation flow. | One hour 
 GitHub | `_octo` | This cookie is used for session management including caching of dynamic content, conditional feature access, support request metadata, and first party analytics. | One year 
 GitHub | `org_transform_notice` | This cookie is used to provide notice during organization transforms. | One hour 
 GitHub | `private_mode_user_session` | This cookie is used for Enterprise authentication requests. | Two weeks 
 GitHub | `saml_csrf_token` | This cookie is set by SAML auth path method to associate a token with the client. | Until user closes browser or completes authentication request 
 GitHub | `saml_csrf_token_legacy` | This cookie is set by SAML auth path method to associate a token with the client. | Until user closes browser or completes authentication request 
 GitHub | `saml_return_to` | This cookie is set by the SAML auth path method to maintain state during the SAML authentication loop. | Until user closes browser or completes authentication request 
 GitHub | `saml_return_to_legacy` | This cookie is set by the SAML auth path method to maintain state during the SAML authentication loop. | Until user closes browser or completes authentication request 
 GitHub | `show_cookie_banner` | Set based on the client’s region and used to determine if a cookie consent banner should be shown | Session 
 GitHub | `tz` | This cookie allows us to customize timestamps to your time zone. | Session 
 GitHub | `user_session` | This cookie is used to log you in. | Two weeks 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `ai_session` | Application Insights session ID | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `ai_user` | Application Insights user ID | 30 minutes 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `ANONCHK` | This Microsoft Clarity cookie monitors website performance | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `isFirstSession` | This cookie is used when user opts-in to saving information | Session 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `MSO` | This cookie identifies a session | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `MC1` | This cookie is used for advertising, site analytics, and other operational purposes | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `MR` | This cookie checks whether to extend the lifetime of the MUID cookie | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `MSFPC` | This cookie is used for advertising, site analytics, and other operational purposes | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `MUID` | This cookie stores Bing’s visitor ID. This cookie is used for advertising, site analytics, and other operational purposes | 13 months 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `SM` | This cookie is used in synchronizing the MUID across Microsoft domains. | Session 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `_uetsid` | This cookie is used for analytics to store and track visits across sites | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `_uetvid` | This cookie is used by Bing Ads to store and track visits across websites | 13 months 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `X-FD-FEATURES` | This cookie is used for tracking analytics and evenly spreading load on the website | One year 
 [Microsoft](https://privacy.microsoft.com/en-us/privacystatement) | `X-FD-Time` | This cookie is used for tracking analytics and evenly spreading load on website | One year 
 [Adobe](https://www.adobe.com/privacy/policy.html) | `aam_uuid` | This cookie is an audience manager | 13 months 
 [Adobe](https://www.adobe.com/privacy/policy.html) | `mboxEdgeCluster` | This cookie is used by Adobe Target load balancer. Adobe Target is used to determine which targeted content to display to visitor | 13 months 
 [Adobe](https://www.adobe.com/privacy/policy.html) | `AMCV_EA76ADE95776D2EC7F000101%40AdobeOrg` | Adobe cookie used to track and analyze user activities on the website | 13 months 
 [Adobe](https://www.adobe.com/privacy/policy.html) | `AMCVS_EA76ADE95776D2EC7F000101%40AdobeOrg` | Adobe cookie used to track and analyze user activities on the website | Session 
 [Adobe](https://www.adobe.com/privacy/policy.html) | `at_check` | Adobe Target to support conversion tracking for new product customers | Session 
 [Adobe](https://www.adobe.com/privacy/policy.html) | `mbox` | Adobe Target to store session ID | 13 months 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `_cs_c` | Consent state: digit between 0 and 3. Used for capturing analytics on web pages | 13 months 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `_cs_cvars` | This cookie is used to capture analytics on the web page | Session 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `_cs_id` | Contains: user ID, timestamp (in seconds) of user creation, number of visits for this user | 13 months 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `_cs_s` | Number of page views for the current session, and the recording state | One year 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `__CT_Data` | This cookie is used to count the number of a guest’s pageviews or visits | One year 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `_CT_RS_` | This cookie is used to capture analytics on the web page | One year 
 [Contentsquare](https://go.contentsquare.com/en/tracking-tag-cookies) | `WRUID` | This cookie is used for analytics | One year 
 [Facebook](https://www.facebook.com/policies/cookies/) | _fbc | This cookie is used to personalize content (including ads), measure ads, produce analytics, and provide a safer experience. | 90 Days 
 [Facebook](https://www.facebook.com/policies/cookies/) | _fbp | This cookie is used to personalize content (including ads), measure ads, produce analytics, and provide a safer experience. | 90 Days 
 [Facebook](https://www.facebook.com/policies/cookies/) | fr | This cookie is used as the primary advertising cookie used to deliver, measure, and improve the relevancy of ads. | 90 Days 
 [Facebook](https://www.facebook.com/policies/cookies/) | wd | This cookie is used to deliver an optimal experience for your device’s screen. | 7 Days 
 [Facebook](https://www.facebook.com/policies/cookies/) | oo | This cookie is an opt out cookie set by a user visiting Digital Advertising Alliance and choosing to opt out. | 5 years 
 [Google](https://policies.google.com/privacy) | _gcl_au | This cookie is used by Google AdSense for experimenting with advertisement efficiency across websites using their services. | 90 Days 
 [Google](https://policies.google.com/privacy) | id | This cookie is used to build a profile of the website visitor's interests and show relevant ads on other sites. | "OPT_OUT: fixed expiration (year 2030/11/09); non-OPT_OUT: 13 months EEA UK 24 months elsewhere" 
 [Google](https://policies.google.com/privacy) | IDE | This cookie is used to build a profile of the website visitor's interests and show relevant ads on other sites. | "13 months EEA UK; 24 months elsewhere" 
 [Google](https://policies.google.com/privacy) | lsid | This cookie is used to provide information about how the end user uses the website and any advertising that the end user may have seen before visiting the website. | 90 Days 
 [Google](https://policies.google.com/privacy) | NID | This cookie is used to build a profile of the website visitor's interests and show relevant ads on other sites. | 90 Days 
 [Google](https://policies.google.com/privacy) | PREF | This cookie is used to build a profile of the website visitor's interests and show relevant ads on other sites. | 90 Days 
 [Google](https://policies.google.com/privacy) | SSID | This cookie is used to provide information about how the end user uses the website and any advertising that the end user may have seen before visiting the website. | 90 Days 
 [Google](https://policies.google.com/privacy) | SAPISID | This cookie is used to build a profile of the website visitor's interests and show relevant ads on other sites. | 90 Days 
 [Google](https://policies.google.com/privacy) | test_cookie | This cookie is used to determine if the website visitor's browser supports cookies. | 90 Days 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | bcookie | This cookie is a browser identifier cookie to uniquely identify devices accessing LinkedIn to detect abuse on the platform. | One year 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | bscookie | This cookie is used for remembering that a logged in user is verified by two factor authentication. | One year 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | u | This cookie is used to provide a platform to enable advertisers to track users across multiple devices. | 3 months 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | UserMatchHistory | This cookie is used to track visitors so that more relevant ads can be presented based on the visitor's preferences. | One month 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | JSESSIONID | This cookie is used for Cross Site Request Forgery (CSRF) protection. | Session 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | lang | This cookie is used to remember a user's language setting to ensure LinkedIn.com displays in the language selected by the user in their settings. | Session 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | lidc | This cookie is used to faciliatate data center selection | 24 hours 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | sdsc | This cookie is used for database routing to ensure consistency across all databases when a change is made and to ensure that user-inputted content is immediately available to the submitting user upon submission. | Session 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | li_gc | This cookie is used to store consent of visitors regarding the use of cookies for non-essential purposes. | 6 months 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | li_mc | This cookie is used as a temporary cache to avoid database lookups for a member's consent for use of non-essential cookies and used for having consent information on the client side to enforce consent on the client side. | 6 months 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | AnalyticsSyncHistory | This cookie is used to store information about the time a sync took place with the lms_analytics cookie. | 30 Days 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | lms_ads | This cookie is used to identify LinkedIn Members off LinkedIn for advertising. | 30 Days 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | lms_analytics | This cookie is used to identify LinkedIn Members off LinkedIn for analytics. | 30 Days 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | li_fat_id | This cookie is used for conversion tracking, retargeting, analytics. | 30 Days 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | li_sugr | This cookie is used to make a probabilistic match of a user's identity. | 90 Days 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | U | This cookie is used as a browser identifier. | 3 months 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | BizographicsOptOutBizographicsOptOut | This cookie is used to determine opt-out status for non-members. | 10 years 
 [LinkedIn](https://www.linkedin.com/legal/privacy-policy) | li_giant | This cookie is used for conversion tracking. | 7 Days | https://www.linkedin.com/legal/privacy-policy 
 [Quantcast](https://www.quantcast.com/privacy/) | cref | This cookie is used for Market and Audience Segmentation and Targeted advertising services. | 13 months 
 [Quantcast](https://www.quantcast.com/privacy/) | d | This cookie is used for Market and Audience Segmentation and Targeted advertising services. | 3 months 
 [Quantcast](https://www.quantcast.com/privacy/) | mc | This cookie is used to track anonymous information about how website visitors use the site. | 13 months 
 [Yahoo](https://policies.yahoo.com/us/en/yahoo/privacy/index.htm?redirect=no) | A3 | This cookie is used for search and advertising. | One year         
 [Yahoo](https://policies.yahoo.com/us/en/yahoo/privacy/index.htm?redirect=no) | b | This cookie collects anonymous data related to the visitor's website visits, such as the number of visits, average time spent on the website and what pages have been loaded. The registered data is used to categorize the users' interest and demographical profiles with the purpose of customizing the website content depending on the visitor. | One year 
  
 (*) The expiration dates for the cookies listed above generally apply on a rolling basis.  
   
 ⚠️ Please note while we limit our use of third party cookies to those necessary to provide external functionality when rendering external content, certain pages on our website may set other third party cookies. For example, we may embed content, such as videos, from another site that sets a cookie. While we try to minimize these third party cookies, we can’t always control what cookies this third party content   / 
If you have suggestions on how this process could be improved please submit a
[pull request](https://github.com/nodejs/nodejs.org) or
[file an issue](https://github.com/nodejs/security-wg/issues/new) to discuss./
