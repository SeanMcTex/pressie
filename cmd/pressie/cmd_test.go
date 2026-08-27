package main

import (
	"testing"

	"github.com/SeanMcTex/pressie/internal/store"
)

func TestNormalizeItem(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Letterpress print", "letterpress print"},
		{"  Wool  Blanket  ", "wool blanket"},
		{"Irish   wool   scarf", "irish wool scarf"},
		{"", ""},
		{"UPPER", "upper"},
	}
	for _, tc := range tests {
		got := normalizeItem(tc.input)
		if got != tc.want {
			t.Errorf("normalizeItem(%q) = %q, want %q", tc.input, got, tc.want)
		}
	}
}

func TestItemsMatch_Exact(t *testing.T) {
	if !itemsMatch("Letterpress print", "Letterpress print") {
		t.Error("exact match should return true")
	}
}

func TestItemsMatch_CaseInsensitive(t *testing.T) {
	if !itemsMatch("Letterpress Print", "letterpress print") {
		t.Error("case-insensitive exact match should return true")
	}
}

func TestItemsMatch_SubstringAContainsB(t *testing.T) {
	if !itemsMatch("Irish wool scarf", "Irish wool") {
		t.Error("substring match (a contains b) should return true")
	}
}

func TestItemsMatch_SubstringBContainsA(t *testing.T) {
	if !itemsMatch("Irish wool", "Irish wool scarf") {
		t.Error("substring match (b contains a) should return true")
	}
}

func TestItemsMatch_NoMatch(t *testing.T) {
	if itemsMatch("Letterpress print", "Wool scarf") {
		t.Error("unrelated items should not match")
	}
}

func TestItemsMatch_EmptyString(t *testing.T) {
	if itemsMatch("", "something") {
		t.Error("empty string should not match non-empty")
	}
	if itemsMatch("something", "") {
		t.Error("non-empty should not match empty")
	}
}

func TestItemsMatch_WhitespaceNormalization(t *testing.T) {
	if !itemsMatch("wool  Blanket", "Wool blanket") {
		t.Error("whitespace differences should be normalized")
	}
}

func TestMergeTags_Union(t *testing.T) {
	result := mergeTags([]string{"art", "irish"}, []string{"irish", "warm"})
	want := []string{"art", "irish", "warm"}
	if len(result) != len(want) {
		t.Fatalf("len = %d, want %d", len(result), len(want))
	}
	for i, w := range want {
		if result[i] != w {
			t.Errorf("result[%d] = %q, want %q", i, result[i], w)
		}
	}
}

func TestMergeTags_CaseInsensitiveDedup(t *testing.T) {
	result := mergeTags([]string{"Art", "Irish"}, []string{"art", "music"})
	if len(result) != 3 {
		t.Fatalf("len = %d, want 3 (case-insensitive dedup)", len(result))
	}
}

func TestMergeTags_EmptyExisting(t *testing.T) {
	result := mergeTags(nil, []string{"art", "irish"})
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
}

func TestMergeTags_EmptyAdditions(t *testing.T) {
	result := mergeTags([]string{"art"}, nil)
	if len(result) != 1 || result[0] != "art" {
		t.Errorf("result = %v, want [art]", result)
	}
}

func TestTagsIntersect_HasIntersection(t *testing.T) {
	if !tagsIntersect([]string{"art", "irish"}, []string{"irish", "warm"}) {
		t.Error("should intersect on 'irish'")
	}
}

func TestTagsIntersect_NoIntersection(t *testing.T) {
	if tagsIntersect([]string{"art"}, []string{"music"}) {
		t.Error("should not intersect")
	}
}

func TestTagsIntersect_CaseInsensitive(t *testing.T) {
	if !tagsIntersect([]string{"Art"}, []string{"art"}) {
		t.Error("should intersect case-insensitively")
	}
}

func TestTagsIntersect_EmptyTags(t *testing.T) {
	if tagsIntersect([]string{}, []string{"art"}) {
		t.Error("empty tags should not intersect")
	}
}

func TestFilterByStatus_Open(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "A", Status: "open"},
		{ID: "2", Item: "B", Status: "purchased"},
		{ID: "3", Item: "C", Status: "open"},
	}
	result := filterByStatus(ideas, "open")
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
	for _, idea := range result {
		if idea.Status != "open" {
			t.Errorf("Status = %q, want open", idea.Status)
		}
	}
}

