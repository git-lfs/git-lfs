package commands

import (
	"path/filepath"
	"strings"

	"github.com/git-lfs/git-lfs/v3/errors"
	"github.com/git-lfs/git-lfs/v3/filepathfilter"
	"github.com/git-lfs/git-lfs/v3/lfs"
	"github.com/git-lfs/git-lfs/v3/tools"
	"github.com/git-lfs/git-lfs/v3/tr"
	"github.com/spf13/cobra"
)

func splitArgsAtDash(cmd *cobra.Command, args []string) (commandArgs, paths []string) {
	if n := cmd.ArgsLenAtDash(); n >= 0 {
		return args[:n], args[n:]
	}
	return args, nil
}

type pathFilterPattern struct {
	filter *filepathfilter.Filter
	paths  []filepathfilter.Pattern
}

func (p *pathFilterPattern) Match(filename string) bool {
	if !p.filter.Allows(filename) {
		return false
	}
	for _, path := range p.paths {
		if path.Match(filename) {
			return true
		}
	}
	return false
}

func (p *pathFilterPattern) String() string {
	paths := make([]string, 0, len(p.paths))
	for _, path := range p.paths {
		paths = append(paths, path.String())
	}
	return strings.Join(paths, ",")
}

func buildFilepathFilterForPaths(filter *filepathfilter.Filter, paths []string) (*filepathfilter.Filter, error) {
	rooted, err := repositoryRelativePaths(paths)
	if err != nil {
		return nil, err
	}

	patterns := make([]filepathfilter.Pattern, 0, len(rooted))
	for _, path := range rooted {
		patterns = append(patterns, filepathfilter.NewLiteralPathPattern(path, cfg.Git))
	}

	return filepathfilter.NewFromPatterns(
		[]filepathfilter.Pattern{&pathFilterPattern{filter: filter, paths: patterns}},
		nil,
		filepathfilter.DefaultValue(false),
		determineFilepathFilterCache(cfg),
	), nil
}

func repositoryRelativePaths(paths []string) ([]string, error) {
	rooted := make([]string, 0, len(paths))
	workingDir := cfg.LocalWorkingDir()

	var converter lfs.PathConverter
	if workingDir != "" {
		var err error
		converter, err = lfs.NewCurrentToRepoPathConverter(cfg)
		if err != nil {
			return nil, err
		}
	}

	for _, original := range paths {
		if original == "" {
			return nil, errors.New(tr.Tr.Get("empty path is not valid"))
		}

		converted := original
		if converter != nil {
			converted = converter.Convert(original)
			// The converter passes paths through when invoked at the repository
			// root, so make absolute paths relative explicitly in that case.
			if filepath.IsAbs(converted) {
				var err error
				converted, err = filepath.Rel(workingDir, tools.ResolveSymlinks(converted))
				if err != nil {
					return nil, err
				}
			}
		} else if filepath.IsAbs(converted) {
			return nil, errors.New(tr.Tr.Get("path %q is outside the repository", original))
		}

		converted = filepath.ToSlash(filepath.Clean(converted))
		if converted == ".." || strings.HasPrefix(converted, "../") {
			return nil, errors.New(tr.Tr.Get("path %q is outside the repository", original))
		}
		rooted = append(rooted, converted)
	}

	return rooted, nil
}
