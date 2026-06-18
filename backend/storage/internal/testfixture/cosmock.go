package testfixture

import (
	"encoding/xml"
	"fmt"
	"hash/crc64"
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

type cosStore struct {
	mu           sync.Mutex
	objects      map[string][]byte
	parts        map[string]map[int][]byte // uploadID -> partNumber -> data
	lifecycleXML []byte
}

// StartCOSMock launches an httptest COS-compatible server for adapter tests.
func StartCOSMock(t *testing.T, bucket string) storage.Config {
	t.Helper()
	if bucket == "" {
		bucket = "voxera-cos"
	}
	store := &cosStore{objects: make(map[string][]byte), parts: make(map[string]map[int][]byte)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleCOS(w, r, store)
	}))
	t.Cleanup(srv.Close)
	hostPort := strings.TrimPrefix(srv.URL, "http://")
	return storage.Config{
		Endpoint:  hostPort + "/" + bucket,
		Bucket:    bucket,
		AccessKey: "cos-ak",
		SecretKey: "cos-sk",
		UseSSL:    false,
	}
}

func handleCOS(w http.ResponseWriter, r *http.Request, store *cosStore) {
	key := strings.TrimPrefix(r.URL.Path, "/")
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
			writeCOSOK(w, body)
			return
		}
		store.mu.Lock()
		store.objects[key] = append([]byte(nil), body...)
		store.mu.Unlock()
		writeCOSOK(w, body)
	case http.MethodPost:
		if q.Has("uploads") {
			uploadID := fmt.Sprintf("cos-mp-%d", time.Now().UnixNano())
			store.mu.Lock()
			store.parts[uploadID] = make(map[int][]byte)
			store.mu.Unlock()
			w.Header().Set("Content-Type", "application/xml")
			_, _ = fmt.Fprintf(w, `<InitiateMultipartUploadResult><UploadId>%s</UploadId></InitiateMultipartUploadResult>`, uploadID)
			return
		}
		if uploadID := q.Get("uploadId"); uploadID != "" {
			store.mu.Lock()
			parts := store.parts[uploadID]
			var assembled []byte
			for i := 1; i <= len(parts)+10; i++ {
				if p, ok := parts[i]; ok {
					assembled = append(assembled, p...)
				}
			}
			baseKey := strings.Split(key, "?")[0]
			if baseKey == "" {
				baseKey = key
			}
			store.objects[baseKey] = assembled
			delete(store.parts, uploadID)
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
		if q.Has("prefix") || key == "" || strings.HasSuffix(r.URL.Path, "/") {
			writeCOSList(w, store, q.Get("prefix"))
			return
		}
		store.mu.Lock()
		body, ok := store.objects[key]
		store.mu.Unlock()
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		writeCOSOK(w, body)
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

func writeCOSOK(w http.ResponseWriter, body []byte) {
	sum := crc64.Checksum(body, crc64.MakeTable(crc64.ECMA))
	w.Header().Set("x-cos-hash-crc64ecma", strconv.FormatUint(sum, 10))
	w.Header().Set("ETag", `"mock-etag"`)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(body)
}

func writeCOSList(w http.ResponseWriter, store *cosStore, prefix string) {
	type content struct {
		Key          string `xml:"Key"`
		Size         int    `xml:"Size"`
		ETag         string `xml:"ETag"`
		LastModified string `xml:"LastModified"`
	}
	type result struct {
		XMLName      xml.Name  `xml:"ListBucketResult"`
		IsTruncated  bool      `xml:"IsTruncated"`
		Contents     []content `xml:"Contents"`
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
			Size:         len(v),
			ETag:         `"mock-etag"`,
			LastModified: time.Now().UTC().Format(time.RFC3339),
		})
	}
	w.Header().Set("Content-Type", "application/xml")
	_ = xml.NewEncoder(w).Encode(result{IsTruncated: false, Contents: contents})
}
