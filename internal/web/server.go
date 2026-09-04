package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/SeanMcTex/pressie/internal/config"
	"github.com/SeanMcTex/pressie/internal/gifts"
	"github.com/SeanMcTex/pressie/internal/store"
)

// Server is the pressie web server. It reads/writes the same JSON files
// as the CLI, using the internal/store and internal/gifts packages directly.
type Server struct {
	giftsDir  string
	authToken string
	mux       *http.ServeMux
}

// NewServer creates a web server for the given gifts directory.
// If authToken is non-empty, requests must include it as a Bearer token
// or ?token= query param. If empty, no auth is required (localhost only).
func NewServer(giftsDir, authToken string) *Server {
	s := &Server{
		giftsDir:  giftsDir,
		authToken: authToken,
		mux:       http.NewServeMux(),
	}
	s.registerRoutes()
	return s
}

// Handler returns the HTTP handler (for testing).
func (s *Server) Handler() http.Handler {
	return s.authMiddleware(s.mux)
}

// ListenAddr returns the address to listen on.
// With auth: 0.0.0.0:<port> (accessible from network)
// Without auth: 127.0.0.1:<port> (localhost only)
func (s *Server) ListenAddr(port int) string {
	if s.authToken != "" {
		return fmt.Sprintf("0.0.0.0:%d", port)
	}
	return fmt.Sprintf("127.0.0.1:%d", port)
}

// registerRoutes wires up all API endpoints and the embedded SPA.
func (s *Server) registerRoutes() {
	// API routes
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("GET /api/contacts", s.handleListContacts)
	s.mux.HandleFunc("GET /api/contacts/{name}", s.handleGetContact)
	s.mux.HandleFunc("POST /api/contacts/{name}/ideas", s.handleAddIdea)
	s.mux.HandleFunc("POST /api/contacts/{name}/gifts/given", s.handleAddGiven)
	s.mux.HandleFunc("POST /api/contacts/{name}/gifts/received", s.handleAddReceived)
	s.mux.HandleFunc("POST /api/contacts/{name}/archive", s.handleArchiveContact)
	s.mux.HandleFunc("POST /api/contacts/{name}/unarchive", s.handleUnarchiveContact)
	s.mux.HandleFunc("PUT /api/contacts/{name}/preferences", s.handleSetPreferences)
	s.mux.HandleFunc("GET /api/ideas", s.handleGetGeneralIdeas)
	s.mux.HandleFunc("POST /api/ideas", s.handleAddGeneralIdea)
	s.mux.HandleFunc("PUT /api/contacts/{name}/ideas/{id}", s.handleEditContactIdea)
	s.mux.HandleFunc("PUT /api/contacts/{name}/gifts/{id}", s.handleEditGift)
	s.mux.HandleFunc("PUT /api/ideas/{id}", s.handleEditGeneralIdea)
	s.mux.HandleFunc("DELETE /api/ideas/{id}", s.handleDeleteGeneralIdea)
	s.mux.HandleFunc("DELETE /api/contacts/{name}/ideas/{id}", s.handleDeleteContactIdea)

	// SPA static files (embedded)
	s.mux.HandleFunc("GET /", s.handleSPA)
	s.mux.HandleFunc("POST /login", s.handleLoginSubmit)
}

// --- Auth middleware ---

func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if s.authToken == "" {
			// No auth configured — localhost only enforced by bind address.
			next.ServeHTTP(w, r)
			return
		}

		// Check Bearer token
		auth := r.Header.Get("Authorization")
		if strings.HasPrefix(auth, "Bearer ") {
			if strings.TrimPrefix(auth, "Bearer ") == s.authToken {
				next.ServeHTTP(w, r)
				return
			}
		}

		// Check ?token= query param (for iOS Shortcuts)
		if q := r.URL.Query().Get("token"); q == s.authToken {
			next.ServeHTTP(w, r)
			return
		}

		// Check cookie (for browser sessions)
		if c, err := r.Cookie("pressie_auth"); err == nil && c.Value == s.authToken {
			next.ServeHTTP(w, r)
			return
		}

		// Allow POST /login through (the login form submit handler authenticates itself).
		if r.Method == "POST" && r.URL.Path == "/login" {
			next.ServeHTTP(w, r)
			return
		}

		// Login page for browser requests, 401 for API
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.Error(w, `{"error":"unauthorized"}`, http.StatusUnauthorized)
			return
		}
		s.serveLogin(w, r)
	})
}

