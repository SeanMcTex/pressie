package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/SeanMcTex/pressie/internal/store"
)

// setupTestServer creates a server with an initialized gifts directory.
func setupTestServer(t *testing.T, authToken string) (*Server, string) {
	t.Helper()
	dir := t.TempDir()
	os.MkdirAll(filepath.Join(dir, "_private"), 0o755)
	os.MkdirAll(filepath.Join(dir, "_shared"), 0o755)
	store.SaveIndex(dir, &store.IndexFile{Version: 1, Plugins: []store.PluginConfig{}})
	store.SaveGeneralIdeas(dir, []store.Idea{})
	srv := NewServer(dir, authToken)
	return srv, dir
}

func doRequest(t *testing.T, srv *Server, method, path string, body any, token string) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		json.NewEncoder(&buf).Encode(body)
	}
	req := httptest.NewRequest(method, path, &buf)
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	return w
}

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var v map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &v); err != nil {
		t.Fatalf("failed to decode response: %v (body: %s)", err, w.Body.String())
	}
	return v
}

// --- Status ---

func TestHandleStatus(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "GET", "/api/status", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	resp := decodeResponse(t, w)
	if resp["status"] != "ok" {
		t.Errorf("status = %v, want ok", resp["status"])
	}
}

// --- Auth ---

func TestAuth_NoToken_AllowsAPI(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "GET", "/api/status", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d (no auth configured)", w.Code, http.StatusOK)
	}
}

func TestAuth_WithToken_NoCreds_Returns401(t *testing.T) {
	srv, _ := setupTestServer(t, "secret123")
	w := doRequest(t, srv, "GET", "/api/status", nil, "")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_WithToken_BearerHeader_Allows(t *testing.T) {
	srv, _ := setupTestServer(t, "secret123")
	w := doRequest(t, srv, "GET", "/api/status", nil, "secret123")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_WithToken_QueryParam_Allows(t *testing.T) {
	srv, _ := setupTestServer(t, "secret123")
	req := httptest.NewRequest("GET", "/api/status?token=secret123", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_WithToken_Cookie_Allows(t *testing.T) {
	srv, _ := setupTestServer(t, "secret123")
	req := httptest.NewRequest("GET", "/api/status", nil)
	req.AddCookie(&http.Cookie{Name: "pressie_auth", Value: "secret123"})
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
}

func TestAuth_WithToken_WrongToken_Returns401(t *testing.T) {
	srv, _ := setupTestServer(t, "secret123")
	w := doRequest(t, srv, "GET", "/api/status", nil, "wrong")
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestAuth_WithToken_Browser_GetsLoginPage(t *testing.T) {
	srv, _ := setupTestServer(t, "secret123")
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("Access token")) {
		t.Error("expected login page with 'Access token' text")
	}
}

// --- Contacts ---

func TestHandleListContacts_Empty(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "GET", "/api/contacts", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var contacts []map[string]any
	json.Unmarshal(w.Body.Bytes(), &contacts)
	if len(contacts) != 0 {
		t.Errorf("contacts len = %d, want 0", len(contacts))
	}
}

func TestHandleAddIdea_CreatesContact(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"item": "Letterpress print", "tags": []string{"art", "irish"}}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}

	// Verify contact appears in list.
	w = doRequest(t, srv, "GET", "/api/contacts", nil, "")
	var contacts []map[string]any
	json.Unmarshal(w.Body.Bytes(), &contacts)
	if len(contacts) != 1 {
		t.Fatalf("contacts len = %d, want 1", len(contacts))
	}
	if contacts[0]["name"] != "Kris" {
		t.Errorf("name = %v, want Kris", contacts[0]["name"])
	}
}

func TestHandleAddIdea_MissingItem(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"tags": []string{"art"}}, "")
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
}

