package components

import (
	"fmt"
	"sort"
	"strings"
	"unicode"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"yukti/internal/domain/project"
	"yukti/internal/tui/styles"
)

// FuzzyItem represents an item that can be fuzzy matched.
type FuzzyItem struct {
	// Display
	Title       string
	Description string
	Icon        string

	// Data
	File     *project.File
	Function *project.Function
	LineNum  int

	// Matching
	score int
}

// FuzzyFinder provides fuzzy search across files and functions.
type FuzzyFinder struct {
	// State
	input    textinput.Model
	items    []FuzzyItem
	filtered []FuzzyItem
	selected int
	visible  bool

	// Layout
	width  int
	height int

	// Styling
	containerStyle lipgloss.Style
	inputStyle     lipgloss.Style
	itemStyle      lipgloss.Style
	selectedStyle  lipgloss.Style
	matchStyle     lipgloss.Style
	hintStyle      lipgloss.Style
}

// NewFuzzyFinder creates a new fuzzy finder.
func NewFuzzyFinder() *FuzzyFinder {
	ti := textinput.New()
	ti.Placeholder = "Search files and functions..."
	ti.Prompt = "> "
	ti.PromptStyle = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.TextStyle = lipgloss.NewStyle().Foreground(styles.TextPrimary)
	ti.Cursor.Style = lipgloss.NewStyle().Foreground(styles.Primary)
	ti.Focus()

	return &FuzzyFinder{
		input:    ti,
		items:    []FuzzyItem{},
		filtered: []FuzzyItem{},
		selected: 0,
		visible:  false,
		width:    60,
		height:   20,
		containerStyle: lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(styles.Primary).
			Background(styles.Background).
			Padding(1),
		inputStyle: lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(styles.Border).
			Padding(0, 1).
			MarginBottom(1),
		itemStyle: lipgloss.NewStyle().
			Padding(0, 1),
		selectedStyle: lipgloss.NewStyle().
			Background(styles.Surface).
			Foreground(styles.Primary).
			Bold(true).
			Padding(0, 1),
		matchStyle: lipgloss.NewStyle().
			Foreground(styles.Warning).
			Bold(true),
		hintStyle: lipgloss.NewStyle().
			Foreground(styles.TextMuted).
			Italic(true),
	}
}

// SetItems populates the fuzzy finder with searchable items.
func (f *FuzzyFinder) SetItems(files []project.File) {
	f.items = make([]FuzzyItem, 0, len(files)*2)

	for i := range files {
		file := &files[i]

		// Add file entry
		icon := "📄"
		switch file.Type {
		case project.FileTypeHTML:
			icon = "🌐"
		case project.FileTypeJSON:
			icon = "📋"
		}

		lines := strings.Count(file.Source, "\n") + 1
		f.items = append(f.items, FuzzyItem{
			Title:       file.Name,
			Description: fmt.Sprintf("%d lines", lines),
			Icon:        icon,
			File:        file,
		})

		// Add function entries
		if file.FunctionSet != nil {
			for j := range file.FunctionSet.Functions {
				fn := &file.FunctionSet.Functions[j]
				// Estimate line number (would need actual parsing for accuracy)
				lineNum := findFunctionLine(file.Source, fn.Name)
				f.items = append(f.items, FuzzyItem{
					Title:       fn.Name + "()",
					Description: file.Name,
					Icon:        "ƒ",
					File:        file,
					Function:    fn,
					LineNum:     lineNum,
				})
			}
		}
	}

	f.filtered = f.items
}

// findFunctionLine finds the approximate line number for a function.
func findFunctionLine(source, funcName string) int {
	lines := strings.Split(source, "\n")
	pattern := "function " + funcName
	for i, line := range lines {
		if strings.Contains(line, pattern) {
			return i + 1
		}
	}
	return 1
}

// Show makes the fuzzy finder visible and focuses input.
func (f *FuzzyFinder) Show() tea.Cmd {
	f.visible = true
	f.input.SetValue("")
	f.input.Focus()
	f.filtered = f.items
	f.selected = 0
	return textinput.Blink
}

// Hide closes the fuzzy finder.
func (f *FuzzyFinder) Hide() {
	f.visible = false
	f.input.Blur()
}

// IsVisible returns whether the finder is shown.
func (f *FuzzyFinder) IsVisible() bool {
	return f.visible
}

// Init implements tea.Model.
func (f *FuzzyFinder) Init() tea.Cmd {
	return nil
}