// --- Request/response helpers ---

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func decodeBody(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// --- API Handlers ---

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":   "ok",
		"version":  "0.1.0-dev",
		"giftsDir": s.giftsDir,
	})
}

func (s *Server) handleListContacts(w http.ResponseWriter, r *http.Request) {
	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	type contactSummary struct {
		Key        string   `json:"key"`
		Name       string   `json:"name"`
		File       string   `json:"file"`
		Visibility string   `json:"visibility"`
		Tags       []string `json:"tags,omitempty"`
		Archived   bool     `json:"archived,omitempty"`
	}

	includeArchived := r.URL.Query().Get("archived") == "true"
	contacts := make([]contactSummary, 0, len(idx.Contacts))
	for key, m := range idx.Contacts {
		if m.Archived && !includeArchived {
			continue
		}
		name := s.lookupName(m.File)
		contacts = append(contacts, contactSummary{
			Key:        key,
			Name:       name,
			File:       m.File,
			Visibility: m.Visibility,
			Tags:       m.Tags,
			Archived:   m.Archived,
		})
	}

	sort.Slice(contacts, func(i, j int) bool { return contacts[i].Name < contacts[j].Name })
	writeJSON(w, http.StatusOK, contacts)
}

func (s *Server) handleGetContact(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := store.LoadContactFile(filepath.Join(s.giftsDir, relPath))
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no contact file for %s", displayName))
		return
	}
	m, hasMapping := idx.Contacts[key]
	archived := hasMapping && m.Archived
	writeJSON(w, http.StatusOK, struct {
		*store.ContactFile
		ContactKey string `json:"contact_key"`
		Archived   bool   `json:"archived"`
	}{cf, key, archived})
}

// addIdeaRequest is the body for adding an idea to a contact.
type addIdeaRequest struct {
	Item  string   `json:"item"`
	URL   string   `json:"url,omitempty"`
	Tags  []string `json:"tags,omitempty"`
	Notes string   `json:"notes,omitempty"`
	Price *float64 `json:"price,omitempty"`
}

