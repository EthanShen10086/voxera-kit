package contract

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/EthanShen10086/voxera-kit/task/memory"
)

func TestTaskAdvancedContract_Memory(t *testing.T) {
	RunTaskAdvancedContract(t, func(t *testing.T, handler task.Handler) (task.TaskQueue, func()) {
		q := memory.New(memory.Config{Handler: handler})
		return q, func() { q.Stop() }
	})
}