func TestFilterByStatus_Purchased(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "A", Status: "open"},
		{ID: "2", Item: "B", Status: "purchased"},
	}
	result := filterByStatus(ideas, "purchased")
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1", len(result))
	}
	if result[0].ID != "2" {
		t.Errorf("ID = %q, want 2", result[0].ID)
	}
}

func TestFilterByStatus_EmptyStatusReturnsAll(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Status: "open"},
		{ID: "2", Status: "purchased"},
	}
	result := filterByStatus(ideas, "")
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2", len(result))
	}
}

func TestFilterDuplicates_RemovesMatching(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "Letterpress print"},
		{ID: "2", Item: "Irish wool scarf"},
		{ID: "3", Item: "Ceramic set"},
	}
	gifts := []store.Gift{
		{Item: "Letterpress print"},
	}
	result := filterDuplicates(ideas, gifts)
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2 (Letterpress filtered)", len(result))
	}
	for _, idea := range result {
		if idea.Item == "Letterpress print" {
			t.Error("Letterpress print should have been filtered")
		}
	}
}

func TestFilterDuplicates_SubstringMatch(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "Irish wool scarf"},
		{ID: "2", Item: "Irish wool sweater"},
		{ID: "3", Item: "Ceramic set"},
	}
	gifts := []store.Gift{
		{Item: "Irish wool"},
	}
	result := filterDuplicates(ideas, gifts)
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1 (both Irish wool items filtered)", len(result))
	}
	if result[0].Item != "Ceramic set" {
		t.Errorf("Item = %q, want %q", result[0].Item, "Ceramic set")
	}
}

func TestFilterDuplicates_NoGiftsGiven(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "A"},
		{ID: "2", Item: "B"},
	}
	result := filterDuplicates(ideas, nil)
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2 (no filtering with empty gifts)", len(result))
	}
}

func TestFilterDuplicates_CaseInsensitive(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "Wool Blanket"},
	}
	gifts := []store.Gift{
		{Item: "wool blanket"},
	}
	result := filterDuplicates(ideas, gifts)
	if len(result) != 0 {
		t.Fatalf("len = %d, want 0 (case-insensitive match should filter)", len(result))
	}
}

func TestFilterDuplicates_NoMatch(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Item: "Letterpress print"},
	}
	gifts := []store.Gift{
		{Item: "Wool scarf"},
	}
	result := filterDuplicates(ideas, gifts)
	if len(result) != 1 {
		t.Fatalf("len = %d, want 1 (no match)", len(result))
	}
}

func TestFilterGifts_NoYearFilter(t *testing.T) {
	gifts := []store.Gift{
		{ID: "1", Date: "2025-12-25", Item: "A"},
		{ID: "2", Date: "2024-01-15", Item: "B"},
	}
	result := filterGifts(gifts, "")
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2 (no year filter)", len(result))
	}
}

func TestFilterGifts_YearMatch(t *testing.T) {
	gifts := []store.Gift{
		{ID: "1", Date: "2025-12-25", Item: "A"},
		{ID: "2", Date: "2024-01-15", Item: "B"},
		{ID: "3", Date: "2025-06-10", Item: "C"},
	}
	result := filterGifts(gifts, "2025")
	if len(result) != 2 {
		t.Fatalf("len = %d, want 2 (2025 gifts)", len(result))
	}
	for _, g := range result {
		if g.Date[:4] != "2025" {
			t.Errorf("Date = %q, want 2025-*", g.Date)
		}
	}
}

func TestFilterGifts_YearNoMatch(t *testing.T) {
	gifts := []store.Gift{
		{ID: "1", Date: "2024-12-25", Item: "A"},
	}
	result := filterGifts(gifts, "2025")
	if len(result) != 0 {
		t.Fatalf("len = %d, want 0 (no 2025 gifts)", len(result))
	}
}

func TestFilterGifts_EmptyList(t *testing.T) {
	result := filterGifts(nil, "2025")
	if len(result) != 0 {
		t.Fatalf("len = %d, want 0", len(result))
	}
}

