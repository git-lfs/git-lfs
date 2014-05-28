require File.expand_path("../suite", __FILE__)
config = Suite.config

Suite.test config.root do |t|
  t.repository File.join(config.root, "integration") # sub directory!

  t.command "version",
    "git-media v#{config.version}"

  t.command "version -comics",
    <<-END
git-media v#{config.version}
Nothing may see Gah Lak Tus and survive.
    END

  t.command "config",
    <<-END
Endpoint=https://github.com/github/git-media.git/info/media
LocalWorkingDir=#{config.root}
LocalGitDir=#{File.join config.root, ".git"}
LocalMediaDir=#{File.join config.root, ".git", "media"}
TempDir=#{File.join config.tmp, "git-media"}
#{config.env_string}
    END
end

Suite.run!
