require "tmpdir"
require "fileutils"

class Suite
  def self.root
    @root ||= File.expand_path("../..", __FILE__)
  end

  def self.bin
    @bin ||= File.join(root, "bin/git-media")
  end

  def self.version
    @version ||= go_cmd(:version)
  end

  def self.tmp
    @tmp ||= Dir.tmpdir
  end

  # Gets existing GIT_* env vars
  def self.env
    @env ||= ENV.inject({}) do |memo, (k, v)|
      if k =~ /\AGIT_/
        memo.update k => v
      else
        memo
      end
    end
  end

  # Packs the ENV values into a string like:
  #
  #   ENV_KEY_1=VALUE
  #   ENV_KEY_2=VALUE
  #
  def self.env_string
    env.inject [] do |memo, (key, value)|
      memo << "#{key}=#{value}"
    end.join("\n")
  end

  def self.tests
    @tests ||= []
  end

  def self.test_tmpdir
    @test_tmpdir ||= begin
      t = File.join(tmp, "git-media-tests")
      FileUtils.rm_rf(t)
      t
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

  # Represents a set of tests to run against a repository.
  class Test
    attr_reader :path

    def initialize(name)
      @repository_name = name
      @path = expand(File.join(Suite.test_tmpdir, name.to_s))
      @paths = [@path]
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

    def write(contents, *relative_parts)
      File.open(File.join(@path, *relative_parts), "w") { |f| f.write(contents)}
    end

    def exec(cmd)
      Suite.exec cmd
    end

    def failed?
      !@successful
    end

    def add_path(path)
      @paths << path
    end

    def command(cmd, output = nil)
      c = Command.new(cmd, output)
      yield c if block_given?
      @commands << c
    end

    def run!
      @paths.each { |p| run(p) }
    end

    def run(path)
      puts "Integration tests for #{path}"
      puts
      @commands.each do |c|
        clone(path) do
          @successful = false unless c.run!
        end
      end
      puts
    end

  private
    def clone(path)
      FileUtils.rm_rf @path
      Dir.chdir File.join(Suite.root, "integration", "repos") do
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

  # Represents a single git media command to run.
  class Command
    attr_accessor :expected

    def initialize(cmd, expected, &block)
      @cmd = cmd
      @expected = expected.strip if expected
      @after = block
    end

    def before(&block)
      @before = block
    end

    def after(&block)
      @after = block
    end

    def run!
      puts "$ git media #{@cmd}"

      @before.call if @before

      actual = Suite.exec @cmd

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

  # run a go file to get some data that only go would know
  def self.go_cmd(file)
    %x{go run #{root}/integration/#{file}.go}.strip
  end

  def self.exec(cmd)
    %x{#{Suite.bin} #{cmd}}.strip
  end
end
