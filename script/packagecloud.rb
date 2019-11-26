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
$distro_name_map = {
  # RHEL EOL https://access.redhat.com/support/policy/updates/errata
  "centos/5" => [
    "el/5" # End of Extended Support November 30, 2020
  ],
  "centos/6" => [
    "el/6", # End of Extended Support June 30, 2024
    "scientific/6",
  ],
  "centos/7" => [
    "el/7",
    "scientific/7",
    #"el/8", # BOL ~2019-2020?
    # Fedora EOL check https://fedoraproject.org/wiki/End_of_life
    # or https://en.wikipedia.org/wiki/Fedora_version_history#Version_history
    "fedora/28", # EOL ~Oct 2019
    "fedora/29", # EOL ~2020
    "fedora/30", # EOL ~2020
    # "fedora/30", # BOL ~May 2019
    # opensuse https://en.opensuse.org/Lifetime
    # or https://en.wikipedia.org/wiki/OpenSUSE_version_history
    "opensuse/42.3", # EOL 2019-06-30
    "opensuse/15.0", # EOL 2019-11-25
    "opensuse/15.1", # EOL 2020-11
    # SLES EOL https://www.suse.com/lifecycle/
    "sles/11.4", # LTSS ends 31 Mar 2022
    "sles/12.0", # LTSS ends 01 July 2019
    "sles/12.1", # LTSS ends 31 May 2020
    "sles/12.2", # LTSS ends 31 Mar 2021
    "sles/12.3", # LTSS ends 30 Jun 2022
    "sles/15.0"  # Current
  ],
  "centos/8" => [
    "el/8",
  ],
  # Debian EOL https://wiki.debian.org/LTS/
  # Ubuntu EOL https://wiki.ubuntu.com/Releases
  # Mint EOL https://linuxmint.com/download_all.php
  "debian/8" => [
    "debian/jessie",     # EOL June 30, 2020
    "linuxmint/qiana",   # EOL April 2019
    "linuxmint/rafaela", # EOL April 2019
    "linuxmint/rebecca", # EOL April 2019
    "linuxmint/rosa",    # EOL April 2019
    "ubuntu/trusty",     # ESM April 2022
    "ubuntu/vivid",      # EOL February 4, 2016
    "ubuntu/wily"        # EOL July 28, 2016
  ],
  "debian/9" => [
    "debian/stretch",   # EOL June 2022
    "linuxmint/sarah",  # EOL April 2021
    "linuxmint/serena", # EOL April 2021
    "linuxmint/sonya",  # EOL April 2021
    "linuxmint/sylvia", # EOL April 2021
    "linuxmint/tara",   # EOL April 2023
    "linuxmint/tessa",  # EOL April 2023
    "ubuntu/xenial",    # ESM April 2024
    "ubuntu/yakkety",   # EOL July 20, 2017
    "ubuntu/zesty",     # EOL January 13, 2018
    "ubuntu/artful",    # EOL July 19 2018
    "ubuntu/bionic",    # ESM April 2028
    "ubuntu/cosmic",    # EOL July 2019
    "ubuntu/disco",     # EOL April 2020
  ],
  "debian/10" => [
    "debian/buster",    # Current
    "ubuntu/eoan",      # Current
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
  os, distro = case full_path
  when /debian\/8/  then ["Debian 8",  "debian/jessie"]
  when /debian\/9/  then ["Debian 9",  "debian/stretch"]
  when /debian\/10/ then ["Debian 10", "debian/buster"]
  when /centos\/5/  then ["RPM RHEL 5/CentOS 5", "el/5"]
  when /centos\/6/  then ["RPM RHEL 6/CentOS 6", "el/6"]
  when /centos\/7/  then ["RPM RHEL 7/CentOS 7", "el/7"]
  end

  next unless os

  puts "[#{os}](https://packagecloud.io/#{packagecloud_user}/git-lfs/packages/#{distro}/#{File.basename(full_path)}/download)"
end
