# Pushes all deb and rpm files from ./repos to PackageCloud.

packagecloud_user = ENV["PACKAGECLOUD_USER"] || "github"
packagecloud_token = ENV["PACKAGECLOUD_TOKEN"] || begin
  puts "PACKAGECLOUD_TOKEN env required"
  exit 1
end

begin
  require "packagecloud"
rescue LoadError
  puts %(gem install packagecloud-ruby)
  exit 1
end

credentials = Packagecloud::Credentials.new(packagecloud_user, packagecloud_token)
$client = Packagecloud::Client.new(credentials)

# matches package directories to packagecloud distro names
$distro_name_map = {
  "centos/5" => "el/5",
  "centos/6" => "el/6",
  "centos/7" => "el/7",
  "debian/7" => "debian/wheezy",
  "debian/8" => "debian/jessie",
}

# caches distro id lookups
$distro_id_map = {}

def distro_name_for(filename)
  $distro_name_map.each do |pattern, name|
    return name if filename.include?(pattern)
  end

  raise "no distro for #{filename.inspect}"
end

def build_package(filename)
  distro_name = distro_name_for(filename)
  distro_id = $distro_id_map[distro_name] ||= $client.find_distribution_id(distro_name)
  Packagecloud::Package.new(open(filename), distro_id)
end

packages = Dir.glob("repos/**/*.rpm") + Dir.glob("repos/**/*.deb")
packages.each do |full_path|
  next if full_path =~ /repo-release/
  puts "pushing #{full_path}"
  $client.put_package("git-lfs", build_package(full_path))
end
