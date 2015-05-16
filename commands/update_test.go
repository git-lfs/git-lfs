package commands

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

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

		if string(by) != "#!/bin/sh\ngit lfs pre-push \"$@\"\n" {
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
		repo.WriteFile(prePushHookFile, "#!/bin/sh\ngit lfs pre-push \"$@\"\n")
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "#!/bin/sh\ngit lfs pre-push \"$@\"\n" {
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
		repo.WriteFile(prePushHookFile, "sup\n")
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "#!/bin/sh\ngit lfs pre-push \"$@\"\n" {
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
		repo.WriteFile(prePushHookFile, "test\n")
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
		repo.WriteFile(prePushHookFile, "#!/bin/sh\ngit lfs push --stdin $*\n")
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "#!/bin/sh\ngit lfs pre-push \"$@\"\n" {
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
		repo.WriteFile(prePushHookFile, "#!/bin/sh\ngit lfs push --stdin \"$@\"\n")
	})

	cmd.After(func() {
		by, err := ioutil.ReadFile(prePushHookFile)
		if err != nil {
			t.Fatalf("Error writing pre-push in After(): %s", err)
		}

		if string(by) != "#!/bin/sh\ngit lfs pre-push \"$@\"\n" {
			t.Errorf("Unexpected pre-push hook:\n%s", string(by))
		}
	})
}
