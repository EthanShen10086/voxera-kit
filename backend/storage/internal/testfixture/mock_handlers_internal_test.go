package testfixture

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newCOSStore() *cosStore {
	return &cosStore{
		objects: make(map[string][]byte),
		parts:   make(map[string]map[int][]byte),
	}
}

func newOSSStore() *ossStore {
	return &ossStore{
		objects:   make(map[string][]byte),
		uploadIDs: make(map[string]string),
		parts:     make(map[string]map[int][]byte),
	}
}

func TestHandleCOSObjectLifecycle(t *testing.T) {
	store := newCOSStore()
	store.objects["obj.txt"] = []byte("data")

	// versioning
	w := httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodPut, "/?versioning", nil), store)
	if w.Code != http.StatusOK {
		t.Fatalf("versioning put: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodGet, "/?versioning", nil), store)
	if w.Code != http.StatusOK {
		t.Fatalf("versioning get: %d", w.Code)
	}

	// lifecycle round-trip
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodPut, "/?lifecycle", strings.NewReader("<Lifecycle/>")), store)
	if w.Code != http.StatusOK {
		t.Fatalf("lifecycle put: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodGet, "/?lifecycle", nil), store)
	if w.Code != http.StatusOK {
		t.Fatalf("lifecycle get: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodDelete, "/?lifecycle", nil), store)
	if w.Code != http.StatusNoContent {
		t.Fatalf("lifecycle delete: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodGet, "/?lifecycle", nil), store)
	if w.Code != http.StatusNotFound {
		t.Fatalf("lifecycle get empty: %d", w.Code)
	}

	// get/head/list/delete object
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodGet, "/obj.txt", nil), store)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "data") {
		t.Fatalf("get object: %d body=%q", w.Code, w.Body.String())
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodHead, "/obj.txt", nil), store)
	if w.Code != http.StatusOK {
		t.Fatalf("head object: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodGet, "/?prefix=obj", nil), store)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "obj.txt") {
		t.Fatalf("list: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodDelete, "/obj.txt", nil), store)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodGet, "/missing", nil), store)
	if w.Code != http.StatusNotFound {
		t.Fatalf("get missing: %d", w.Code)
	}

	// multipart
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodPost, "/mp.txt?uploads", nil), store)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "UploadId") {
		t.Fatalf("initiate mp: %d %q", w.Code, w.Body.String())
	}
	uploadID := strings.TrimSuffix(strings.Split(strings.Split(w.Body.String(), "<UploadId>")[1], "</UploadId>")[0], "")
	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPut, "/mp.txt?partNumber=1&uploadId="+uploadID, strings.NewReader("part"))
	handleCOS(w, req, store)
	if w.Code != http.StatusOK {
		t.Fatalf("upload part: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodPost, "/mp.txt?uploadId="+uploadID, nil), store)
	if w.Code != http.StatusOK {
		t.Fatalf("complete mp: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodDelete, "/mp.txt?uploadId="+uploadID, nil), store)
	if w.Code != http.StatusNoContent {
		t.Fatalf("abort mp: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleCOS(w, httptest.NewRequest(http.MethodPatch, "/x", nil), store)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("method not allowed: %d", w.Code)
	}
}

func TestHandleOSSObjectLifecycle(t *testing.T) {
	const bucket = "test-bucket"
	store := newOSSStore()
	store.objects["key.bin"] = []byte("oss")

	w := httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodPut, "/"+bucket+"/?versioning", nil), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("versioning put: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodPut, "/"+bucket+"/?lifecycle", strings.NewReader("<Lifecycle/>")), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("lifecycle put: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodGet, "/"+bucket+"/?lifecycle", nil), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("lifecycle get: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodGet, "/"+bucket+"/key.bin", nil), bucket, store)
	body, _ := io.ReadAll(w.Result().Body)
	if w.Code != http.StatusOK || string(body) != "oss" {
		t.Fatalf("get: %d %q", w.Code, body)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodHead, "/"+bucket+"/key.bin", nil), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("head: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodGet, "/"+bucket+"/?prefix=key", nil), bucket, store)
	if w.Code != http.StatusOK || !strings.Contains(w.Body.String(), "key.bin") {
		t.Fatalf("list: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodPost, "/"+bucket+"/big?uploads", nil), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("initiate mp: %d", w.Code)
	}
	uploadID := strings.TrimSuffix(strings.Split(strings.Split(w.Body.String(), "<UploadId>")[1], "</UploadId>")[0], "")

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodPut, "/"+bucket+"/big?partNumber=1&uploadId="+uploadID, strings.NewReader("p")), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("part: %d", w.Code)
	}
	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodPost, "/"+bucket+"/big?uploadId="+uploadID, nil), bucket, store)
	if w.Code != http.StatusOK {
		t.Fatalf("complete: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodDelete, "/"+bucket+"/key.bin", nil), bucket, store)
	if w.Code != http.StatusNoContent {
		t.Fatalf("delete: %d", w.Code)
	}

	w = httptest.NewRecorder()
	handleOSS(w, httptest.NewRequest(http.MethodGet, "/"+bucket+"/missing", nil), bucket, store)
	if w.Code != http.StatusNotFound {
		t.Fatalf("missing: %d", w.Code)
	}
}

func TestS3ServerMinIOEndpoint(t *testing.T) {
	s := StartS3(t, "minio-endpoint")
	cfg := s.MinIOEndpoint()
	if cfg.Endpoint == "" || cfg.UseSSL {
		t.Fatalf("MinIOEndpoint: %#v", cfg)
	}
	if cfg.Bucket != "minio-endpoint" {
		t.Fatalf("bucket = %q", cfg.Bucket)
	}
}
