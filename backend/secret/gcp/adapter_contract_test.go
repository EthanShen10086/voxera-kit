package gcp_test

import (
	"context"
	"net"
	"strings"
	"sync"
	"testing"

	secretmanager "cloud.google.com/go/secretmanager/apiv1"
	"cloud.google.com/go/secretmanager/apiv1/secretmanagerpb"
	"github.com/EthanShen10086/voxera-kit/secret"
	"github.com/EthanShen10086/voxera-kit/secret/contract"
	gcpstore "github.com/EthanShen10086/voxera-kit/secret/gcp"
	"google.golang.org/api/option"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/test/bufconn"
	"google.golang.org/protobuf/types/known/emptypb"
)

const bufSize = 1024 * 1024

type fakeSecretManager struct {
	secretmanagerpb.UnimplementedSecretManagerServiceServer
	mu      sync.Mutex
	secrets map[string][]byte
}

func secretKey(name string) string {
	const marker = "/secrets/"
	if i := strings.Index(name, marker); i >= 0 {
		rest := name[i+len(marker):]
		if j := strings.Index(rest, "/versions/"); j >= 0 {
			rest = rest[:j]
		}
		return rest
	}
	return name
}

func (f *fakeSecretManager) AccessSecretVersion(_ context.Context, req *secretmanagerpb.AccessSecretVersionRequest) (*secretmanagerpb.AccessSecretVersionResponse, error) {
	key := secretKey(req.GetName())
	f.mu.Lock()
	defer f.mu.Unlock()
	val, ok := f.secrets[key]
	if !ok {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &secretmanagerpb.AccessSecretVersionResponse{
		Payload: &secretmanagerpb.SecretPayload{Data: val},
	}, nil
}

func (f *fakeSecretManager) GetSecret(_ context.Context, req *secretmanagerpb.GetSecretRequest) (*secretmanagerpb.Secret, error) {
	key := secretKey(req.GetName())
	f.mu.Lock()
	_, ok := f.secrets[key]
	f.mu.Unlock()
	if !ok {
		return nil, status.Error(codes.NotFound, "not found")
	}
	return &secretmanagerpb.Secret{Name: req.GetName()}, nil
}

func (f *fakeSecretManager) CreateSecret(_ context.Context, req *secretmanagerpb.CreateSecretRequest) (*secretmanagerpb.Secret, error) {
	f.mu.Lock()
	f.secrets[req.GetSecretId()] = nil
	f.mu.Unlock()
	return &secretmanagerpb.Secret{Name: req.GetParent() + "/secrets/" + req.GetSecretId()}, nil
}

func (f *fakeSecretManager) AddSecretVersion(_ context.Context, req *secretmanagerpb.AddSecretVersionRequest) (*secretmanagerpb.SecretVersion, error) {
	key := secretKey(req.GetParent())
	f.mu.Lock()
	f.secrets[key] = append([]byte(nil), req.GetPayload().GetData()...)
	f.mu.Unlock()
	return &secretmanagerpb.SecretVersion{Name: req.GetParent() + "/versions/1"}, nil
}

func (f *fakeSecretManager) DeleteSecret(_ context.Context, req *secretmanagerpb.DeleteSecretRequest) (*emptypb.Empty, error) {
	key := secretKey(req.GetName())
	f.mu.Lock()
	delete(f.secrets, key)
	f.mu.Unlock()
	return &emptypb.Empty{}, nil
}

func (f *fakeSecretManager) ListSecrets(_ context.Context, req *secretmanagerpb.ListSecretsRequest) (*secretmanagerpb.ListSecretsResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var list []*secretmanagerpb.Secret
	for key := range f.secrets {
		list = append(list, &secretmanagerpb.Secret{Name: req.GetParent() + "/secrets/" + key})
	}
	return &secretmanagerpb.ListSecretsResponse{Secrets: list}, nil
}

func startFakeGCP(t *testing.T) *secretmanager.Client {
	t.Helper()
	lis := bufconn.Listen(bufSize)
	fake := &fakeSecretManager{secrets: make(map[string][]byte)}
	srv := grpc.NewServer()
	secretmanagerpb.RegisterSecretManagerServiceServer(srv, fake)
	go func() { _ = srv.Serve(lis) }()
	t.Cleanup(func() { srv.Stop() })

	ctx := context.Background()
	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) { return lis.Dial() }),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { _ = conn.Close() })

	client, err := secretmanager.NewClient(ctx, option.WithGRPCConn(conn))
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}
	t.Cleanup(func() { _ = client.Close() })
	return client
}

func TestGCPSecretContract(t *testing.T) {
	contract.RunSecretContract(t, func(t *testing.T) (secret.Manager, func()) {
		mgr, err := gcpstore.NewManager(context.Background(), gcpstore.Config{
			ProjectID: "test-project",
			Client:    startFakeGCP(t),
		})
		if err != nil {
			t.Fatalf("NewManager: %v", err)
		}
		return mgr, func() { _ = mgr.Close() }
	})
}

func TestGCPGetNotFound(t *testing.T) {
	mgr, err := gcpstore.NewManager(context.Background(), gcpstore.Config{
		ProjectID: "test-project",
		Client:    startFakeGCP(t),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = mgr.Close() }()
	_, err = mgr.Get(context.Background(), "missing")
	if err != secret.ErrNotFound {
		t.Fatalf("Get() = %v", err)
	}
}
