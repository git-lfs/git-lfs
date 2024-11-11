#!/usr/bin/env ruby

require 'optionparser'

class DistroMap
  attr_reader :entries

  def initialize(map = nil)
    @entries = map || self.class.builtin_map
  end
  # Returns the map for our distros.
  #
  # The key in each case is a string containing a lowercase OS name, a slash,
  # and a version number.  The value is a map containing the following fields:
  #
  # name:: a human-readable name for this distro.
  # component:: a component suitable for a packagecloud.io URL.
  # image:: a Docker image name from build_dockers without any extension.
  # equivalent:: packagecloud.io components for which we can upload the same
  # package.
  # package_type:: the extension for the package format on this OS.
  # package_tag:: the trailing component after the version number on this OS.
  def self.builtin_map
    {
      # RHEL EOL https://access.redhat.com/support/policy/updates/errata
      # Fedora EOL https://docs.fedoraproject.org/en-US/releases/
      # SLES EOL https://www.suse.com/lifecycle/
      # opensuse https://en.opensuse.org/Lifetime
      # or https://en.wikipedia.org/wiki/OpenSUSE_version_history
      "centos/7" => {
        name: "RPM RHEL 7/CentOS 7",
        component: "el/7",
        image: "centos_7",
        package_type: "rpm",
        package_tag: "-1.el7",
        equivalent: [
          "el/7",                      # EOL June 2024
          "sles/12.5",                 # EOL October 2024
        ],
      },
      "centos/8" => {
        name: "RPM RHEL 8/Rocky Linux 8",
        component: "el/8",
        image: "centos_8",
        package_type: "rpm",
        package_tag: "-1.el8",
        equivalent: [
          "el/8",                      # EOL May 2029
          "opensuse/15.5",             # EOL December 2024
          "sles/15.5",                 # EOL December 2024
        ],
      },
      "rocky/9" => {
        name: "RPM RHEL 9/Rocky Linux 9",
        component: "el/9",
        image: "rocky_9",
        package_type: "rpm",
        package_tag: "-1.el9",
        equivalent: [
          "el/9",                      # EOL May 2032
          "fedora/39",                 # EOL November 2024
          "fedora/40",                 # EOL May 2025
          "fedora/41",                 # EOL November 2025
          "opensuse/15.6",             # EOL December 2025
          "sles/15.6",                 # Current
        ],
      },
      # Debian EOL https://wiki.debian.org/LTS/
      # Ubuntu EOL https://wiki.ubuntu.com/Releases
      # Mint EOL https://linuxmint.com/download_all.php
      "debian/10" => {
        name: "Debian 10",
        component: "debian/buster",
        image: "debian_10",
        package_type: "deb",
        package_tag: "",
        equivalent: [
          "debian/buster",             # EOL June 2024
          "linuxmint/ulyana",          # EOL April 2025
          "linuxmint/ulyssa",          # EOL April 2025
          "linuxmint/uma",             # EOL April 2025
          "linuxmint/una",             # EOL April 2025
          "ubuntu/focal",              # EOL April 2025
        ],
      },
      "debian/11" => {
        name: "Debian 11",
        component: "debian/bullseye",
        image: "debian_11",
        package_type: "deb",
        package_tag: "",
        equivalent: [
          "debian/bullseye",           # EOL August 2026
          "linuxmint/vanessa",         # EOL April 2027
          "linuxmint/vera",            # EOL April 2027
          "linuxmint/victoria",        # EOL April 2027
          "linuxmint/virginia",        # EOL April 2027
          "ubuntu/jammy",              # EOL April 2027
        ],
      },
      "debian/12" => {
        name: "Debian 12",
        component: "debian/bookworm",
        image: "debian_12",
        package_type: "deb",
        package_tag: "",
        equivalent: [
          "debian/bookworm",           # EOL June 2028
          "debian/trixie",             # Current testing (Debian 13)
          "linuxmint/wilma",           # EOL April 2029
          "ubuntu/noble",              # EOL June 2029
          "ubuntu/oracular",           # EOL July 2025
        ]
      },
    }
  end

  def distro_name_map
    entries.map { |k, v| [k, v[:equivalent]] }.to_h
  end

  def image_names
    entries.values.map { |v| v[:image] }.to_a
  end
end

class DistroMapProgram
  def initialize(stdout, stderr, dmap = nil)
    @dmap = DistroMap.new(dmap)
    @stdout = stdout
    @stderr = stderr
  end

  def image_names
    @stdout.puts @dmap.image_names.join(" ")
  end

  def distro_markdown
    arch = {
      "rpm" => ".x86_64",
      "deb" => "_amd64",
    }
    separator = {
      "rpm" => "-",
      "deb" => "_",
    }
    result = @dmap.entries.map do |_k, v|
      type = v[:package_type]
      "[#{v[:name]}](https://packagecloud.io/github/git-lfs/packages/#{v[:component]}/git-lfs#{separator[type]}VERSION#{v[:package_tag]}#{arch[type]}.#{type}/download)\n"
    end.join
    @stdout.puts result
  end

  def run(args)
    options = {}
    OptionParser.new do |parser|
      parser.on("--image-names", "Print the names of all images") do
        options[:mode] = :image_names
      end

      parser.on("--distro-markdown", "Print links to packages for all distros") do
        options[:mode] = :distro_markdown
      end
    end.parse!(args)

    case options[:mode]
    when nil
      @stderr.puts "A mode option is required"
      2
    when :image_names
      image_names
      0
    when :distro_markdown
      distro_markdown
      0
    end
  end
end

if $PROGRAM_NAME == __FILE__
  exit DistroMapProgram.new($stdout, $stderr).run(ARGV)
end
