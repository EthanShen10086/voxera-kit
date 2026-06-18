package contract

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/task"
	"github.com/EthanShen10086/voxera-kit/task/memory"
)

func TestTaskContract_Memory(t *testing.T) {
	RunTaskContract(t, func(t *testing.T, handler task.Handler) (task.TaskQueue, func()) {
		return memory.New(memory.Config{Handler: handler}), nil
	})
}
