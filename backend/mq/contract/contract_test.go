package contract

import "testing"

func TestMQContract_Memory(t *testing.T) {
	RunMQContract(t, memoryFactory)
}
