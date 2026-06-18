package aws_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/aws"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	awslib "github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type secretsStore struct {
	mu      sync.Mutex
	secrets map[string]string
}

func startSecretsMock(t *testing.T) *secretsmanager.Client {
	t.Helper()
	store := &secretsStore{secrets: make(map[string]string)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		target := r.Header.Get("X-Amz-Target")
		body, _ := io.ReadAll(r.Body)
		var req map[string]any
		_ = json.Unmarshal(body, &req)
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")

		notFound := func() {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]any{
				"__type":  "ResourceNotFoundException",
				"message": "Secrets Manager can't find the specified secret.",
			})
		}

		switch target {
		case "secretsmanager.GetSecretValue":
			name, _ := req["SecretId"].(string)
			store.mu.Lock()
			val, ok := store.secrets[name]
			store.mu.Unlock()
			if !ok {
				notFound()
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"SecretString": val, "Name": name})
		case "secretsmanager.DescribeSecret":
			name, _ := req["SecretId"].(string)
			store.mu.Lock()
			_, ok := store.secrets[name]
			store.mu.Unlock()
			if !ok {
				notFound()
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"Name": name})
		case "secretsmanager.CreateSecret":
			name, _ := req["Name"].(string)
			val, _ := req["SecretString"].(string)
			store.mu.Lock()
			store.secrets[name] = val
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"ARN": "arn:aws:secretsmanager:us-east-1:123:secret:" + name, "Name": name})
		case "secretsmanager.PutSecretValue":
			name, _ := req["SecretId"].(string)
			val, _ := req["SecretString"].(string)
			store.mu.Lock()
			store.secrets[name] = val
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"ARN": "arn:aws:secretsmanager:us-east-1:123:secret:" + name, "Name": name})
		case "secretsmanager.DeleteSecret":
			name, _ := req["SecretId"].(string)
			store.mu.Lock()
			_, ok := store.secrets[name]
			if ok {
				delete(store.secrets, name)
			}
			store.mu.Unlock()
			if !ok {
				notFound()
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"ARN": "arn:aws:secretsmanager:us-east-1:123:secret:" + name, "Name": name})
		case "secretsmanager.ListSecrets":
			store.mu.Lock()
			list := make([]map[string]any, 0, len(store.secrets))
			for name := range store.secrets {
				list = append(list, map[string]any{"Name": name})
			}
			store.mu.Unlock()
			_ = json.NewEncoder(w).Encode(map[string]any{"SecretList": list})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
	t.Cleanup(srv.Close)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("ak", "sk", "")),
	)
	if err != nil {
		t.Fatalf("LoadDefaultConfig: %v", err)
	}
	return secretsmanager.NewFromConfig(cfg, func(o *secretsmanager.Options) {
		o.BaseEndpoint = awslib.String(srv.URL)
	})
}

func TestAWSSecretContract(t *testing.T) {
	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		client := startSecretsMock(t)
		mgr, err := aws.NewManager(context.Background(), aws.Config{Client: client})
		if err != nil {
			t.Fatalf("NewManager: %v", err)
		}
		return mgr, nil
	})
}

func TestAWSGetNotFound(t *testing.T) {
	client := startSecretsMock(t)
	mgr, err := aws.NewManager(context.Background(), aws.Config{Client: client})
	if err != nil {
		t.Fatal(err)
	}
	_, err = mgr.Get(context.Background(), "missing")
	if !errors.Is(err, secret.ErrNotFound) {
		t.Fatalf("Get() = %v, want ErrNotFound", err)
	}
}