func TestHandleGetContact(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	// Create a contact with an idea.
	doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"item": "Art print", "tags": []string{"art"}}, "")

	w := doRequest(t, srv, "GET", "/api/contacts/Kris", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var cf store.ContactFile
	json.Unmarshal(w.Body.Bytes(), &cf)
	if cf.Name != "Kris" {
		t.Errorf("Name = %q, want Kris", cf.Name)
	}
	if len(cf.Ideas) != 1 {
		t.Fatalf("Ideas len = %d, want 1", len(cf.Ideas))
	}
	if cf.Ideas[0].Item != "Art print" {
		t.Errorf("Item = %q", cf.Ideas[0].Item)
	}
}

// --- Ideas ---

func TestHandleGetGeneralIdeas_Empty(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "GET", "/api/ideas", nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	var ideas []store.Idea
	json.Unmarshal(w.Body.Bytes(), &ideas)
	if len(ideas) != 0 {
		t.Errorf("ideas len = %d, want 0", len(ideas))
	}
}

func TestHandleAddGeneralIdea(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "POST", "/api/ideas",
		map[string]any{"item": "Ceramic pour-over", "tags": []string{"kitchen"}}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusCreated)
	}
	var idea store.Idea
	json.Unmarshal(w.Body.Bytes(), &idea)
	if idea.Item != "Ceramic pour-over" {
		t.Errorf("Item = %q", idea.Item)
	}
	if idea.Status != "open" {
		t.Errorf("Status = %q, want open", idea.Status)
	}

	// Verify it appears in list.
	w = doRequest(t, srv, "GET", "/api/ideas", nil, "")
	var ideas []store.Idea
	json.Unmarshal(w.Body.Bytes(), &ideas)
	if len(ideas) != 1 {
		t.Fatalf("ideas len = %d, want 1", len(ideas))
	}
}

// --- Gifts ---

func TestHandleAddGiven(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "POST", "/api/contacts/Kris/gifts/given",
		map[string]any{"item": "Custom map", "occasion": "christmas", "date": "2025-12-25"}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	var gift store.Gift
	json.Unmarshal(w.Body.Bytes(), &gift)
	if gift.Item != "Custom map" {
		t.Errorf("Item = %q", gift.Item)
	}
	if gift.Occasion != "christmas" {
		t.Errorf("Occasion = %q", gift.Occasion)
	}
}

func TestHandleAddReceived(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "POST", "/api/contacts/Sam/gifts/received",
		map[string]any{"item": "DADGAD poster", "occasion": "christmas"}, "")
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	var gift store.Gift
	json.Unmarshal(w.Body.Bytes(), &gift)
	if gift.Item != "DADGAD poster" {
		t.Errorf("Item = %q", gift.Item)
	}
}

func TestHandleAddGiven_RetiresMatchingIdea(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	// Add idea first.
	doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"item": "Letterpress print", "tags": []string{"art"}}, "")
	// Give the gift.
	doRequest(t, srv, "POST", "/api/contacts/Kris/gifts/given",
		map[string]any{"item": "Letterpress print", "occasion": "christmas"}, "")
	// Check idea is retired.
	w := doRequest(t, srv, "GET", "/api/contacts/Kris", nil, "")
	var cf store.ContactFile
	json.Unmarshal(w.Body.Bytes(), &cf)
	if len(cf.Ideas) != 1 {
		t.Fatalf("Ideas len = %d, want 1", len(cf.Ideas))
	}
	if cf.Ideas[0].Status != "given" {
		t.Errorf("Status = %q, want given", cf.Ideas[0].Status)
	}
}

// --- Preferences ---

func TestHandleSetPreferences(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "PUT", "/api/contacts/Kris/preferences",
		map[string]any{"preferences": "Favorite colors: blue, green. Shoe size: 11."}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify it's saved.
	w = doRequest(t, srv, "GET", "/api/contacts/Kris", nil, "")
	var cf store.ContactFile
	json.Unmarshal(w.Body.Bytes(), &cf)
	if cf.Preferences != "Favorite colors: blue, green. Shoe size: 11." {
		t.Errorf("Preferences = %q", cf.Preferences)
	}
}

