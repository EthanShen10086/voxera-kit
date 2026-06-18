//go:build integration

package vault_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	vaultadapter "github.com/EthanShen10086/voxera-kit/secret/vault"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestSecretContract_Vault(t *testing.T) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "hashicorp/vault:1.15",
		ExposedPorts: []string{"8200/tcp"},
		Env: map[string]string{
			"VAULT_DEV_ROOT_TOKEN_ID":  "root",
			"VAULT_DEV_LISTEN_ADDRESS": "0.0.0.0:8200",
		},
		Cmd: []string{"server", "-dev"},
		WaitingFor: wait.ForHTTP("/v1/sys/health").
			WithPort("8200/tcp").
			WithStartupTimeout(60 * time.Second),
	}

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("start vault: %v", err)
	}
	defer func() { _ = container.Terminate(ctx) }()

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatal(err)
	}
	port, err := container.MappedPort(ctx, "8200/tcp")
	if err != nil {
		t.Fatal(err)
	}

	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		mgr, err := vaultadapter.NewManager(vaultadapter.Config{
			Address: fmt.Sprintf("http://%s:%s", host, port.Port()),
			Token:   "root",
			Mount:   "secret",
		})
		if err != nil {
			t.Fatalf("NewManager: %v", err)
		}
		return mgr, nil
	})
}
