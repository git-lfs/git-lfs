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
  "centos/5" => %w(
    el/5
  ),
  "centos/6" => %w(
    el/6
  ),
  "centos/7" => %w(
    el/7
    fedora/22
    fedora/23
  ),
  "debian/7" => %w(
    debian/wheezy
    ubuntu/precise
  ),
  "debian/8" => %w(
    debian/jessie
    linuxmint/sarah
    linuxmint/rebecca
    linuxmint/rafaela
    linuxmint/rosa
    ubuntu/trusty
    ubuntu/vivid
    ubuntu/wily
    ubuntu/xenial
  ),
  "debian/9" => %w(
    debian/stretch
  )
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
      raise "packagecloud put_package failed, error: #{result.response}"
    end
  end
end

package_files.each do |full_path|
  next if full_path.include?("SRPM") || full_path.include?("i386") || full_path.include?("i686")
  next unless full_path =~ /\/git-lfs[-|_]\d/
  os, distro = case full_path
  when /debian\/7/ then ["Debian 7", "debian/wheezy"]
  when /debian\/8/ then ["Debian 8", "debian/jessie"]
  when /debian\/9/ then ["Debian 9", "debian/stretch"]
  when /centos\/5/ then ["RPM RHEL 5/CentOS 5", "el/5"]
  when /centos\/6/ then ["RPM RHEL 6/CentOS 6", "el/6"]
  when /centos\/7/ then ["RPM RHEL 7/CentOS 7", "el/7"]
  end

  next unless os

  puts "[#{os}](https://packagecloud.io/#{packagecloud_user}/git-lfs/packages/#{distro}/#{File.basename(full_path)}/download)"
end
