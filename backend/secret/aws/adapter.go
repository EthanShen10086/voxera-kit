// Package aws provides a secret Manager backed by AWS Secrets Manager.
package aws

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
)

// Config holds AWS Secrets Manager settings.
type Config struct {
	// Region is the AWS region. Loaded from the environment when empty.
	Region string
	// Client overrides the default Secrets Manager client.
	Client *secretsmanager.Client
}

// Manager reads and writes secrets using AWS Secrets Manager.
type Manager struct {
	client *secretsmanager.Client
}

// NewManager creates an AWS Secrets Manager-backed secret Manager.
func NewManager(ctx context.Context, cfg Config) (*Manager, error) {
	if cfg.Client != nil {
		return &Manager{client: cfg.Client}, nil
	}

	opts := []func(*config.LoadOptions) error{}
	if cfg.Region != "" {
		opts = append(opts, config.WithRegion(cfg.Region))
	}

	awsCfg, err := config.LoadDefaultConfig(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("aws secrets: load config: %w", err)
	}

	return &Manager{client: secretsmanager.NewFromConfig(awsCfg)}, nil
}

// Get retrieves a secret value by name or ARN.
func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	out, err := m.client.GetSecretValue(ctx, &secretsmanager.GetSecretValueInput{
		SecretId: aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			return "", secret.ErrNotFound
		}
		return "", err
	}
	if out.SecretString == nil {
		return "", secret.ErrNotFound
	}
	return *out.SecretString, nil
}

// Set creates or updates a secret value.
func (m *Manager) Set(ctx context.Context, key string, value string) error {
	_, err := m.client.DescribeSecret(ctx, &secretsmanager.DescribeSecretInput{
		SecretId: aws.String(key),
	})
	if err != nil {
		if isNotFound(err) {
			_, createErr := m.client.CreateSecret(ctx, &secretsmanager.CreateSecretInput{
				Name:         aws.String(key),
				SecretString: aws.String(value),
			})
			return createErr
		}
		return err
	}

	_, err = m.client.PutSecretValue(ctx, &secretsmanager.PutSecretValueInput{
		SecretId:     aws.String(key),
		SecretString: aws.String(value),
	})
	return err
}

// Delete schedules deletion of a secret.
func (m *Manager) Delete(ctx context.Context, key string) error {
	_, err := m.client.DeleteSecret(ctx, &secretsmanager.DeleteSecretInput{
		SecretId:                   aws.String(key),
		ForceDeleteWithoutRecovery: aws.Bool(true),
	})
	if err != nil && isNotFound(err) {
		return secret.ErrNotFound
	}
	return err
}

// List returns secret names matching the given prefix.
func (m *Manager) List(ctx context.Context, prefix string) ([]string, error) {
	var keys []string
	paginator := secretsmanager.NewListSecretsPaginator(m.client, &secretsmanager.ListSecretsInput{})

	for paginator.HasMorePages() {
		page, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		for _, item := range page.SecretList {
			if item.Name == nil {
				continue
			}
			name := *item.Name
			if prefix == "" || strings.HasPrefix(name, prefix) {
				keys = append(keys, name)
			}
		}
	}
	return keys, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var notFound *types.ResourceNotFoundException
	return errors.As(err, &notFound)
}
