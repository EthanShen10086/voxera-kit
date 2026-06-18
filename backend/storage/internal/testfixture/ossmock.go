package testfixture

import (
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/EthanShen10086/voxera-kit/storage"
)

type ossStore struct {
	mu           sync.Mutex
	objects      map[string][]byte
	uploadIDs    map[string]string // uploadID -> object key
	parts        map[string]map[int][]byte
	lifecycleXML []byte
}

// StartOSSMock launches an httptest OSS-compatible server for adapter tests.
func StartOSSMock(t *testing.T, bucket string) storage.Config {
	t.Helper()
	if bucket == "" {
		bucket = "voxera-oss"
	}
	store := &ossStore{
		objects:   make(map[string][]byte),
		uploadIDs: make(map[string]string),
		parts:     make(map[string]map[int][]byte),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleOSS(w, r, bucket, store)
	}))
	t.Cleanup(srv.Close)
	return storage.Config{
		Endpoint:  srv.URL,
		Bucket:    bucket,
		AccessKey: "oss-ak",
		SecretKey: "oss-sk",
	}
}

func handleOSS(w http.ResponseWriter, r *http.Request, bucket string, store *ossStore) {
	path := strings.TrimPrefix(r.URL.Path, "/")
	path = strings.TrimPrefix(path, bucket+"/")
	key := path
	q := r.URL.Query()

	switch r.Method {
	case http.MethodPut:
		if q.Has("versioning") {
			w.WriteHeader(http.StatusOK)
			return
		}
		if q.Has("lifecycle") {
			body, _ := io.ReadAll(r.Body)
			store.mu.Lock()
			store.lifecycleXML = append([]byte(nil), body...)
			store.mu.Unlock()
			w.WriteHeader(http.StatusOK)
			return
		}
		body, _ := io.ReadAll(r.Body)
		if uploadID := q.Get("uploadId"); uploadID != "" {
			partNum, _ := strconv.Atoi(q.Get("partNumber"))
			store.mu.Lock()
			if store.parts[uploadID] == nil {
				store.parts[uploadID] = make(map[int][]byte)
			}
			store.parts[uploadID][partNum] = append([]byte(nil), body...)
			store.mu.Unlock()
			w.Header().Set("ETag", fmt.Sprintf(`"part-%d"`, partNum))
			w.WriteHeader(http.StatusOK)
			return
		}
		store.mu.Lock()
		store.objects[key] = append([]byte(nil), body...)
		store.mu.Unlock()
		w.Header().Set("ETag", `"mock-etag"`)
		w.WriteHeader(http.StatusOK)
	case http.MethodPost:
		if q.Has("uploads") {
			uploadID := fmt.Sprintf("oss-mp-%d", time.Now().UnixNano())
			store.mu.Lock()
			store.parts[uploadID] = make(map[int][]byte)
			store.uploadIDs[uploadID] = key
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprintf(w, `<InitiateMultipartUploadResult><UploadId>%s</UploadId><Bucket>%s</Bucket><Key>%s</Key></InitiateMultipartUploadResult>`, uploadID, bucket, key)
			return
		}
		if uploadID := q.Get("uploadId"); uploadID != "" {
			store.mu.Lock()
			objKey := store.uploadIDs[uploadID]
			parts := store.parts[uploadID]
			var assembled []byte
			for i := 1; i <= len(parts)+10; i++ {
				if p, ok := parts[i]; ok {
					assembled = append(assembled, p...)
				}
			}
			store.objects[objKey] = assembled
			delete(store.parts, uploadID)
			delete(store.uploadIDs, uploadID)
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<CompleteMultipartUploadResult><ETag>"mp"</ETag></CompleteMultipartUploadResult>`))
			return
		}
		w.WriteHeader(http.StatusBadRequest)
	case http.MethodGet:
		if q.Has("versioning") {
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write([]byte(`<VersioningConfiguration><Status>Suspended</Status></VersioningConfiguration>`))
			return
		}
		if q.Has("lifecycle") {
			store.mu.Lock()
			xmlBody := store.lifecycleXML
			store.mu.Unlock()
			if len(xmlBody) == 0 {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.Header().Set("Content-Type", "application/xml")
			_, _ = w.Write(xmlBody)
			return
		}
		if q.Has("prefix") || key == "" {
			writeOSSList(w, store, q.Get("prefix"))
			return
		}
		store.mu.Lock()
		body, ok := store.objects[key]
		store.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	case http.MethodHead:
		store.mu.Lock()
		_, ok := store.objects[key]
		store.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	case http.MethodDelete:
		if q.Has("lifecycle") {
			store.mu.Lock()
			store.lifecycleXML = nil
			store.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if uploadID := q.Get("uploadId"); uploadID != "" {
			store.mu.Lock()
			delete(store.parts, uploadID)
			delete(store.uploadIDs, uploadID)
			store.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
			return
		}
		store.mu.Lock()
		delete(store.objects, key)
		store.mu.Unlock()
		w.WriteHeader(http.StatusNoContent)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func writeOSSList(w http.ResponseWriter, store *ossStore, prefix string) {
	type content struct {
		Key          string `xml:"Key"`
		Size         int64  `xml:"Size"`
		ETag         string `xml:"ETag"`
		LastModified string `xml:"LastModified"`
	}
	type result struct {
		XMLName     xml.Name  `xml:"ListBucketResult"`
		IsTruncated bool      `xml:"IsTruncated"`
		Contents    []content `xml:"Contents"`
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	var contents []content
	for k, v := range store.objects {
		if prefix != "" && !strings.HasPrefix(k, prefix) {
			continue
		}
		contents = append(contents, content{
			Key:          k,
			Size:         int64(len(v)),
			ETag:         `"mock-etag"`,
			LastModified: time.Now().UTC().Format(time.RFC3339),
		})
	}
	w.Header().Set("Content-Type", "application/xml")
	_ = xml.NewEncoder(w).Encode(result{IsTruncated: false, Contents: contents})
}