func (s *Server) handleAddIdea(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req addIdeaRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Item == "" {
		writeError(w, http.StatusBadRequest, "item is required")
		return
	}

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := gifts.EnsureContactFile(s.giftsDir, relPath, key, displayName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	idea := store.Idea{
		ID:     store.NewUUID(),
		Item:   req.Item,
		URL:    req.URL,
		Tags:   req.Tags,
		Status: "open",
		Added:  time.Now().UTC().Format("2006-01-02"),
		Notes:  req.Notes,
		Images: []string{},
	}
	if req.Price != nil {
		idea.PriceEstimate = req.Price
		idea.Currency = "USD"
	}

	cf.Ideas = append(cf.Ideas, idea)
	cf.Tags = mergeTagsWeb(cf.Tags, idea.Tags)

	absPath := filepath.Join(s.giftsDir, relPath)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gifts.RegisterContact(idx, key, relPath, "private", cf.Tags)
	store.SaveIndex(s.giftsDir, idx)

	writeJSON(w, http.StatusCreated, idea)
}

func (s *Server) handleAddGiven(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req addGiftRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Item == "" {
		writeError(w, http.StatusBadRequest, "item is required")
		return
	}

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := gifts.EnsureContactFile(s.giftsDir, relPath, key, displayName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	date := req.Date
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	gift := store.Gift{
		ID:       store.NewUUID(),
		Date:     date,
		Occasion: req.Occasion,
		Item:     req.Item,
		Notes:    req.Notes,
		Images:   []string{},
		Source:   "web",
		Added:    time.Now().UTC().Format(time.RFC3339),
	}
	if req.Price != nil {
		gift.Price = req.Price
		gift.Currency = "USD"
	}

	cf.GiftsGiven = append(cf.GiftsGiven, gift)

	// Retire matching ideas
	for i := range cf.Ideas {
		if cf.Ideas[i].Status == "open" && itemsMatchWeb(cf.Ideas[i].Item, gift.Item) {
			cf.Ideas[i].Status = "given"
		}
	}

	absPath := filepath.Join(s.giftsDir, relPath)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gifts.RegisterContact(idx, key, relPath, "private", nil)
	store.SaveIndex(s.giftsDir, idx)

	writeJSON(w, http.StatusCreated, gift)
}

func (s *Server) handleAddReceived(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req addGiftRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Item == "" {
		writeError(w, http.StatusBadRequest, "item is required")
		return
	}

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := gifts.EnsureContactFile(s.giftsDir, relPath, key, displayName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	date := req.Date
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	}

	gift := store.Gift{
		ID:       store.NewUUID(),
		Date:     date,
		Occasion: req.Occasion,
		Item:     req.Item,
		Notes:    req.Notes,
		Images:   []string{},
		Source:   "web",
		Added:    time.Now().UTC().Format(time.RFC3339),
	}
	if req.Price != nil {
		gift.Price = req.Price
		gift.Currency = "USD"
	}

	cf.GiftsReceived = append(cf.GiftsReceived, gift)

	absPath := filepath.Join(s.giftsDir, relPath)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gifts.RegisterContact(idx, key, relPath, "private", nil)
	store.SaveIndex(s.giftsDir, idx)

	writeJSON(w, http.StatusCreated, gift)
}

// addGiftRequest is the body for adding a gift (given or received).
type addGiftRequest struct {
	Item     string   `json:"item"`
	Occasion string   `json:"occasion,omitempty"`
	Date     string   `json:"date,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	Notes    string   `json:"notes,omitempty"`
}

func (s *Server) handleSetPreferences(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var req struct {
		Preferences string `json:"preferences"`
	}
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := gifts.EnsureContactFile(s.giftsDir, relPath, key, displayName)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	cf.Preferences = req.Preferences

	absPath := filepath.Join(s.giftsDir, relPath)
	if err := store.SaveContactFile(absPath, cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	gifts.RegisterContact(idx, key, relPath, "private", nil)
	store.SaveIndex(s.giftsDir, idx)

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleGetGeneralIdeas(w http.ResponseWriter, r *http.Request) {
	ideas, err := store.LoadGeneralIdeas(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, ideas)
}

func (s *Server) handleAddGeneralIdea(w http.ResponseWriter, r *http.Request) {
	var req addIdeaRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if req.Item == "" {
		writeError(w, http.StatusBadRequest, "item is required")
		return
	}

	idea := store.Idea{
		ID:     store.NewUUID(),
		Item:   req.Item,
		URL:    req.URL,
		Tags:   req.Tags,
		Status: "open",
		Added:  time.Now().UTC().Format("2006-01-02"),
		Notes:  req.Notes,
		Images: []string{},
	}
	if req.Price != nil {
		idea.PriceEstimate = req.Price
		idea.Currency = "USD"
	}

	ideas, err := store.LoadGeneralIdeas(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	ideas = append(ideas, idea)
	if err := store.SaveGeneralIdeas(s.giftsDir, ideas); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusCreated, idea)
}

func (s *Server) handleDeleteGeneralIdea(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	ideas, err := store.LoadGeneralIdeas(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	found := false
	filtered := make([]store.Idea, 0, len(ideas))
	for _, idea := range ideas {
		if idea.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, idea)
	}

	if !found {
		writeError(w, http.StatusNotFound, "idea not found")
		return
	}

	if err := store.SaveGeneralIdeas(s.giftsDir, filtered); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

func (s *Server) handleDeleteContactIdea(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	id := r.PathValue("id")

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := store.LoadContactFile(filepath.Join(s.giftsDir, relPath))
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no contact file for %s", displayName))
		return
	}

	found := false
	filtered := make([]store.Idea, 0, len(cf.Ideas))
	for _, idea := range cf.Ideas {
		if idea.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, idea)
	}

	if !found {
		writeError(w, http.StatusNotFound, "idea not found")
		return
	}

	cf.Ideas = filtered
	cf.Tags = recomputeTagsWeb(filtered)

	if err := store.SaveContactFile(filepath.Join(s.giftsDir, relPath), cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "deleted"})
}

// editIdeaRequest is the body for editing an idea. Pointer fields: nil
// means "leave unchanged", non-nil (including empty) means "set".
type editIdeaRequest struct {
	Item  *string   `json:"item,omitempty"`
	URL   *string   `json:"url,omitempty"`
	Tags  *[]string `json:"tags,omitempty"`
	Notes *string   `json:"notes,omitempty"`
	Price *float64  `json:"price,omitempty"`
}

func applyIdeaEdit(idea *store.Idea, req editIdeaRequest) {
	if req.Item != nil {
		idea.Item = *req.Item
	}
	if req.URL != nil {
		idea.URL = *req.URL
	}
	if req.Tags != nil {
		idea.Tags = *req.Tags
	}
	if req.Notes != nil {
		idea.Notes = *req.Notes
	}
	if req.Price != nil {
		idea.PriceEstimate = req.Price
		if *req.Price > 0 && idea.Currency == "" {
			idea.Currency = "USD"
		}
	}
}

func (s *Server) handleEditContactIdea(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	id := r.PathValue("id")

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := store.LoadContactFile(filepath.Join(s.giftsDir, relPath))
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no contact file for %s", displayName))
		return
	}

	found := false
	for i := range cf.Ideas {
		if cf.Ideas[i].ID == id {
			var req editIdeaRequest
			if err := decodeBody(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			applyIdeaEdit(&cf.Ideas[i], req)
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "idea not found")
		return
	}

	cf.Tags = recomputeTagsWeb(cf.Ideas)
	if err := store.SaveContactFile(filepath.Join(s.giftsDir, relPath), cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	if m, ok := idx.Contacts[key]; ok {
		m.Tags = cf.Tags
		idx.Contacts[key] = m
		store.SaveIndex(s.giftsDir, idx)
	}

	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

func (s *Server) handleEditGeneralIdea(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	ideas, err := store.LoadGeneralIdeas(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	found := false
	for i := range ideas {
		if ideas[i].ID == id {
			var req editIdeaRequest
			if err := decodeBody(r, &req); err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			applyIdeaEdit(&ideas[i], req)
			found = true
			break
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "idea not found")
		return
	}

	if err := store.SaveGeneralIdeas(s.giftsDir, ideas); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// editGiftRequest is the body for editing a logged gift.
type editGiftRequest struct {
	Item     *string  `json:"item,omitempty"`
	Occasion *string  `json:"occasion,omitempty"`
	Date     *string  `json:"date,omitempty"`
	Price    *float64 `json:"price,omitempty"`
	Notes    *string  `json:"notes,omitempty"`
}

func applyGiftEditWeb(gift *store.Gift, req editGiftRequest) {
	if req.Item != nil {
		gift.Item = *req.Item
	}
	if req.Occasion != nil {
		gift.Occasion = *req.Occasion
	}
	if req.Date != nil {
		gift.Date = *req.Date
	}
	if req.Price != nil {
		gift.Price = req.Price
		if *req.Price > 0 && gift.Currency == "" {
			gift.Currency = "USD"
		}
	}
	if req.Notes != nil {
		gift.Notes = *req.Notes
	}
}

func (s *Server) handleEditGift(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	id := r.PathValue("id")

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	_, displayName, relPath, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	cf, err := store.LoadContactFile(filepath.Join(s.giftsDir, relPath))
	if err != nil {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no contact file for %s", displayName))
		return
	}

	var req editGiftRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	found := false
	for i := range cf.GiftsGiven {
		if cf.GiftsGiven[i].ID == id {
			applyGiftEditWeb(&cf.GiftsGiven[i], req)
			found = true
			break
		}
	}
	if !found {
		for i := range cf.GiftsReceived {
			if cf.GiftsReceived[i].ID == id {
				applyGiftEditWeb(&cf.GiftsReceived[i], req)
				found = true
				break
			}
		}
	}
	if !found {
		writeError(w, http.StatusNotFound, "gift not found")
		return
	}

	if err := store.SaveContactFile(filepath.Join(s.giftsDir, relPath), cf); err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// handleArchiveContact marks a contact archived in the index.
func (s *Server) handleArchiveContact(w http.ResponseWriter, r *http.Request) {
	s.setContactArchived(w, r, true)
}

// handleUnarchiveContact clears a contact's archived flag.
func (s *Server) handleUnarchiveContact(w http.ResponseWriter, r *http.Request) {
	s.setContactArchived(w, r, false)
}

func (s *Server) setContactArchived(w http.ResponseWriter, r *http.Request, archived bool) {
	name := r.PathValue("name")

	idx, err := store.LoadIndex(s.giftsDir)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}

	key, displayName, _, err := gifts.ResolveContact(context.Background(), s.giftsDir, idx, name, "private")
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	m, ok := idx.Contacts[key]
	if !ok {
		writeError(w, http.StatusNotFound, fmt.Sprintf("no index entry for %s", displayName))
		return
	}

	if m.Archived != archived {
		m.Archived = archived
		idx.Contacts[key] = m
		if err := store.SaveIndex(s.giftsDir, idx); err != nil {
			writeError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{"status": "ok", "archived": archived, "name": displayName})
}

// --- Helpers ---

// lookupName loads a contact file's Name field by relative path.
func (s *Server) lookupName(relPath string) string {
	cf, err := store.LoadContactFile(filepath.Join(s.giftsDir, relPath))
	if err != nil {
		return ""
	}
	return cf.Name
}

// mergeTagsWeb is the same merge logic as the CLI's mergeTags.
func mergeTagsWeb(existing, additions []string) []string {
	seen := make(map[string]bool, len(existing)+len(additions))
	merged := make([]string, 0, len(existing)+len(additions))
	for _, t := range existing {
		k := strings.ToLower(t)
		if !seen[k] {
			seen[k] = true
			merged = append(merged, t)
		}
	}
	for _, t := range additions {
		k := strings.ToLower(t)
		if !seen[k] {
			seen[k] = true
			merged = append(merged, t)
		}
	}
	return merged
}

// itemsMatchWeb is the same fuzzy match as the CLI's itemsMatch.
func itemsMatchWeb(a, b string) bool {
	na := normalizeItemWeb(a)
	nb := normalizeItemWeb(b)
	if na == nb {
		return true
	}
	if na == "" || nb == "" {
		return false
	}
	return strings.Contains(na, nb) || strings.Contains(nb, na)
}

// normalizeItemWeb lowercases, trims, and collapses whitespace.
func normalizeItemWeb(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	for strings.Contains(s, "  ") {
		s = strings.ReplaceAll(s, "  ", " ")
	}
	return s
}

// recomputeTagsWeb recomputes contact tags from remaining ideas.
func recomputeTagsWeb(ideas []store.Idea) []string {
	seen := make(map[string]bool)
	tags := make([]string, 0)
	for _, idea := range ideas {
		for _, t := range idea.Tags {
			k := strings.ToLower(t)
			if !seen[k] {
				seen[k] = true
				tags = append(tags, t)
			}
		}
	}
	return tags
}

// ensure unused imports don't break build
var _ = os.Stat
var _ = config.DefaultGiftsDir