// --- Delete idea ---

func TestHandleDeleteGeneralIdea(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	// Add a general idea.
	w := doRequest(t, srv, "POST", "/api/ideas",
		map[string]any{"item": "Test idea", "tags": []string{"test"}}, "")
	var idea store.Idea
	json.Unmarshal(w.Body.Bytes(), &idea)

	// Delete it.
	w = doRequest(t, srv, "DELETE", "/api/ideas/"+idea.ID, nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify it's gone.
	w = doRequest(t, srv, "GET", "/api/ideas", nil, "")
	var ideas []store.Idea
	json.Unmarshal(w.Body.Bytes(), &ideas)
	if len(ideas) != 0 {
		t.Errorf("ideas len = %d, want 0", len(ideas))
	}
}

func TestHandleDeleteContactIdea(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	// Add an idea to a contact.
	w := doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"item": "Art print", "tags": []string{"art"}}, "")
	var idea store.Idea
	json.Unmarshal(w.Body.Bytes(), &idea)

	// Delete it.
	w = doRequest(t, srv, "DELETE", "/api/contacts/Kris/ideas/"+idea.ID, nil, "")
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify it's gone.
	w = doRequest(t, srv, "GET", "/api/contacts/Kris", nil, "")
	var cf store.ContactFile
	json.Unmarshal(w.Body.Bytes(), &cf)
	if len(cf.Ideas) != 0 {
		t.Errorf("Ideas len = %d, want 0", len(cf.Ideas))
	}
}

func TestHandleDeleteContactIdea_PrunesTags(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	// Add two ideas with different tags.
	doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"item": "Art print", "tags": []string{"art", "irish"}}, "")
	w := doRequest(t, srv, "POST", "/api/contacts/Kris/ideas",
		map[string]any{"item": "Music poster", "tags": []string{"music"}}, "")
	var musicIdea store.Idea
	json.Unmarshal(w.Body.Bytes(), &musicIdea)

	// Delete the music idea.
	doRequest(t, srv, "DELETE", "/api/contacts/Kris/ideas/"+musicIdea.ID, nil, "")

	// Check tags — music should be gone, art and irish should remain.
	w = doRequest(t, srv, "GET", "/api/contacts/Kris", nil, "")
	var cf store.ContactFile
	json.Unmarshal(w.Body.Bytes(), &cf)
	for _, tag := range cf.Tags {
		if tag == "music" {
			t.Error("tag 'music' should have been pruned")
		}
	}
}

func TestHandleDeleteIdea_NotFound(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	w := doRequest(t, srv, "DELETE", "/api/ideas/fake-id", nil, "")
	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}

// --- SPA ---

func TestHandleSPA_ServesIndex(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	body := w.Body.String()
	if !bytes.Contains([]byte(body), []byte("Pressie")) {
		t.Error("expected HTML with 'Pressie' text")
	}
}

func TestHandleSPA_ServesCSS(t *testing.T) {
	srv, _ := setupTestServer(t, "")
	req := httptest.NewRequest("GET", "/style.css", nil)
	w := httptest.NewRecorder()
	srv.Handler().ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if !bytes.Contains([]byte(w.Body.String()), []byte("--accent")) {
		t.Error("expected CSS with --accent variable")
	}
}

// --- ListenAddr ---

func TestListenAddr_NoAuth_Localhost(t *testing.T) {
	srv := NewServer("/tmp", "")
	addr := srv.ListenAddr(7612)
	if addr != "127.0.0.1:7612" {
		t.Errorf("addr = %q, want 127.0.0.1:7612", addr)
	}
}

func TestListenAddr_WithAuth_AllInterfaces(t *testing.T) {
	srv := NewServer("/tmp", "secret")
	addr := srv.ListenAddr(7612)
	if addr != "0.0.0.0:7612" {
		t.Errorf("addr = %q, want 0.0.0.0:7612", addr)
	}
}