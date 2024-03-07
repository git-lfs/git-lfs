#!/usr/bin/env ruby
# Pushes all deb and rpm files from ./repos to PackageCloud.

packagecloud_user = ENV["PACKAGECLOUD_USER"] || "github"
packagecloud_token = ENV["PACKAGECLOUD_TOKEN"] || begin
  puts "PACKAGECLOUD_TOKEN env required"
  exit 1
end

require "json"
require_relative 'lib/distro'

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
$distro_name_map = DistroMap.new.distro_name_map

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
