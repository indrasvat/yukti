package components

import (
	"testing"

	"yukti/internal/domain/project"
)

func TestFuzzyScore(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		pattern  string
		wantZero bool // true if score should be 0
		wantHigh bool // true if score should be high (exact/substring)
	}{
		{
			name:     "empty pattern matches everything",
			text:     "hello",
			pattern:  "",
			wantZero: false,
		},
		{
			name:     "empty text matches nothing",
			text:     "",
			pattern:  "abc",
			wantZero: true,
		},
		{
			name:     "exact match",
			text:     "hello",
			pattern:  "hello",
			wantHigh: true,
		},
		{
			name:     "substring match",
			text:     "hello world",
			pattern:  "world",
			wantHigh: true,
		},
		{
			name:     "fuzzy match in order",
			text:     "getfiles",
			pattern:  "gf",
			wantZero: false,
		},
		{
			name:     "fuzzy match fails out of order",
			text:     "getfiles",
			pattern:  "fg",
			wantZero: true,
		},
		{
			name:     "word boundary bonus",
			text:     "get_files",
			pattern:  "gf",
			wantZero: false,
		},
		{
			name:     "consecutive char bonus",
			text:     "myfunction",
			pattern:  "func",
			wantZero: false,
		},
		{
			name:     "no match",
			text:     "hello",
			pattern:  "xyz",
			wantZero: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := fuzzyScore(tt.text, tt.pattern)

			if tt.wantZero && score != 0 {
				t.Errorf("fuzzyScore(%q, %q) = %d, want 0", tt.text, tt.pattern, score)
			}
			if !tt.wantZero && score == 0 {
				t.Errorf("fuzzyScore(%q, %q) = 0, want non-zero", tt.text, tt.pattern)
			}
			if tt.wantHigh && score < 100 {
				t.Errorf("fuzzyScore(%q, %q) = %d, want >= 100 for exact/substring match", tt.text, tt.pattern, score)
			}
		})
	}
}

func TestFuzzyScore_Ranking(t *testing.T) {
	// Test that scores produce correct ranking order
	tests := []struct {
		name   string
		text1  string
		text2  string
		query  string
		higher string // which text should score higher
	}{
		{
			name:   "exact beats fuzzy",
			text1:  "code",
			text2:  "codeviewer",
			query:  "code",
			higher: "code",
		},
		{
			name:   "prefix beats middle",
			text1:  "codeviewer",
			text2:  "mycodeviewer",
			query:  "code",
			higher: "codeviewer",
		},
		{
			name:   "word boundary beats middle of word",
			text1:  "get_files",
			text2:  "getfiles",
			query:  "gf",
			higher: "get_files", // g and f are at word boundaries
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score1 := fuzzyScore(tt.text1, tt.query)
			score2 := fuzzyScore(tt.text2, tt.query)

			if tt.higher == tt.text1 && score1 <= score2 {
				t.Errorf("expected %q (score=%d) > %q (score=%d) for query %q",
					tt.text1, score1, tt.text2, score2, tt.query)
			}
			if tt.higher == tt.text2 && score2 <= score1 {
				t.Errorf("expected %q (score=%d) > %q (score=%d) for query %q",
					tt.text2, score2, tt.text1, score1, tt.query)
			}
		})
	}
}

func TestFindFunctionLine(t *testing.T) {
	tests := []struct {
		name     string
		source   string
		funcName string
		wantLine int
	}{
		{
			name:     "function at start",
			source:   "function myFunc() {\n  return 1;\n}",
			funcName: "myFunc",
			wantLine: 1,
		},
		{
			name:     "function in middle",
			source:   "// comment\nvar x = 1;\n\nfunction doSomething() {\n  return x;\n}",
			funcName: "doSomething",
			wantLine: 4,
		},
		{
			name:     "function not found",
			source:   "var x = 1;",
			funcName: "missing",
			wantLine: 1, // defaults to 1
		},
		{
			name:     "multiple functions",
			source:   "function first() {}\nfunction second() {}\nfunction third() {}",
			funcName: "second",
			wantLine: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findFunctionLine(tt.source, tt.funcName)
			if got != tt.wantLine {
				t.Errorf("findFunctionLine() = %d, want %d", got, tt.wantLine)
			}
		})
	}
}

