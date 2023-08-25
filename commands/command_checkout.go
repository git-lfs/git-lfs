package commands

import (
	"fmt"
	"os"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/git"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tasklog"
	"github.com/git-lfs/git-lfs/v3/tq"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

var (
	checkoutTo     string
	checkoutBase   bool
	checkoutOurs   bool
	checkoutTheirs bool
)

func checkoutCommand(cmd *cobra.Command, args []string) {
	setupRepository()

	stage, err := whichCheckout()
	if err != nil {
		Exit(tr.Tr.Get("Error parsing args: %v", err))
	}

	if checkoutTo != "" && stage != git.IndexStageDefault {
		if len(args) != 1 {
			Exit(tr.Tr.Get("--to requires exactly one Git LFS object file path"))
		}
		checkoutConflict(rootedPaths(args)[0], stage)
		return
	} else if checkoutTo != "" || stage != git.IndexStageDefault {
		Exit(tr.Tr.Get("--to and exactly one of --theirs, --ours, and --base must be used together"))
	}

	ref, err := git.CurrentRef()
	if err != nil {
		Panic(err, tr.Tr.Get("Could not checkout"))
	}

	singleCheckout := newSingleCheckout(cfg.Git, "")
	if singleCheckout.Skip() {
		fmt.Println(tr.Tr.Get("Cannot checkout LFS objects, Git LFS is not installed."))
		return
	}

	var totalBytes int64
	var pointers []*lfs.WrappedPointer
	logger := tasklog.NewLogger(os.Stdout,
		tasklog.ForceProgress(cfg.ForceProgress()),
	)
	meter := tq.NewMeter(cfg)
	meter.Direction = tq.Checkout
	meter.Logger = meter.LoggerFromEnv(cfg.Os)
	logger.Enqueue(meter)
	chgitscanner := lfs.NewGitScanner(cfg, func(p *lfs.WrappedPointer, err error) {
		if err != nil {
			LoggedError(err, tr.Tr.Get("Scanner error: %s", err))
			return
		}

		totalBytes += p.Size
		meter.Add(p.Size)
		meter.StartTransfer(p.Name)
		pointers = append(pointers, p)
	})

	chgitscanner.Filter = filepathfilter.New(rootedPaths(args), nil, filepathfilter.GitIgnore)

	if err := chgitscanner.ScanTree(ref.Sha, nil); err != nil {
		ExitWithError(err)
	}

	meter.Start()
	for _, p := range pointers {
		singleCheckout.Run(p)

		// not strictly correct (parallel) but we don't have a callback & it's just local
		// plus only 1 slot in channel so it'll block & be close
		meter.TransferBytes("checkout", p.Name, p.Size, totalBytes, int(p.Size))
		meter.FinishTransfer(p.Name)
	}

	meter.Finish()
	singleCheckout.Close()
}

func checkoutConflict(file string, stage git.IndexStage) {
	singleCheckout := newSingleCheckout(cfg.Git, "")
	if singleCheckout.Skip() {
		fmt.Println(tr.Tr.Get("Cannot checkout LFS objects, Git LFS is not installed."))
		return
	}

	ref, err := git.ResolveRef(fmt.Sprintf(":%d:%s", stage, file))
	if err != nil {
		Exit(tr.Tr.Get("Could not checkout (are you not in the middle of a merge?): %v", err))
	}

	scanner, err := git.NewObjectScanner(cfg.GitEnv(), cfg.OSEnv())
	if err != nil {
		Exit(tr.Tr.Get("Could not create object scanner: %v", err))
	}

	if !scanner.Scan(ref.Sha) {
		Exit(tr.Tr.Get("Could not find object %q", ref.Sha))
	}

	ptr, err := lfs.DecodePointer(scanner.Contents())
	if err != nil {
		Exit(tr.Tr.Get("Could not find decoder pointer for object %q: %v", ref.Sha, err))
	}

	p := &lfs.WrappedPointer{Name: file, Pointer: ptr}

	if err := singleCheckout.RunToPath(p, checkoutTo); err != nil {
		Exit(tr.Tr.Get("Error checking out %v to %q: %v", ref.Sha, checkoutTo, err))
	}
	singleCheckout.Close()
}

func whichCheckout() (stage git.IndexStage, err error) {
	seen := 0
	stage = git.IndexStageDefault

	if checkoutBase {
		seen++
		stage = git.IndexStageBase
	}
	if checkoutOurs {
		seen++
		stage = git.IndexStageOurs
	}
	if checkoutTheirs {
		seen++
		stage = git.IndexStageTheirs
	}

	if seen > 1 {
		return 0, errors.New(tr.Tr.Get("at most one of --base, --theirs, and --ours is allowed"))
	}
	return stage, nil
}

// Parameters are filters
// firstly convert any pathspecs to patterns relative to the root of the repo,
// in case this is being executed in a sub-folder
func rootedPaths(args []string) []string {
	pathConverter, err := lfs.NewCurrentToRepoPatternConverter(cfg)
	if err != nil {
		Panic(err, tr.Tr.Get("Could not checkout"))
	}

	rootedpaths := make([]string, 0, len(args))
	for _, arg := range args {
		rootedpaths = append(rootedpaths, pathConverter.Convert(arg))
	}
	return rootedpaths
}

func init() {
	RegisterCommand("checkout", checkoutCommand, func(cmd *cobra.Command) {
		cmd.Flags().StringVar(&checkoutTo, "to", "", "Checkout a conflicted file to this path")
		cmd.Flags().BoolVar(&checkoutOurs, "ours", false, "Checkout our version of a conflicted file")
		cmd.Flags().BoolVar(&checkoutTheirs, "theirs", false, "Checkout their version of a conflicted file")
		cmd.Flags().BoolVar(&checkoutBase, "base", false, "Checkout the base version of a conflicted file")
	})
}
