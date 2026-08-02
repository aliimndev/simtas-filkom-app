package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newSupabaseTestServer returns a fake Supabase Storage API server that records
// the last request it received and responds with canned bodies. The returned
// pointers are updated live by the handler, so callers can assert on them.
func newSupabaseTestServer(t *testing.T) (*httptest.Server, **http.Request, *string) {
	t.Helper()
	var lastReq *http.Request
	var lastBody string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		lastBody = string(body)
		lastReq = r
		switch {
		case strings.Contains(r.URL.Path, "/object/sign/"):
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"signedURL":"/object/sign/docs-bucket/theses/x/v1_a.pdf?token=abc123"}`))
		case r.Method == http.MethodDelete:
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`[{"message":"Successfully deleted"}]`))
		default: // upload
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Key":"theses/x/v1_a.pdf"}`))
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &lastReq, &lastBody
}

func TestSupabaseStorageService_Upload(t *testing.T) {
	srv, lastReqPtr, _ := newSupabaseTestServer(t)
	svc := NewSupabaseStorageService(srv.URL, "service-role-key", "docs-bucket", "archives-bucket")

	path, err := svc.Upload(context.Background(), "theses/x/v1_a.pdf", bytes.NewReader([]byte("pdf-bytes")), 9, "application/pdf")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	if path != "theses/x/v1_a.pdf" {
		t.Errorf("Upload returned path %q, want relative path", path)
	}
	lastReq := *lastReqPtr
	if lastReq == nil {
		t.Fatal("no request received")
	}
	if lastReq.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", lastReq.Method)
	}
	if got := lastReq.URL.Path; got != "/storage/v1/object/docs-bucket/theses/x/v1_a.pdf" {
		t.Errorf("upload path = %q, want /storage/v1/object/docs-bucket/...", got)
	}
	if got := lastReq.Header.Get("Authorization"); got != "Bearer service-role-key" {
		t.Errorf("Authorization = %q, want Bearer service-role-key", got)
	}
	if got := lastReq.Header.Get("content-type"); got != "application/pdf" {
		t.Errorf("content-type = %q, want application/pdf", got)
	}
}

func TestSupabaseStorageService_Upload_ArchivesBucket(t *testing.T) {
	srv, lastReqPtr, _ := newSupabaseTestServer(t)
	svc := NewSupabaseStorageService(srv.URL, "k", "docs-bucket", "archives-bucket")

	_, err := svc.Upload(context.Background(), "archives/2027/x/skripsi.pdf", bytes.NewReader([]byte("x")), 1, "application/pdf")
	if err != nil {
		t.Fatalf("Upload returned error: %v", err)
	}
	lastReq := *lastReqPtr
	if lastReq == nil {
		t.Fatal("no request received")
	}
	if got := lastReq.URL.Path; !strings.Contains(got, "/object/archives-bucket/archives/2027/x/skripsi.pdf") {
		t.Errorf("upload path = %q, want archives bucket routing", got)
	}
}

func TestSupabaseStorageService_GeneratePresignedURL(t *testing.T) {
	srv, lastReqPtr, bodyPtr := newSupabaseTestServer(t)
	svc := NewSupabaseStorageService(srv.URL, "k", "docs-bucket", "archives-bucket")

	url, err := svc.GeneratePresignedURL(context.Background(), "theses/x/v1_a.pdf", 900)
	if err != nil {
		t.Fatalf("GeneratePresignedURL returned error: %v", err)
	}
	if !strings.Contains(url, "/object/sign/docs-bucket/theses/x/v1_a.pdf?token=abc123") {
		t.Errorf("signed url = %q, want signed path for docs bucket", url)
	}
	lastReq := *lastReqPtr
	if lastReq == nil {
		t.Fatal("no request received")
	}
	if lastReq.Method != http.MethodPost {
		t.Errorf("method = %s, want POST", lastReq.Method)
	}
	var payload map[string]int
	if err := json.Unmarshal([]byte(*bodyPtr), &payload); err != nil {
		t.Fatalf("sign request body not JSON: %v", err)
	}
	if payload["expiresIn"] != 900 {
		t.Errorf("expiresIn = %d, want 900", payload["expiresIn"])
	}
}

func TestSupabaseStorageService_Delete(t *testing.T) {
	srv, lastReqPtr, bodyPtr := newSupabaseTestServer(t)
	svc := NewSupabaseStorageService(srv.URL, "k", "docs-bucket", "archives-bucket")

	if err := svc.Delete(context.Background(), "theses/x/v1_a.pdf"); err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	lastReq := *lastReqPtr
	if lastReq == nil {
		t.Fatal("no request received")
	}
	if lastReq.Method != http.MethodDelete {
		t.Errorf("method = %s, want DELETE", lastReq.Method)
	}
	if !strings.Contains(*bodyPtr, "theses/x/v1_a.pdf") {
		t.Errorf("delete body = %q, want prefixes containing the path", *bodyPtr)
	}
}

func TestSupabaseStorageService_BaseURLNormalization(t *testing.T) {
	srv, _, _ := newSupabaseTestServer(t)

	// URL without /storage/v1 suffix must still hit the fake server's
	// /storage/v1/* endpoints (the suffix is appended automatically).
	svc := NewSupabaseStorageService(srv.URL, "k", "", "")
	if svc.documentsBucket != "simtas-documents" {
		t.Errorf("documents bucket default = %q", svc.documentsBucket)
	}
	if svc.archivesBucket != "simtas-archives" {
		t.Errorf("archives bucket default = %q", svc.archivesBucket)
	}

	path, err := svc.Upload(context.Background(), "theses/x/v1_a.pdf", bytes.NewReader([]byte("p")), 1, "application/pdf")
	if err != nil {
		t.Fatalf("Upload with normalized URL returned error: %v", err)
	}
	if path != "theses/x/v1_a.pdf" {
		t.Errorf("Upload path = %q", path)
	}
}

func TestSupabaseStorageService_ErrorSurface(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(`{"statusCode":"400","error":"invalid_request","message":"bad file"}`))
	}))
	t.Cleanup(srv.Close)

	svc := NewSupabaseStorageService(srv.URL, "k", "docs-bucket", "archives-bucket")
	if _, err := svc.Upload(context.Background(), "theses/x/v1_a.pdf", bytes.NewReader([]byte("p")), 1, "application/pdf"); err == nil {
		t.Fatal("Upload expected error for 400 response, got nil")
	}
	if _, err := svc.GeneratePresignedURL(context.Background(), "theses/x/v1_a.pdf", 900); err == nil {
		t.Fatal("GeneratePresignedURL expected error for 400 response, got nil")
	}
	if err := svc.Delete(context.Background(), "theses/x/v1_a.pdf"); err == nil {
		t.Fatal("Delete expected error for 400 response, got nil")
	}
}