// retireIdeaByID tests the ID-based retirement logic used by add-given --idea.
func TestRetireIdeaByID(t *testing.T) {
	ideas := []store.Idea{
		{ID: "idea-1", Item: "Letterpress print", Status: "open"},
		{ID: "idea-2", Item: "Wool scarf", Status: "open"},
		{ID: "idea-3", Item: "Ceramic set", Status: "open"},
	}
	targetID := "idea-2"
	retired := 0
	for i := range ideas {
		if ideas[i].ID == targetID && ideas[i].Status == "open" {
			ideas[i].Status = "purchased"
			retired++
		}
	}
	if retired != 1 {
		t.Fatalf("retired = %d, want 1", retired)
	}
	if ideas[1].Status != "purchased" {
		t.Errorf("idea-2 status = %q, want purchased", ideas[1].Status)
	}
	if ideas[0].Status != "open" {
		t.Errorf("idea-1 status = %q, want open (should not be retired)", ideas[0].Status)
	}
	if ideas[2].Status != "open" {
		t.Errorf("idea-3 status = %q, want open (should not be retired)", ideas[2].Status)
	}
}

func TestRetireIdeaByID_NotFound(t *testing.T) {
	ideas := []store.Idea{
		{ID: "idea-1", Status: "open"},
	}
	retired := 0
	for i := range ideas {
		if ideas[i].ID == "nonexistent" && ideas[i].Status == "open" {
			ideas[i].Status = "purchased"
			retired++
		}
	}
	if retired != 0 {
		t.Fatalf("retired = %d, want 0 (ID not found)", retired)
	}
	if ideas[0].Status != "open" {
		t.Errorf("status = %q, want open", ideas[0].Status)
	}
}

func TestRetireIdeaByID_SkipsNonOpen(t *testing.T) {
	ideas := []store.Idea{
		{ID: "idea-1", Status: "purchased"},
	}
	retired := 0
	for i := range ideas {
		if ideas[i].ID == "idea-1" && ideas[i].Status == "open" {
			ideas[i].Status = "purchased"
			retired++
		}
	}
	if retired != 0 {
		t.Fatalf("retired = %d, want 0 (already purchased)", retired)
	}
}

func TestRecomputeTags_Union(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Tags: []string{"art", "irish"}},
		{ID: "2", Tags: []string{"irish", "warm"}},
		{ID: "3", Tags: []string{"kitchen"}},
	}
	tags := recomputeTags(ideas)
	want := []string{"art", "irish", "warm", "kitchen"}
	if len(tags) != len(want) {
		t.Fatalf("len = %d, want %d", len(tags), len(want))
	}
	for i, w := range want {
		if tags[i] != w {
			t.Errorf("tags[%d] = %q, want %q", i, tags[i], w)
		}
	}
}

func TestRecomputeTags_CaseInsensitiveDedup(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Tags: []string{"Art"}},
		{ID: "2", Tags: []string{"art", "music"}},
	}
	tags := recomputeTags(ideas)
	if len(tags) != 2 {
		t.Fatalf("len = %d, want 2 (case-insensitive dedup)", len(tags))
	}
	if tags[0] != "Art" {
		t.Errorf("tags[0] = %q, want %q (first-seen preserved)", tags[0], "Art")
	}
}

func TestRecomputeTags_EmptyIdeas(t *testing.T) {
	tags := recomputeTags(nil)
	if len(tags) != 0 {
		t.Fatalf("len = %d, want 0", len(tags))
	}
}

func TestRecomputeTags_IdeasWithNoTags(t *testing.T) {
	ideas := []store.Idea{
		{ID: "1", Tags: nil},
		{ID: "2", Tags: []string{}},
		{ID: "3", Tags: []string{"art"}},
	}
	tags := recomputeTags(ideas)
	if len(tags) != 1 || tags[0] != "art" {
		t.Errorf("tags = %v, want [art]", tags)
	}
}

func TestRecomputeTags_AfterDeletion(t *testing.T) {
	// Simulate: had 3 ideas, deleted idea 2 (which had "music" tag).
	// Tags from remaining ideas should not include "music".
	remaining := []store.Idea{
		{ID: "1", Tags: []string{"art", "irish"}},
		{ID: "3", Tags: []string{"kitchen"}},
	}
	tags := recomputeTags(remaining)
	for _, tag := range tags {
		if tag == "music" {
			t.Error("tag 'music' should have been pruned after deletion")
		}
	}
	if len(tags) != 3 {
		t.Errorf("len = %d, want 3 (art, irish, kitchen)", len(tags))
	}
}