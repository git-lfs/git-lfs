require File.expand_path("../suite", __FILE__)

Suite.test :empty do |t|
  t.add_path File.join(t.path, ".git")
  t.add_path File.join(t.path, "subdir") # sub directory!

  # really simple test
  t.command "version",
    "git-media v#{Suite.version}"

  # test against a longer expected output
  t.command "version -comics",
    <<-END
git-media v#{Suite.version}
Nothing may see Gah Lak Tus and survive.
    END

  t.command "config",
    <<-END
Endpoint=https://example.com/git/media.git/info/media
LocalWorkingDir=#{t.path}
LocalGitDir=#{File.join t.path, ".git"}
LocalMediaDir=#{File.join t.path, ".git", "media"}
TempDir=#{File.join Suite.tmp, "git-media"}
#{Suite.env_string}
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

Suite.test :empty do |t|
  # add a path, make sure its written to .gitattributes, and check `git media path`
  t.command "path add *.gif", "Adding path *.gif" do |cmd|
    cmd.after do
      if t.read(".gitattributes") != "*.gif filter=media -crlf"
        next ".gitattributes not set"
      end

      actual = t.exec("path")
      expected = "Listing paths
    *.gif (.gitattributes)"
      if actual != expected
        next ".git path not shown by 'git media path':\n#{actual}"
      end
    end
  end

  # test the default path output
  t.command "path", "Listing paths"
end

Suite.test :config_media_url do |t|
  t.command "config",
    <<-END
Endpoint=http://foo/bar
LocalWorkingDir=#{t.path}
LocalGitDir=#{File.join t.path, ".git"}
LocalMediaDir=#{File.join t.path, ".git", "media"}
TempDir=#{File.join Suite.tmp, "git-media"}
#{Suite.env_string}
    END
end

Suite.test :attributes do |t|
  output =<<-END
Listing paths
    *.mov (.git/info/attributes)
    *.jpg (.gitattributes)
    *.gif (a/.gitattributes)
    *.png (a/b/.gitattributes)
    END

  t.command "path", output do |cmd|
    cmd.before do
      t.write("*.mov filter=media -crlf\n", ".git", "info", "attributes")
    end
  end
end

Suite.run!
