package tq

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSSHAdapterWorkerStartingNilTransfer(t *testing.T) {
	a := &SSHAdapter{
		adapterBase: newAdapterBase(nil, SSHAdapterName, Download, nil),
		transfer:    nil,
	}
	a.transferImpl = a

	_, err := a.WorkerStarting(0)
	assert.NotNil(t, err)
	assert.Contains(t, err.Error(), "SSH transfer adapter")
}
