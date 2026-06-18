// Package gcp provides a secret Manager backed by Google Cloud Secret Manager.
package gcp

import (
	"context"
	"fmt"
	"strings"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/EthanShen10086/voxera-kit/secret"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config holds Google Cloud Secret Manager settings.
type Config struct {
	// ProjectID is the GCP project ID.
	ProjectID string
	// Client overrides the default Secret Manager client.
	Client *secretmanager.Client
}

// Manager reads and writes secrets using Google Cloud Secret Manager.
type Manager struct {
	client    *secretmanager.Client
	projectID string
}

// NewManager creates a GCP Secret Manager-backed secret Manager.
func NewManager(ctx context.Context, cfg Config) (*Manager, error) {
	client := cfg.Client
	if client == nil {
		var err error
		client, err = secretmanager.NewClient(ctx)
		if err != nil {
			return nil, fmt.Errorf("gcp secrets: create client: %w", err)
		}
	}
	if cfg.ProjectID == "" {
		return nil, fmt.Errorf("gcp secrets: project ID is required")
	}
	return &Manager{client: client, projectID: cfg.ProjectID}, nil
}

// Get retrieves the latest version of a secret.
func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	name := fmt.Sprintf("projects/%s/secrets/%s/versions/latest", m.projectID, key)
	result, err := m.client.AccessSecretVersion(ctx, &secretmanagerpb.AccessSecretVersionRequest{
		Name: name,
	})
	if err != nil {
		if isNotFound(err) {
			return "", secret.ErrNotFound
		}
		return "", err
	}
	return string(result.Payload.Data), nil
}

// Set creates or updates a secret by adding a new version.
func (m *Manager) Set(ctx context.Context, key string, value string) error {
	parent := fmt.Sprintf("projects/%s", m.projectID)
	secretName := fmt.Sprintf("%s/secrets/%s", parent, key)

	_, err := m.client.GetSecret(ctx, &secretmanagerpb.GetSecretRequest{Name: secretName})
	if err != nil {
		if !isNotFound(err) {
			return err
		}
		_, err = m.client.CreateSecret(ctx, &secretmanagerpb.CreateSecretRequest{
			Parent:   parent,
			SecretId: key,
			Secret: &secretmanagerpb.Secret{
				Replication: &secretmanagerpb.Replication{
					Replication: &secretmanagerpb.Replication_Automatic_{
						Automatic: &secretmanagerpb.Replication_Automatic{},
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}

	_, err = m.client.AddSecretVersion(ctx, &secretmanagerpb.AddSecretVersionRequest{
		Parent: secretName,
		Payload: &secretmanagerpb.SecretPayload{
			Data: []byte(value),
		},
	})
	return err
}

// Delete permanently deletes a secret.
func (m *Manager) Delete(ctx context.Context, key string) error {
	name := fmt.Sprintf("projects/%s/secrets/%s", m.projectID, key)
	err := m.client.DeleteSecret(ctx, &secretmanagerpb.DeleteSecretRequest{Name: name})
	if err != nil && isNotFound(err) {
		return secret.ErrNotFound
	}
	return err
}

// List returns secret IDs matching the given prefix.
func (m *Manager) List(ctx context.Context, prefix string) ([]string, error) {
	parent := fmt.Sprintf("projects/%s", m.projectID)
	it := m.client.ListSecrets(ctx, &secretmanagerpb.ListSecretsRequest{Parent: parent})

	var keys []string
	for {
		s, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}
		id := secretIDFromName(s.Name)
		if prefix == "" || strings.HasPrefix(id, prefix) {
			keys = append(keys, id)
		}
	}
	return keys, nil
}

// Close releases the underlying client when owned by this manager.
func (m *Manager) Close() error {
	return m.client.Close()
}

func secretIDFromName(name string) string {
	parts := strings.Split(name, "/")
	if len(parts) == 0 {
		return name
	}
	return parts[len(parts)-1]
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	return status.Code(err) == codes.NotFound
}
