require 'asciidoctor'
require 'asciidoctor/extensions'
require 'asciidoctor/converter/manpage'

module GitLFS
  module Documentation
    class GitLFSManPageConverter < Asciidoctor::Converter::ManPageConverter

      extend Asciidoctor::Converter::Config
      register_for 'manpage'

      def convert_listing node
        result = super
        if node.role == 'synopsis'
          # Remove any indentation from literal blocks in our "Synopsis"
          # sections, specifically any opening ".RS <n>" and closing ".RE"
          # macros.
          #
          # At present, Asciidoctor's manpage converter precedes these macros
          # with ".if n" conditional requests so that the indentation is
          # only added when in "n" (i.e., "nroff") mode, which by default
          # is the case when the output device is a terminal.
          #
          # So as to be somewhat forgiving of potential future changes to
          # Asciidoctor's manpage converter, we allow for the conditionals
          # to be absent and for any indentation inset value (at present,
          # Asciidoctor uses four spaces).
          #
          # References:
          # https://www.gnu.org/software/groff/manual/groff.html#Indented-regions-in-ms
          # https://www.gnu.org/software/groff/manual/groff.html#troff-and-nroff-Modes
          # https://github.com/asciidoctor/asciidoctor/blob/fc0d033577d30adbffba73ce06709292fc2cf3ce/lib/asciidoctor/converter/manpage.rb#L253-L260

          indent_re = /\A(?:\.if n )?(?:\.RS \d+|\.RE)\z/
          result = result.split(Asciidoctor::LF)
            .map {|l| l.sub(indent_re, '')}.reject {|l| l.empty?}
            .join(Asciidoctor::LF)
        end
        result
      end
    end
  end
end
