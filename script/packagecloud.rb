#!/usr/bin/env ruby
# Pushes all deb and rpm files from ./repos to PackageCloud.

packagecloud_user = ENV["PACKAGECLOUD_USER"] || "github"
packagecloud_token = ENV["PACKAGECLOUD_TOKEN"] || begin
  puts "PACKAGECLOUD_TOKEN env required"
  exit 1
end

require "json"

packagecloud_ruby_minimum_version = "1.0.4"
begin
  gem "packagecloud-ruby", ">=#{packagecloud_ruby_minimum_version}"
  require "packagecloud"
  puts "Using packagecloud-ruby:#{Gem.loaded_specs["packagecloud-ruby"].version}"
rescue LoadError
  puts "Requires packagecloud-ruby >=#{packagecloud_ruby_minimum_version}"
  puts %(gem install packagecloud-ruby)
  exit 1
end

credentials = Packagecloud::Credentials.new(packagecloud_user, packagecloud_token)
$client = Packagecloud::Client.new(credentials)

# matches package directories built by docker to one or more packagecloud distros
# https://packagecloud.io/docs#os_distro_version
#
# If you change the keys of this list, change the values below, as well as
# script/upload and docker/run_dockers.bsh.
$distro_name_map = {
  # RHEL EOL https://access.redhat.com/support/policy/updates/errata
  "centos/7" => [
    "el/7",         # EOL June 2024
    "scientific/7", # EOL June 2024
    # opensuse https://en.opensuse.org/Lifetime
    # or https://en.wikipedia.org/wiki/OpenSUSE_version_history
    "opensuse/15.4", # EOL November 2023
    # SLES EOL https://www.suse.com/lifecycle/
    "sles/12.5", # EOL October 2024 (LTSS October 2027)
    "sles/15.4", # Current
  ],
  "centos/8" => [
    "el/8",
  ],
  "rocky/9" => [
    "el/9",
    "fedora/37", # EOL November 2023
    "fedora/38", # EOL May 2024
  ],
  # Debian EOL https://wiki.debian.org/LTS/
  # Ubuntu EOL https://wiki.ubuntu.com/Releases
  # Mint EOL https://linuxmint.com/download_all.php
  "debian/10" => [
    "debian/buster",    # EOL June 2024
    "linuxmint/ulyana", # EOL April 2025
    "linuxmint/ulyssa", # EOL April 2025
    "linuxmint/uma",    # EOL April 2025
    "linuxmint/una",    # EOL April 2025
    "ubuntu/focal",     # EOL April 2025
  ],
  "debian/11" => [
    "debian/bullseye",  # EOL June 2026
    "ubuntu/jammy",     # EOL April 2027
    "ubuntu/kinetic",   # EOL July 2023
    "ubuntu/lunar",     # EOL January 2024
    "linuxmint/vanessa",# EOL April 2027
    "linuxmint/vera",   # EOL April 2027
  ],
  "debian/12" => [
    "debian/bookworm",  # Current stable
    "debian/trixie",    # Current testing
  ]
}

# caches distro id lookups
$distro_id_map = {}

def distro_names_for(filename)
  $distro_name_map.each do |pattern, distros|
    return distros if filename.include?(pattern)
  end

  raise "no distro for #{filename.inspect}"
end

package_files = Dir.glob("repos/**/*.rpm") + Dir.glob("repos/**/*.deb")
package_files.each do |full_path|
  next if full_path =~ /repo-release/
  pkg = Packagecloud::Package.new(:file => full_path)
  distro_names = distro_names_for(full_path)
  distro_names.map do |distro_name|
    distro_id = $distro_id_map[distro_name] ||= $client.find_distribution_id(distro_name)
    if !distro_id
      raise "no distro id for #{distro_name.inspect}"
    end

    puts "pushing #{full_path} to #{$distro_id_map.key(distro_id).inspect}"
    result = $client.put_package("git-lfs", pkg, distro_id)
    result.succeeded || begin
      # We've already uploaded this package in an earlier invocation of this
      # script and our attempt to upload over the existing package failed
      # because PackageCloud doesn't allow that. Ignore the failure since we
      # already have the package uploaded.
      if result.response != '{"filename":["has already been taken"]}'
        raise "packagecloud put_package failed, error: #{result.response}"
      end
    end
  end
end

package_files.each do |full_path|
  next if full_path.include?("SRPM") || full_path.include?("i386") || full_path.include?("i686")
  next unless full_path =~ /\/git-lfs[-|_]\d/
  # If you change the entries below, change the keys above, as well as
  # script/upload and docker/run_dockers.bsh.
  os, distro = case full_path
  when /debian\/10/ then ["Debian 10", "debian/buster"]
  when /debian\/11/ then ["Debian 11", "debian/bullseye"]
  when /debian\/12/ then ["Debian 12", "debian/bookworm"]
  when /centos\/7/  then ["RPM RHEL 7/CentOS 7", "el/7"]
  when /centos\/8/  then ["RPM RHEL 8/CentOS 8", "el/8"]
  when /rocky\/9/  then ["RPM RHEL 9/Rocky Linux 9", "el/9"]
  end

  next unless os

  puts "[#{os}](https://packagecloud.io/#{packagecloud_user}/git-lfs/packages/#{distro}/#{File.basename(full_path)}/download)"
end