func TestFuzzyFinder_SetItems(t *testing.T) {
	ff := NewFuzzyFinder()

	files := []project.File{
		{
			Name:   "Code",
			Type:   project.FileTypeServer,
			Source: "function myFunc() {}\nfunction otherFunc() {}",
			FunctionSet: &project.FunctionSet{
				Functions: []project.Function{
					{Name: "myFunc"},
					{Name: "otherFunc"},
				},
			},
		},
		{
			Name:   "index",
			Type:   project.FileTypeHTML,
			Source: "<html></html>",
		},
		{
			Name:   "appsscript",
			Type:   project.FileTypeJSON,
			Source: "{}",
		},
	}

	ff.SetItems(files)

	// Should have 3 files + 2 functions = 5 items
	if len(ff.items) != 5 {
		t.Errorf("SetItems() created %d items, want 5", len(ff.items))
	}

	// Check file icons
	icons := make(map[string]string)
	for _, item := range ff.items {
		if item.Function == nil {
			icons[item.Title] = item.Icon
		}
	}

	if icons["Code"] != "📄" {
		t.Errorf("Server file icon = %q, want 📄", icons["Code"])
	}
	if icons["index"] != "🌐" {
		t.Errorf("HTML file icon = %q, want 🌐", icons["index"])
	}
	if icons["appsscript"] != "📋" {
		t.Errorf("JSON file icon = %q, want 📋", icons["appsscript"])
	}

	// Check functions have ƒ icon
	for _, item := range ff.items {
		if item.Function != nil && item.Icon != "ƒ" {
			t.Errorf("Function icon = %q, want ƒ", item.Icon)
		}
	}
}

func TestFuzzyFinder_Filter(t *testing.T) {
	ff := NewFuzzyFinder()

	files := []project.File{
		{
			Name:   "Code",
			Type:   project.FileTypeServer,
			Source: "function listFiles() {}",
			FunctionSet: &project.FunctionSet{
				Functions: []project.Function{
					{Name: "listFiles"},
				},
			},
		},
		{
			Name:   "Utils",
			Type:   project.FileTypeServer,
			Source: "function formatDate() {}",
			FunctionSet: &project.FunctionSet{
				Functions: []project.Function{
					{Name: "formatDate"},
				},
			},
		},
	}

	ff.SetItems(files)

	// Initially all items visible
	if len(ff.filtered) != 4 { // 2 files + 2 functions
		t.Errorf("Initial filtered count = %d, want 4", len(ff.filtered))
	}

	// Filter for "list"
	ff.input.SetValue("list")
	ff.filter()

	// Should match "listFiles" function and possibly file
	if len(ff.filtered) == 0 {
		t.Error("filter('list') returned no results, expected matches")
	}

	// Check that listFiles is in results
	found := false
	for _, item := range ff.filtered {
		if item.Title == "listFiles()" {
			found = true
			break
		}
	}
	if !found {
		t.Error("filter('list') should include listFiles() function")
	}

	// Filter for something that doesn't exist
	ff.input.SetValue("xyz123")
	ff.filter()

	if len(ff.filtered) != 0 {
		t.Errorf("filter('xyz123') returned %d results, want 0", len(ff.filtered))
	}

	// Empty filter shows all
	ff.input.SetValue("")
	ff.filter()

	if len(ff.filtered) != 4 {
		t.Errorf("filter('') returned %d results, want 4", len(ff.filtered))
	}
}

func TestFuzzyFinder_Visibility(t *testing.T) {
	ff := NewFuzzyFinder()

	if ff.IsVisible() {
		t.Error("FuzzyFinder should start hidden")
	}

	ff.Show()

	if !ff.IsVisible() {
		t.Error("FuzzyFinder should be visible after Show()")
	}

	ff.Hide()

	if ff.IsVisible() {
		t.Error("FuzzyFinder should be hidden after Hide()")
	}
}

func TestFuzzyFinder_ShowResetsState(t *testing.T) {
	ff := NewFuzzyFinder()

	files := []project.File{
		{Name: "Code", Type: project.FileTypeServer, Source: ""},
		{Name: "Utils", Type: project.FileTypeServer, Source: ""},
	}
	ff.SetItems(files)

	// Set some state
	ff.input.SetValue("test")
	ff.selected = 1
	ff.filter()

	// Show should reset
	ff.Show()

	if ff.input.Value() != "" {
		t.Errorf("Show() should clear input, got %q", ff.input.Value())
	}
	if ff.selected != 0 {
		t.Errorf("Show() should reset selected to 0, got %d", ff.selected)
	}
	if len(ff.filtered) != 2 {
		t.Errorf("Show() should reset filtered to all items, got %d", len(ff.filtered))
	}
}
