package commands

import (
	"fmt"
	"github.com/github/git-media/gitmedia"
	"testing"
)

func TestVersionOnEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("version")
	cmd.Output = gitmedia.UserAgent

	cmd = repo.Command("version", "--comics")
	cmd.Output = fmt.Sprintf("%s\nNothing may see Gah Lak Tus and survive!", gitmedia.UserAgent)

	cmd = repo.Command("version", "-c")
	cmd.Output = fmt.Sprintf("%s\nNothing may see Gah Lak Tus and survive!", gitmedia.UserAgent)
}
