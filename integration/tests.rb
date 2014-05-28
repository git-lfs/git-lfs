require File.expand_path("../suite", __FILE__)
config = Suite.config

Suite.test config.root do |t|
  t.repository File.join(config.root, "integration") # sub directory!

  t.command "version" do
    "git-media v#{config.version}"
  end

  t.command "version -comics" do
    <<-END
git-media v#{config.version}
Nothing may see Gah Lak Tus and survive.
    END
  end

  t.command "config" do
    <<-END
Endpoint=https://github.com/github/git-media.git/info/media
LocalWorkingDir=#{config.root}
LocalGitDir=#{File.join config.root, ".git"}
LocalMediaDir=#{File.join config.root, ".git", "media"}
TempDir=#{File.join config.tmp, "git-media"}
#{config.env_string}
    END
  end
end

Suite.run!