// Update implements tea.Model.
func (f *FuzzyFinder) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if !f.visible {
		return f, nil
	}

	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		f.width = min(msg.Width-4, 80)
		f.height = min(msg.Height-4, 25)

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			f.Hide()
			return f, nil

		case "enter":
			if f.selected < len(f.filtered) {
				item := f.filtered[f.selected]
				f.Hide()
				return f, func() tea.Msg {
					return FuzzySelectMsg{Item: item}
				}
			}

		case "down", "ctrl+n":
			if f.selected < len(f.filtered)-1 {
				f.selected++
			}
			return f, nil

		case "up", "ctrl+p":
			if f.selected > 0 {
				f.selected--
			}
			return f, nil

		default:
			// Update text input
			var cmd tea.Cmd
			f.input, cmd = f.input.Update(msg)
			cmds = append(cmds, cmd)

			// Re-filter on input change
			f.filter()
		}
	}

	return f, tea.Batch(cmds...)
}

// filter applies fuzzy matching to items.
func (f *FuzzyFinder) filter() {
	query := strings.ToLower(f.input.Value())
	if query == "" {
		f.filtered = f.items
		f.selected = 0
		return
	}

	// Score and filter items
	scored := make([]FuzzyItem, 0, len(f.items))
	for _, item := range f.items {
		score := fuzzyScore(strings.ToLower(item.Title), query)
		if score > 0 {
			item.score = score
			scored = append(scored, item)
		}
	}

	// Sort by score (descending)
	sort.Slice(scored, func(i, j int) bool {
		return scored[i].score > scored[j].score
	})

	f.filtered = scored
	f.selected = 0
}

// fuzzyScore calculates a fuzzy match score.
// Higher scores mean better matches.
func fuzzyScore(text, pattern string) int {
	if pattern == "" {
		return 1
	}
	if text == "" {
		return 0
	}

	// Exact substring match gets highest score
	// Shorter texts (more precise matches) score higher via ratio
	if strings.Contains(text, pattern) {
		return 100 + (len(pattern) * 100 / len(text))
	}

	// Fuzzy matching: all pattern chars must appear in order
	score := 0
	patternIdx := 0
	consecutiveBonus := 0
	prevMatch := false

	for i := 0; i < len(text) && patternIdx < len(pattern); i++ {
		if text[i] == pattern[patternIdx] {
			score += 10
			patternIdx++

			// Bonus for consecutive matches
			if prevMatch {
				consecutiveBonus += 5
			}
			prevMatch = true

			// Bonus for matching at word boundaries
			if i == 0 || !unicode.IsLetter(rune(text[i-1])) {
				score += 15
			}
		} else {
			prevMatch = false
		}
	}

	// All pattern characters must be found
	if patternIdx < len(pattern) {
		return 0
	}

	return score + consecutiveBonus
}

// View implements tea.Model.
func (f *FuzzyFinder) View() string {
	if !f.visible {
		return ""
	}

	// Input field
	inputView := f.inputStyle.Width(f.width - 4).Render(f.input.View())

	// Results list
	maxResults := max(1, f.height-6)

	var results strings.Builder
	count := min(len(f.filtered), maxResults)

	if len(f.filtered) == 0 {
		results.WriteString(f.hintStyle.Render("  No results found"))
	} else {
		for i := range count {
			item := f.filtered[i]
			isSelected := i == f.selected

			line := fmt.Sprintf("%s %s", item.Icon, item.Title)
			if item.Description != "" {
				line += " " + f.hintStyle.Render(item.Description)
			}

			if isSelected {
				line = "▸ " + line
				results.WriteString(f.selectedStyle.Width(f.width - 4).Render(line))
			} else {
				line = "  " + line
				results.WriteString(f.itemStyle.Width(f.width - 4).Render(line))
			}

			if i < count-1 {
				results.WriteString("\n")
			}
		}
	}

	// Count indicator
	countText := f.hintStyle.Render(fmt.Sprintf("  %d/%d", len(f.filtered), len(f.items)))

	content := lipgloss.JoinVertical(
		lipgloss.Left,
		inputView,
		results.String(),
		countText,
	)

	return f.containerStyle.Width(f.width).Render(content)
}

// FuzzySelectMsg is sent when an item is selected.
type FuzzySelectMsg struct {
	Item FuzzyItem
}

// ShortHelp returns help bindings for the fuzzy finder.
func (f *FuzzyFinder) ShortHelp() []key.Binding {
	return []key.Binding{
		key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "select"),
		),
		key.NewBinding(
			key.WithKeys("esc"),
			key.WithHelp("esc", "close"),
		),
		key.NewBinding(
			key.WithKeys("up", "down"),
			key.WithHelp("↑/↓", "navigate"),
		),
	}
}
