package commands

import (
	"testing"
)

func TestPushWithEmptyQueue(t *testing.T) {
	repo := NewRepository(t, "empty")
	defer repo.Test()

	cmd := repo.Command("push")
	cmd.Output = ""
}
