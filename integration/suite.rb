class Suite
  class Config < Struct.new(:root)
    def bin
      @bin ||= File.join(root, "bin/git-media")
    end

    def version
      @version ||= cmd(:version)
    end

    def tmp
      @tmp ||= cmd(:temp)
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

  def self.global_git_config
    `git config -l --global`.strip.split("\n")
  end

  def self.test(repository)
    t = Test.new(repository, "#{repository}/.git")
    yield t if block_given?
    tests << t
  end

  class Test
    def initialize(*repositories)
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

    def command(cmd, output, &block)
      @commands << Command.new(cmd, output, &block)
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
  end

  class Command
    def initialize(cmd, expected, &block)
      @cmd = cmd
      @expected = expected.strip
      @checks = block
    end

    def run!(repository)
      puts "$ git media #{@cmd}"
      actual = %x{#{Suite.config.bin} #{@cmd}}.strip

      if actual != @expected
        puts "- expected"
        puts @expected
        puts
        puts "- actual"
        puts actual
        puts

        return false
      end

      if err = @checks && @checks.call
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
  end
end
