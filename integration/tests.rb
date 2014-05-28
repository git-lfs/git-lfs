require File.expand_path("../suite", __FILE__)
config = Suite.config

Suite.test :empty do |t|
  t.repository File.join(t.path, ".git")
  t.repository File.join(t.path, "subdir") # sub directory!

  # really simple test
  t.command "version",
    "git-media v#{config.version}"

  # test against a longer expected output
  t.command "version -comics",
    <<-END
git-media v#{config.version}
Nothing may see Gah Lak Tus and survive.
    END

  t.command "config",
    <<-END
Endpoint=https://example.com/git/media.git/info/media
LocalWorkingDir=#{File.join "/private", t.path}
LocalGitDir=#{File.join "/private", t.path, ".git"}
LocalMediaDir=#{File.join "/private", t.path, ".git", "media"}
TempDir=#{File.join config.tmp, "git-media"}
#{config.env_string}
    END

  # make some other checks besides just the command's output
  t.command "init" do |cmd|
    cmd.expected = "Installing clean filter
Installing smudge filter
git media initialized"

    cmd.after do
      gitconfig = Suite.global_git_config
      if gitconfig.select { |l| l == "filter.media.clean=git media clean %f" }.size != 1
        next "bad filter.media.clean configs"
      end

      if gitconfig.select { |l| l == "filter.media.smudge=git media smudge %f" }.size != 1
        next "bad filter.media.smudge configs"
      end

      if gitconfig.select { |l| l =~ /\Afilter\.media\./ }.size != 2
        next "bad filter.media configs"
      end
    end
  end
end

Suite.test :config_media_url do |t|
  t.command "config",
    <<-END
Endpoint=http://foo/bar
LocalWorkingDir=#{File.join "/private", t.path}
LocalGitDir=#{File.join "/private", t.path, ".git"}
LocalMediaDir=#{File.join "/private", t.path, ".git", "media"}
TempDir=#{File.join config.tmp, "git-media"}
#{config.env_string}
    END
end

Suite.run!
