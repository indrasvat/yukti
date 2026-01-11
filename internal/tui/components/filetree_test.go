package components

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"yukti/internal/domain/project"
)

func TestFileTree_NewFileTree(t *testing.T) {
	files := []project.File{
		{Name: "Code", Type: project.FileTypeServer},
		{Name: "index", Type: project.FileTypeHTML},
	}

	ft := NewFileTree(files)

	if ft == nil {
		t.Fatal("NewFileTree returned nil")
	}

	if len(ft.files) != 2 {
		t.Errorf("NewFileTree files = %d, want 2", len(ft.files))
	}

	if ft.selected != 0 {
		t.Errorf("NewFileTree selected = %d, want 0", ft.selected)
	}
}

func TestFileTree_SetFiles(t *testing.T) {
	ft := NewFileTree(nil)

	files := []project.File{
		{Name: "Code", Type: project.FileTypeServer},
		{Name: "Utils", Type: project.FileTypeServer},
		{Name: "index", Type: project.FileTypeHTML},
	}

	ft.SetFiles(files)

	if len(ft.files) != 3 {
		t.Errorf("SetFiles files = %d, want 3", len(ft.files))
	}

	if len(ft.filtered) != 3 {
		t.Errorf("SetFiles filtered = %d, want 3", len(ft.filtered))
	}
}

func TestFileTree_SetFiles_ResetsSelection(t *testing.T) {
	files := []project.File{
		{Name: "Code", Type: project.FileTypeServer},
		{Name: "Utils", Type: project.FileTypeServer},
	}

	ft := NewFileTree(files)
	ft.selected = 1 // Select second item

	// Set new files with only one item
	newFiles := []project.File{
		{Name: "Single", Type: project.FileTypeServer},
	}
	ft.SetFiles(newFiles)

	// Selection should be clamped to valid range
	if ft.selected != 0 {
		t.Errorf("SetFiles should clamp selection, got %d, want 0", ft.selected)
	}
}

func TestFileTree_Navigation(t *testing.T) {
	files := []project.File{
		{Name: "File1", Type: project.FileTypeServer},
		{Name: "File2", Type: project.FileTypeServer},
		{Name: "File3", Type: project.FileTypeServer},
	}

	ft := NewFileTree(files)
	ft.height = 100 // Ensure all visible

	// Initial selection
	if ft.selected != 0 {
		t.Errorf("Initial selection = %d, want 0", ft.selected)
	}

	// Move down with j
	ft.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if ft.selected != 1 {
		t.Errorf("After j: selection = %d, want 1", ft.selected)
	}

	// Move down with down arrow
	ft.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ft.selected != 2 {
		t.Errorf("After down: selection = %d, want 2", ft.selected)
	}

	// Try to go past end (should stay at last)
	ft.Update(tea.KeyMsg{Type: tea.KeyDown})
	if ft.selected != 2 {
		t.Errorf("After down at end: selection = %d, want 2", ft.selected)
	}

	// Move up with k
	ft.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if ft.selected != 1 {
		t.Errorf("After k: selection = %d, want 1", ft.selected)
	}

	// Jump to end with G
	ft.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if ft.selected != 2 {
		t.Errorf("After G: selection = %d, want 2", ft.selected)
	}

	// Jump to start with g
	ft.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if ft.selected != 0 {
		t.Errorf("After g: selection = %d, want 0", ft.selected)
	}
}

func TestFileTree_Filter(t *testing.T) {
	files := []project.File{
		{
			Name: "Code",
			Type: project.FileTypeServer,
			FunctionSet: &project.FunctionSet{
				Functions: []project.Function{{Name: "processData"}},
			},
		},
		{
			Name: "Utils",
			Type: project.FileTypeServer,
			FunctionSet: &project.FunctionSet{
				Functions: []project.Function{{Name: "formatDate"}},
			},
		},
		{Name: "index", Type: project.FileTypeHTML},
	}

	ft := NewFileTree(files)

	// Filter by file name
	ft.SetFilter("code")
	if len(ft.filtered) != 1 {
		t.Errorf("Filter 'code': filtered = %d, want 1", len(ft.filtered))
	}
	if ft.filtered[0].Name != "Code" {
		t.Errorf("Filter 'code': first file = %q, want 'Code'", ft.filtered[0].Name)
	}

	// Filter by function name
	ft.SetFilter("format")
	if len(ft.filtered) != 1 {
		t.Errorf("Filter 'format': filtered = %d, want 1", len(ft.filtered))
	}
	if ft.filtered[0].Name != "Utils" {
		t.Errorf("Filter 'format': should find Utils containing formatDate")
	}

	// Clear filter
	ft.SetFilter("")
	if len(ft.filtered) != 3 {
		t.Errorf("Clear filter: filtered = %d, want 3", len(ft.filtered))
	}

	// Filter with no matches
	ft.SetFilter("xyz123")
	if len(ft.filtered) != 0 {
		t.Errorf("Filter 'xyz123': filtered = %d, want 0", len(ft.filtered))
	}
}

