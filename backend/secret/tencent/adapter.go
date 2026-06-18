// Package tencent provides a secret Manager backed by Tencent Cloud SSM (Secrets Manager).
package tencent

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	tcerr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	ssm "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/ssm/v20190923"
)

const defaultRegion = "ap-guangzhou"

// Config holds Tencent Cloud SSM settings.
type Config struct {
	// SecretID is the Tencent Cloud API SecretId.
	SecretID string
	// SecretKey is the Tencent Cloud API SecretKey.
	SecretKey string
	// Region is the Tencent Cloud region. Defaults to ap-guangzhou.
	Region string
	// Client overrides the default SSM client.
	Client *ssm.Client
}

// Manager reads and writes secrets using Tencent Cloud SSM.
type Manager struct {
	client *ssm.Client
}

// NewManager creates a Tencent Cloud SSM-backed secret Manager.
func NewManager(cfg Config) (*Manager, error) {
	if cfg.Client != nil {
		return &Manager{client: cfg.Client}, nil
	}
	if cfg.SecretID == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("tencent ssm: SecretID and SecretKey are required")
	}

	region := cfg.Region
	if region == "" {
		region = defaultRegion
	}

	credential := common.NewCredential(cfg.SecretID, cfg.SecretKey)
	cpf := profile.NewClientProfile()
	client, err := ssm.NewClient(credential, region, cpf)
	if err != nil {
		return nil, fmt.Errorf("tencent ssm: create client: %w", err)
	}
	return &Manager{client: client}, nil
}

// Get retrieves a secret value by name.
func (m *Manager) Get(ctx context.Context, key string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}

	req := ssm.NewGetSecretValueRequest()
	req.SecretName = common.StringPtr(key)
	out, err := m.client.GetSecretValue(req)
	if err != nil {
		if isNotFound(err) {
			return "", secret.ErrNotFound
		}
		return "", err
	}
	if out.Response == nil || out.Response.SecretString == nil {
		return "", secret.ErrNotFound
	}
	return *out.Response.SecretString, nil
}

// Set creates or updates a secret value.
func (m *Manager) Set(ctx context.Context, key string, value string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	_, err := m.client.GetSecretValue(&ssm.GetSecretValueRequest{
		SecretName: common.StringPtr(key),
	})
	if err != nil {
		if isNotFound(err) {
			createReq := ssm.NewCreateSecretRequest()
			createReq.SecretName = common.StringPtr(key)
			createReq.SecretString = common.StringPtr(value)
			_, createErr := m.client.CreateSecret(createReq)
			return createErr
		}
		return err
	}

	updateReq := ssm.NewUpdateSecretRequest()
	updateReq.SecretName = common.StringPtr(key)
	updateReq.SecretString = common.StringPtr(value)
	_, err = m.client.UpdateSecret(updateReq)
	return err
}

// Delete removes a secret.
func (m *Manager) Delete(ctx context.Context, key string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	req := ssm.NewDeleteSecretRequest()
	req.SecretName = common.StringPtr(key)
	req.RecoveryWindowInDays = common.Uint64Ptr(0)
	_, err := m.client.DeleteSecret(req)
	if err != nil && isNotFound(err) {
		return secret.ErrNotFound
	}
	return err
}

// List returns secret names matching the given prefix.
func (m *Manager) List(ctx context.Context, prefix string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var keys []string
	offset := uint64(0)
	const pageSize = uint64(100)

	for {
		req := ssm.NewListSecretsRequest()
		req.Offset = common.Uint64Ptr(offset)
		req.Limit = common.Uint64Ptr(pageSize)
		out, err := m.client.ListSecrets(req)
		if err != nil {
			return nil, err
		}
		if out.Response == nil || len(out.Response.SecretMetadatas) == 0 {
			break
		}
		for _, item := range out.Response.SecretMetadatas {
			if item.SecretName == nil {
				continue
			}
			name := *item.SecretName
			if prefix == "" || strings.HasPrefix(name, prefix) {
				keys = append(keys, name)
			}
		}
		if len(out.Response.SecretMetadatas) < int(pageSize) {
			break
		}
		offset += pageSize
	}
	return keys, nil
}

func isNotFound(err error) bool {
	if err == nil {
		return false
	}
	var sdkErr *tcerr.TencentCloudSDKError
	if errors.As(err, &sdkErr) {
		switch sdkErr.GetCode() {
		case "ResourceNotFound", "ResourceNotFound.SecretNotFound", "FailedOperation.ResourceNotFound":
			return true
		}
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "not found") || strings.Contains(msg, "resourcenotfound")
}
