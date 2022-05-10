package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/subprocess"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	mergeDriverAncestor   string
	mergeDriverCurrent    string
	mergeDriverOther      string
	mergeDriverOutput     string
	mergeDriverProgram    string
	mergeDriverMarkerSize int
)

func mergeDriverCommand(cmd *cobra.Command, args []string) {
	if len(mergeDriverAncestor) == 0 || len(mergeDriverCurrent) == 0 || len(mergeDriverOther) == 0 || len(mergeDriverOutput) == 0 {
		Exit(tr.Tr.Get("the --ancestor, --current, --other, and --output options are mandatory"))
	}

	fileSpecifiers := make(map[string]string)
	gf := lfs.NewGitFilter(cfg)
	mergeProcessInput(gf, mergeDriverAncestor, fileSpecifiers, "O")
	mergeProcessInput(gf, mergeDriverCurrent, fileSpecifiers, "A")
	mergeProcessInput(gf, mergeDriverOther, fileSpecifiers, "B")
	mergeProcessInput(gf, "", fileSpecifiers, "D")

	fileSpecifiers["L"] = fmt.Sprintf("%d", mergeDriverMarkerSize)

	if len(mergeDriverProgram) == 0 {
		mergeDriverProgram = "git merge-file --stdout --marker-size=%L %A %O %B >%D"
	}

	status, err := processFiles(fileSpecifiers, mergeDriverProgram, mergeDriverOutput)
	if err != nil {
		ExitWithError(err)
	}
	os.Exit(status)
}

func processFiles(fileSpecifiers map[string]string, program string, outputFile string) (int, error) {
	defer mergeCleanup(fileSpecifiers)

	var exitStatus int
	formattedMergeProgram := subprocess.FormatPercentSequences(mergeDriverProgram, fileSpecifiers)
	cmd, err := subprocess.ExecCommand("sh", "-c", formattedMergeProgram)
	if err != nil {
		return -1, errors.New(tr.Tr.Get("failed to run merge program %q: %s", formattedMergeProgram, err))
	}
	err = cmd.Run()
	// If it runs but exits nonzero, then that means there's conflicts
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitStatus = exitError.ProcessState.ExitCode()
		} else {
			return -1, errors.New(tr.Tr.Get("failed to run merge program %q: %s", formattedMergeProgram, err))
		}
	}

	outputFp, err := os.OpenFile(outputFile, os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return -1, err
	}
	defer outputFp.Close()

	filename := fileSpecifiers["D"]

	stat, err := os.Stat(filename)
	if err != nil {
		return -1, err
	}

	inputFp, err := os.OpenFile(filename, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return -1, err
	}
	defer inputFp.Close()

	gf := lfs.NewGitFilter(cfg)
	_, err = clean(gf, outputFp, inputFp, filename, stat.Size())
	if err != nil {
		return -1, err
	}

	return exitStatus, nil
}

func mergeCleanup(fileSpecifiers map[string]string) {
	ids := []string{"A", "O", "B", "D"}
	for _, id := range ids {
		os.Remove(fileSpecifiers[id])
	}
}

func mergeProcessInput(gf *lfs.GitFilter, filename string, fileSpecifiers map[string]string, tag string) {
	file, err := lfs.TempFile(cfg, fmt.Sprintf("merge-driver-%s", tag))
	if err != nil {
		Exit(tr.Tr.Get("could not create temporary file when merging: %s", err))
	}
	defer file.Close()
	fileSpecifiers[tag] = file.Name()

	if len(filename) == 0 {
		return
	}

	pointer, err := lfs.DecodePointerFromFile(filename)
	if err != nil {
		if errors.IsNotAPointerError(err) {
			file.Close()
			if err := lfs.CopyFileContents(cfg, filename, file.Name()); err != nil {
				os.Remove(file.Name())
				Exit(tr.Tr.Get("could not copy non-LFS content when merging: %s", err))
			}
			return
		} else {
			os.Remove(file.Name())
			Exit(tr.Tr.Get("could not decode pointer when merging: %s", err))
		}
	}
	cb, fp, err := gf.CopyCallbackFile("download", file.Name(), 1, 1)
	if err != nil {
		os.Remove(file.Name())
		Exit(tr.Tr.Get("could not create callback: %s", err))
	}
	defer fp.Close()
	_, err = gf.Smudge(file, pointer, file.Name(), true, getTransferManifestOperationRemote("download", cfg.Remote()), cb)
}

func init() {
	RegisterCommand("merge-driver", mergeDriverCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVarP(&mergeDriverAncestor, "ancestor", "", "", "file with the ancestor version")
		cmd.Flags().StringVarP(&mergeDriverCurrent, "current", "", "", "file with the current version")
		cmd.Flags().StringVarP(&mergeDriverOther, "other", "", "", "file with the other version")
		cmd.Flags().StringVarP(&mergeDriverOutput, "output", "", "", "file with the output version")
		cmd.Flags().StringVarP(&mergeDriverProgram, "program", "", "", "program to run to perform the merge")
		cmd.Flags().IntVarP(&mergeDriverMarkerSize, "marker-size", "", 12, "merge marker size")
	})
}