func TestFileTree_FilterResetsSelection(t *testing.T) {
	files := []project.File{
		{Name: "A", Type: project.FileTypeServer},
		{Name: "B", Type: project.FileTypeServer},
		{Name: "C", Type: project.FileTypeServer},
	}

	ft := NewFileTree(files)
	ft.selected = 2 // Select last item

	ft.SetFilter("A")

	// Selection should reset to 0
	if ft.selected != 0 {
		t.Errorf("SetFilter should reset selection to 0, got %d", ft.selected)
	}
	if ft.offset != 0 {
		t.Errorf("SetFilter should reset offset to 0, got %d", ft.offset)
	}
}

func TestFileTree_SelectedFile(t *testing.T) {
	files := []project.File{
		{Name: "Code", Type: project.FileTypeServer},
		{Name: "Utils", Type: project.FileTypeServer},
	}

	ft := NewFileTree(files)

	selected := ft.SelectedFile()
	if selected == nil {
		t.Fatal("SelectedFile returned nil")
	}
	if selected.Name != "Code" {
		t.Errorf("SelectedFile = %q, want 'Code'", selected.Name)
	}

	ft.selected = 1
	selected = ft.SelectedFile()
	if selected == nil {
		t.Fatal("SelectedFile returned nil")
	}
	if selected.Name != "Utils" {
		t.Errorf("SelectedFile = %q, want 'Utils'", selected.Name)
	}
}

func TestFileTree_SelectedFile_Empty(t *testing.T) {
	ft := NewFileTree(nil)

	selected := ft.SelectedFile()
	if selected != nil {
		t.Errorf("SelectedFile on empty tree = %v, want nil", selected)
	}
}

func TestFileTree_FileIcon(t *testing.T) {
	ft := NewFileTree(nil)

	tests := []struct {
		fileType project.FileType
		wantIcon string
	}{
		{project.FileTypeServer, "📄"},
		{project.FileTypeHTML, "🌐"},
		{project.FileTypeJSON, "📋"},
		{project.FileType("UNKNOWN"), "📄"}, // Default
	}

	for _, tt := range tests {
		icon := ft.fileIcon(tt.fileType)
		if icon != tt.wantIcon {
			t.Errorf("fileIcon(%v) = %q, want %q", tt.fileType, icon, tt.wantIcon)
		}
	}
}

func TestFileTree_EnsureVisible(t *testing.T) {
	files := make([]project.File, 20)
	for i := range 20 {
		files[i] = project.File{Name: "File" + string(rune('A'+i)), Type: project.FileTypeServer}
	}

	ft := NewFileTree(files)
	ft.height = 10 // Can show ~6 items (10 - 4 for header/padding)

	// Select item near bottom
	ft.selected = 15
	ft.ensureVisible()

	// Offset should have adjusted to show selected item
	visibleHeight := max(1, ft.height-4)
	if ft.selected < ft.offset || ft.selected >= ft.offset+visibleHeight {
		t.Errorf("ensureVisible failed: selected=%d not in range [%d, %d)",
			ft.selected, ft.offset, ft.offset+visibleHeight)
	}
}

func TestFileTree_EmptyView(t *testing.T) {
	ft := NewFileTree(nil)

	view := ft.View()

	if view == "" {
		t.Error("View() on empty tree should return message, not empty string")
	}

	if !containsText(view, "No files") {
		t.Errorf("View() on empty tree should indicate no files, got: %q", view)
	}
}

func containsText(s, substr string) bool {
	return len(s) > 0 && len(substr) > 0 && contains(s, substr)
}

func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
