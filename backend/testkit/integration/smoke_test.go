//go:build integration

package integration_test

import (
	"testing"

	"github.com/EthanShen10086/voxera-kit/testkit/contract"
)

func TestDataPlaneSmoke(t *testing.T) {
	contract.RunDataPlaneSmoke(t)
}
