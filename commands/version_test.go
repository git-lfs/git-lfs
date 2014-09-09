package commands

import (
	"fmt"
	"github.com/github/git-media/gitconfig"
	"github.com/github/git-media/gitmedia"
	"testing"
)

func TestVersionOnEmptyRepository(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	v, err := gitconfig.Version()
	if err != nil {
		t.Fatalf("Error getting git version: %s", err)
	}

	basicVersion := fmt.Sprintf("%s\n%s", gitmedia.UserAgent, v)

	cmd := repo.Command("version")
	cmd.Output = basicVersion

	cmd = repo.Command("version", "--comics")
	cmd.Output = fmt.Sprintf("%s\nNothing may see Gah Lak Tus and survive!", basicVersion)

	cmd = repo.Command("version", "-c")
	cmd.Output = fmt.Sprintf("%s\nNothing may see Gah Lak Tus and survive!", basicVersion)
}
