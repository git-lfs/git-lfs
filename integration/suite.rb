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

  # Clones the sample repository to a test
  def self.clone(name)
    path = File.join(config.root, "integration", "repos", name.to_s)
    if !File.exist?(path)
      path += ".git"
    end

    if !File.exist?(path)
      raise ArgumentError, "No example repository #{name} (#{path.inspect})"
    end

    dest = File.join(test_tmpdir, name.to_s)
    Dir.chdir File.join(config.root, "integration", "repos") do
      %x{git clone #{name} #{dest}}
      # set a default origin remote for each test case
      Dir.chdir dest do
        `git remote remove origin`
        `git remote add origin https://example.com/git/media`
      end
    end

    dest
  end

  def self.test(repo_name)
    t = Test.new(clone(repo_name))
    yield t if block_given?
    tests << t
  end

  class Test
    attr_reader :path

    def initialize(*repositories)
      @path = expand(repositories.first)
      @repositories = repositories
      @commands = []
      @successful = true
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
      @repositories.each do |r|
        Dir.chdir(r) { run(r) }
      end
    end

    def run(r)
      puts "Integration tests for #{r}"
      puts
      @commands.each do |c|
        if !c.run!(r)
          @successful = false
        end
      end
      puts
    end

  private
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
