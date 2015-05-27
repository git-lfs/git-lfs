package commands

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

var latestPrePush = "#!/bin/sh\ncommand -v git-lfs >/dev/null 2>&1 || { echo >&2 \"\\nThis repository has been set up with Git LFS but Git LFS is not installed.\\n\"; exit 0; }\ngit lfs pre-push \"$@\"\n"

func TestUpdateWithoutPrePushHook(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update")
	cmd.Output = "Updated pre-push hook"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatal(err)
		}

		if string(by) != latestPrePush {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestUpdateWithLatestPrePushHook(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update")
	cmd.Output = "Updated pre-push hook"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte(latestPrePush), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != latestPrePush {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestUpdateForce(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update", "--force")
	cmd.Output = "Updated pre-push hook"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("sup\n"), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != latestPrePush {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestUpdateWithUnexpectedPrePushHook(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update")
	cmd.Output = "Hook already exists: pre-push\n\n" +
		"test\n\n" +
		"Run `git lfs update --force` to overwrite this hook."

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("test\n"), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "test\n" {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestUpdateWithOldPrePushHook_1(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update")
	cmd.Output = "Updated pre-push hook"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("#!/bin/sh\ngit lfs push --stdin $*\n"), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != latestPrePush {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestUpdateWithOldPrePushHook_2(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update")
	cmd.Output = "Updated pre-push hook"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("#!/bin/sh\ngit lfs push --stdin \"$@\"\n"), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != latestPrePush {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}

func TestUpdateWithOldPrePushHook_3(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	repo.AddPath(repo.Path, ".git")
	repo.AddPath(repo.Path, "subdir")

	cmd := repo.Command("update")
	cmd.Output = "Updated pre-push hook"

	prePushHookFile := filepath.Join(repo.Path, ".git", "hooks", "pre-push")

	cmd.Before(func() {
		err := ioutil.WriteFile(prePushHookFile, []byte("#!/bin/sh\ngit lfs pre-push \"$@\""), 0755)
		if err != nil {
			t.Fatalf("Error writing pre-push in Before(): %s", err)
		}
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != latestPrePush {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}
