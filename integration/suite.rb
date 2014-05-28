require "tmpdir"
require "fileutils"

class Suite
  class Config < Struct.new(:root)
    def bin
      @bin ||= File.join(root, "bin/git-media")
    end

    def version
      @version ||= cmd(:version)
    end

    def tmp
      @tmp ||= Dir.tmpdir
    end

    # Gets existing GIT_* env vars
    def env
      @env ||= ENV.inject({}) do |memo, (k, v)|
        if k =~ /\AGIT_/
          memo.update k => v
        else
          memo
        end
      end
    end

    def env_string
      env.inject [] do |memo, (key, value)|
        memo << "#{key}=#{value}"
      end.join("\n")
    end

  private
    def cmd(file)
      %x{go run #{root}/integration/#{file}.go}.strip
    end
  end

  def self.config
    @config ||= Config.new(File.expand_path("../..", __FILE__))
  end

  def self.tests
    @tests ||= []
  end

  def self.test_tmpdir
    @test_tmpdir ||= begin
      tmp = File.join(config.tmp, "git-media-tests")
      FileUtils.rm_rf(tmp)
      tmp
    end
  end

  # Gets the global git configuration.
  def self.global_git_config
    `git config -l --global`.strip.split("\n")
  end

  def self.test(repo_name)
    t = Test.new(repo_name)
    yield t if block_given?
    tests << t
  end

  class Test
    attr_reader :path

    def initialize(name)
      @repository_name = name
      @path = expand(File.join(Suite.test_tmpdir, name.to_s))
      @repositories = [@path]
      @commands = []
      @successful = true
    end

    def exist?(*relative_parts)
      File.exist?(File.join(@path, *relative_parts))
    end

    def read(*relative_parts)
      return nil unless exist?(*relative_parts)
      IO.read(File.join(@path, *relative_parts)).to_s.strip
    end

    def exec(cmd)
      %x{#{Suite.config.bin} #{cmd}}.strip
    end

    def failed?
      !@successful
    end

    def repository(path)
      @repositories << path
    end

    def command(cmd, output = nil)
      c = Command.new(cmd, output)
      yield c if block_given?
      @commands << c
    end

    def run!
      @repositories.each { |r| run(r) }
    end

    def run(r)
      puts "Integration tests for #{r}"
      puts
      @commands.each do |c|
        clone(r) do
          @successful = false unless c.run!(r)
        end
      end
      puts
    end

  private
    def clone(path)
      FileUtils.rm_rf @path
      Dir.chdir File.join(Suite.config.root, "integration", "repos") do
        %x{git clone #{@repository_name} #{@path} 2> /dev/null}
        # set a default origin remote for each test case
        Dir.chdir @path do
          `git remote remove origin 2> /dev/null`
          `git remote add origin https://example.com/git/media 2> /dev/null`
        end
      end

      Dir.chdir(path) do
        yield
      end
    end

    # expands the /var path which gets symlinked to "private/var" on OSX.
    def expand(path)
      pieces = path.split "/"
      pieces.shift
      expanded = ""
      pieces.each do |part|
        trial = File.join(expanded, part)
        expanded = if File.symlink?(trial)
          File.readlink(trial)
        else
          trial
        end
      end

      if expanded.start_with?("/")
        expanded
      else
        File.join("", expanded)
      end
    end
  end

  class Command
    attr_accessor :expected

    def initialize(cmd, expected, &block)
      @cmd = cmd
      @expected = expected.strip if expected
      @after = block
    end

    def after(&block)
      @after = block
    end

    def run!(repository)
      puts "$ git media #{@cmd}"
      actual = %x{#{Suite.config.bin} #{@cmd}}.strip

      if @expected && @expected != actual
        puts "- expected"
        puts @expected
        puts
        puts "- actual"
        puts actual
        puts

        return false
      end

      if err = @after && @after.call
        puts err
        return false
      end

      true
    end
  end

  def self.run!
    tests.each { |t| t.run! }
    if tests.any?(&:failed?)
      abort "Failed."
    end
    FileUtils.remove_entry_secure(test_tmpdir)
  end
end
