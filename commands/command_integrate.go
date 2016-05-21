package commands

import (
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"strings"

	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/cobra"
	"github.com/github/git-lfs/vendor/_nuts/github.com/spf13/pflag"
)

var (
	fileTemplate1 = `

_git-lfs() {
  local -a commands help
  commands=(
`
	fileTemplate2 = `)

  help=(
  'help:Command help'
  )

  _arguments \
    "1: :{_describe 'command' commands -- help}" \
    "*:: :->args"

  case $state in
    args)
      case $words[1] in
`
	fileTemplate3 = `
        help)
          _arguments \
            "1: :{_describe 'command' commands}"
          ;;
      esac
      ;;
  esac
}

compdef _git-lfs git-lfs
`
	integrateCmd = &cobra.Command{
		Use:   "integrate",
		Short: "Integrate git-lfs with external tools",
		Run:   integrateCommand,
	}

	zshOutputFile  string
	zshOhMyZshCore bool
)

func getHomeDir() string {
	usr, err := user.Current()
	if err != nil {
		Exit("Unable to locate home directory: %v", err)
	}
	return usr.HomeDir
}

func getOhMyZshRootDir() string {
	// Can't rely on ZSH_CUSTOM etc because it's only defined inside zsh shell, it's
	// not actually a proper environment var visible in child processes
	zsh := os.Getenv("ZSH")
	if len(zsh) > 0 {
		// Check that it is ohmyzsh
		testfile := filepath.Join(zsh, "oh-my-zsh.sh")
		if _, err := os.Stat(testfile); err == nil {
			return zsh
		}
	}

	return ""
}

func integrateCommand(cmd *cobra.Command, args []string) {

	if len(args) == 0 {
		Exit("Required: tool argument")
	}

	switch args[0] {
	case "zsh":
		zshCommand(args[1:])
	}
}

func zshCommand(args []string) {
	var outputFile string

	if len(zshOutputFile) > 0 {
		outputFile = zshOutputFile
	} else if zshOhMyZshCore {
		// Write to core oh-my-zsh for submission upstream
		ohmyzsh := getOhMyZshRootDir()
		if len(ohmyzsh) == 0 {
			Exit("Could not locate oh-my-zsh root dir via $ZSH")
		}
		outputFile = filepath.Join(ohmyzsh, "plugins", "git-lfs", "git-lfs.plugin.zsh")
	} else if ohmyzsh := getOhMyZshRootDir(); len(ohmyzsh) > 0 {
		// default user-specific oh-my-zsh
		outputFile = filepath.Join(getOhMyZshRootDir(), "custom", "plugins", "git-lfs", "git-lfs.plugin.zsh")
	} else {
		outputFile = filepath.Join(getHomeDir(), "_git-lfs.zsh")
	}

	Print("Writing zsh script to %q", outputFile)

	os.MkdirAll(filepath.Dir(outputFile), 0755)

	f, err := os.Create(outputFile)
	if err != nil {
		Exit("Unable to write to %q: %v", outputFile, err)
	}
	defer f.Close()

	f.WriteString(fileTemplate1)
	zshWriteCommandList(f)
	f.WriteString(fileTemplate2)
	zshWriteCommandDetails(f)
	f.WriteString(fileTemplate3)

}

func zshWriteCommandList(f *os.File) {
	// Template example:
	// 'env:Display the Git LFS environment'
	for _, cmd := range RootCmd.Commands() {
		// Only display non-deprecated commands
		if len(cmd.Deprecated) == 0 {
			fmt.Fprintf(f, "  '%s:%s'\n", cmd.Name(), cmd.Short)
		}
	}
}

func zshWriteCommandDetails(f *os.File) {

	// Template example:
	// pointer)
	//   _arguments \
	//     '--file=[A local file to build the pointer from]:file:_files' \
	//     '--pointer=[A local file including the contents of a pointer]:file:_files' \
	//     '--stdin[Reads the pointer from STDIN to compare with the pointer generated from --file]'
	//   ;;

	for _, cmd := range RootCmd.Commands() {
		// Only display non-deprecated commands
		if len(cmd.Deprecated) > 0 {
			continue
		}
		// Skip help command, that's manually dealt with in template
		if cmd.Name() == "help" {
			continue
		}

		fmt.Fprintf(f, "        %s)\n", cmd.Name())
		fmt.Fprintf(f, "          _arguments \\\n")
		cmd.Flags().VisitAll(func(flag *pflag.Flag) {
			// Only non-deprecated flags
			if len(flag.Deprecated) > 0 {
				return
			}
			safeUsage := strings.Replace(flag.Usage, "'", "'\\''", -1)

			var argJoin string
			// Only way to detect an option without an explicit value?
			if flag.DefValue != "false" {
				argJoin = "="
			}
			if len(flag.Shorthand) > 0 {
				if len(flag.Name) > 0 {
					fmt.Fprintf(f, "            '(-%s)--%s%s[%s]' \\\n", flag.Shorthand, flag.Name, argJoin, safeUsage)
				} else {
					fmt.Fprintf(f, "            '-%s%s[%s]' \\\n", flag.Shorthand, argJoin, flag.Name, safeUsage)
				}
			} else {
				fmt.Fprintf(f, "            '--%s%s[%s]' \\\n", flag.Name, argJoin, safeUsage)
			}
		})

		// Determine non-opt arguments from ManPages, first line
		if man, ok := ManPages[cmd.Name()]; ok {
			man1stLine := strings.SplitN(man, "\n", 1)[0]
			if strings.Contains(man1stLine, "<remote>") {
				fmt.Fprintf(f, "              '::remote:__git_remotes' \\\n")
			}
			if strings.Contains(man1stLine, "<ref>") {
				fmt.Fprintf(f, "              '::ref:__git_branch_names' \\\n")
			}
			if strings.Contains(man1stLine, "<path>") {
				fmt.Fprintf(f, "              '*:path:_files' \\\n")
			}
		}
		f.WriteString("\n          ;;\n")
	}
}
func init() {
	integrateCmd.Flags().StringVarP(&zshOutputFile, "output", "o", "", "Write zsh script to the named file")
	integrateCmd.Flags().BoolVarP(&zshOhMyZshCore, "oh-my-zsh-core", "", false, "Install in Oh My Zsh core plugins folder instead of custom")
	RootCmd.AddCommand(integrateCmd)
}
