package views

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"yukti/internal/domain/deployment"
	"yukti/internal/domain/project"
	"yukti/internal/domain/version"
	"yukti/internal/tui/styles"
)

type deploymentsState int

const (
	deploymentsLoading deploymentsState = iota
	deploymentsReady
	deploymentsDeploying
	deploymentsError
)

// DeploymentsView displays deployments and can create a versioned deployment.
type DeploymentsView struct {
	proj    project.Project
	depRepo deployment.Repository
	verRepo version.Repository

	deployments []deployment.Deployment
	versions    []version.Version
	spinner     spinner.Model
	state       deploymentsState
	errMsg      string
	status      string
	confirmNew  bool
	confirmUpd  bool
	cursor      int
	scroll      int
	width       int
	height      int
}

// NewDeploymentsView creates a new deployments view.
func NewDeploymentsView(proj project.Project, depRepo deployment.Repository, verRepo version.Repository) *DeploymentsView {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(styles.Primary)

	return &DeploymentsView{
		proj:    proj,
		depRepo: depRepo,
		verRepo: verRepo,
		spinner: s,
		state:   deploymentsLoading,
		width:   100,
		height:  34,
	}
}

// Title implements tui.View.
func (v *DeploymentsView) Title() string {
	return v.proj.Title + " Deployments"
}

// HasModal returns true while an inline deployment confirmation is active.
func (v *DeploymentsView) HasModal() bool {
	return v.confirmNew || v.confirmUpd
}

// ShortHelp implements tui.View.
func (v *DeploymentsView) ShortHelp() []key.Binding {
	if v.confirmNew {
		return []key.Binding{
			key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "confirm new URL")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}
	if v.confirmUpd {
		return []key.Binding{
			key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "confirm update")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "cancel")),
		}
	}
	return []key.Binding{
		key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "select")),
		key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "select")),
		key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "update selected")),
		key.NewBinding(key.WithKeys("n"), key.WithHelp("n", "new URL")),
		key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "refresh")),
	}
}

// Init implements tea.Model.
func (v *DeploymentsView) Init() tea.Cmd {
	return tea.Batch(v.spinner.Tick, v.loadDeployments())
}

// Update implements tea.Model.
func (v *DeploymentsView) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		v.width = msg.Width
		v.height = msg.Height
		v.ensureCursorVisible()
		return v, nil

	case tea.KeyMsg:
		return v.handleKeyMsg(msg)

	case spinner.TickMsg:
		if v.state == deploymentsLoading || v.state == deploymentsDeploying {
			var cmd tea.Cmd
			v.spinner, cmd = v.spinner.Update(msg)
			cmds = append(cmds, cmd)
		}

	case deploymentsLoadedMsg:
		v.state = deploymentsReady
		v.deployments = msg.deployments
		v.versions = msg.versions
		v.errMsg = ""
		v.confirmNew = false
		v.confirmUpd = false
		v.normalizeCursor()
		return v, nil

	case deploymentCreatedMsg:
		if msg.updated {
			v.status = fmt.Sprintf("Moved %s to v%d", msg.deploymentID, msg.versionNumber)
		} else {
			v.status = fmt.Sprintf("Created %s at v%d", msg.deploymentID, msg.versionNumber)
		}
		v.state = deploymentsLoading
		return v, tea.Batch(v.spinner.Tick, v.loadDeployments())

	case deploymentsErrorMsg:
		v.state = deploymentsError
		v.errMsg = msg.err.Error()
		return v, nil
	}

	return v, tea.Batch(cmds...)
}

func (v *DeploymentsView) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		return v.refresh()
	case "j", "down":
		v.moveCursor(1)
	case "k", "up":
		v.moveCursor(-1)
	case "n":
		return v.confirmNewDeployment()
	case "u":
		return v.confirmUpdateDeployment()
	case "esc", "escape":
		if v.confirmNew || v.confirmUpd {
			v.confirmNew = false
			v.confirmUpd = false
			v.status = ""
			return v, nil
		}
	}
	return v, nil
}

func (v *DeploymentsView) refresh() (tea.Model, tea.Cmd) {
	v.state = deploymentsLoading
	v.errMsg = ""
	v.status = ""
	v.confirmNew = false
	v.confirmUpd = false
	return v, tea.Batch(v.spinner.Tick, v.loadDeployments())
}

func (v *DeploymentsView) moveCursor(delta int) {
	if v.state != deploymentsReady || len(v.deployments) == 0 {
		return
	}
	v.cursor += delta
	v.confirmNew = false
	v.confirmUpd = false
	v.normalizeCursor()
}

func (v *DeploymentsView) confirmNewDeployment() (tea.Model, tea.Cmd) {
	if v.state != deploymentsReady {
		return v, nil
	}
	if !v.confirmNew {
		v.confirmNew = true
		v.confirmUpd = false
		v.status = "Press n again to create a new deployment URL."
		return v, nil
	}
	v.state = deploymentsDeploying
	v.errMsg = ""
	v.confirmNew = false
	v.status = "Creating immutable version and new deployment..."
	return v, tea.Batch(v.spinner.Tick, v.deployHead(""))
}

func (v *DeploymentsView) confirmUpdateDeployment() (tea.Model, tea.Cmd) {
	if v.state != deploymentsReady {
		return v, nil
	}
	selected, ok := v.selectedDeployment()
	if !ok {
		v.status = "Select a versioned deployment to update."
		return v, nil
	}
	if selected.Config.VersionID == 0 {
		v.status = "The automatic HEAD deployment cannot be updated here."
		return v, nil
	}
	if !v.confirmUpd {
		v.confirmUpd = true
		v.confirmNew = false
		v.status = fmt.Sprintf("Press u again to move %s to a new version.", selected.ID)
		return v, nil
	}
	v.state = deploymentsDeploying
	v.errMsg = ""
	v.confirmUpd = false
	v.status = fmt.Sprintf("Creating immutable version and updating %s...", selected.ID)
	return v, tea.Batch(v.spinner.Tick, v.deployHead(selected.ID))
}

// View implements tea.Model.
func (v *DeploymentsView) View() string {
	switch v.state {
	case deploymentsLoading, deploymentsDeploying:
		return v.renderLoading()
	case deploymentsError:
		return v.renderError()
	default:
		return v.renderReady()
	}
}

func (v *DeploymentsView) renderLoading() string {
	message := "Loading deployments..."
	if v.state == deploymentsDeploying {
		message = v.status
	}
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		v.spinner.View(),
		"",
		lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(message),
	)
	return lipgloss.Place(v.width, v.height, lipgloss.Center, lipgloss.Center, content)
}

func (v *DeploymentsView) renderError() string {
	content := lipgloss.JoinVertical(
		lipgloss.Center,
		lipgloss.NewStyle().Foreground(styles.Error).Bold(true).Render("Failed to load deployments"),
		"",
		lipgloss.NewStyle().Foreground(styles.TextSecondary).Render(v.errMsg),
		"",
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render("Press r to retry or esc to go back"),
	)
	return lipgloss.Place(v.width, v.height, lipgloss.Center, lipgloss.Center, content)
}

func (v *DeploymentsView) renderReady() string {
	lines := make([]string, 0, 8+(v.maxVisibleDeployments()*4))
	title := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true).Render("Deployments")
	meta := lipgloss.NewStyle().Foreground(styles.TextMuted).Render(fmt.Sprintf("Script ID %s", v.proj.ID))
	lines = append(lines, title, meta, "")

	if v.status != "" {
		statusColor := styles.Success
		if v.confirmNew || v.confirmUpd {
			statusColor = styles.Warning
		}
		lines = append(lines, lipgloss.NewStyle().Foreground(statusColor).Render(v.status), "")
	}
	if v.confirmNew || v.confirmUpd {
		lines = append(lines, v.renderConfirmationPreview()...)
		lines = append(lines, "")
	}

	if len(v.deployments) == 0 {
		lines = append(lines,
			lipgloss.NewStyle().Foreground(styles.TextSecondary).Render("No deployments returned by Google."),
			lipgloss.NewStyle().Foreground(styles.TextMuted).Render("Press n twice to create a versioned deployment from current remote HEAD."),
		)
		return padDeploymentsView(lines, v.width, v.height)
	}

	visible := v.visibleDeployments()
	for _, dep := range visible {
		selected := v.deployments[v.cursor].ID == dep.ID
		lines = append(lines, renderDeploymentLine(dep, selected))
		if dep.Config.Description != "" {
			lines = append(lines, "  "+lipgloss.NewStyle().Foreground(styles.TextMuted).Render(dep.Config.Description))
		}
		for _, entry := range dep.EntryPoints {
			lines = append(lines, "  "+renderEntryPoint(entry))
		}
		lines = append(lines, "")
	}
	if v.scroll > 0 || v.scroll+len(visible) < len(v.deployments) {
		lines = append(lines, lipgloss.NewStyle().Foreground(styles.TextMuted).Render(
			fmt.Sprintf("%d of %d deployments visible", len(visible), len(v.deployments)),
		))
	}

	return padDeploymentsView(lines, v.width, v.height)
}

func (v *DeploymentsView) renderConfirmationPreview() []string {
	labelStyle := lipgloss.NewStyle().Foreground(styles.TextMuted)
	valueStyle := lipgloss.NewStyle().Foreground(styles.TextPrimary)
	nextVersion := fmt.Sprintf("v%d", v.nextVersionNumber())

	if v.confirmNew {
		return []string{
			labelStyle.Render("Version") + "  " + valueStyle.Render("remote HEAD -> "+nextVersion),
			labelStyle.Render("Deployment") + "  " + valueStyle.Render("new versioned deployment"),
			labelStyle.Render("URL") + "  " + valueStyle.Render("new web app URL"),
			labelStyle.Render("Result") + "  " + valueStyle.Render("adds another Apps Script deployment"),
		}
	}

	selected, ok := v.selectedDeployment()
	if !ok {
		return []string{labelStyle.Render("No deployment selected")}
	}
	currentVersion := "HEAD"
	if selected.Config.VersionID > 0 {
		currentVersion = fmt.Sprintf("v%d", selected.Config.VersionID)
	}
	url := selected.WebAppURL()
	if url == "" {
		url = "unchanged"
	}
	return []string{
		labelStyle.Render("Deployment") + "  " + valueStyle.Render(selected.ID),
		labelStyle.Render("Version") + "  " + valueStyle.Render(currentVersion+" -> "+nextVersion),
		labelStyle.Render("URL") + "  " + valueStyle.Render("retains "+url),
		labelStyle.Render("Result") + "  " + valueStyle.Render("moves the selected deployment to a new immutable version"),
	}
}

func renderDeploymentLine(dep deployment.Deployment, selected bool) string {
	versionLabel := "HEAD"
	if dep.Config.VersionID > 0 {
		versionLabel = fmt.Sprintf("v%d", dep.Config.VersionID)
	}
	marker := " "
	idStyle := lipgloss.NewStyle().Foreground(styles.Primary).Bold(true)
	if selected {
		marker = lipgloss.NewStyle().Foreground(styles.Accent).Bold(true).Render(">")
		idStyle = idStyle.Foreground(styles.Accent)
	}
	return fmt.Sprintf("%s  %s  %s",
		marker+" "+idStyle.Render(dep.ID),
		lipgloss.NewStyle().Foreground(styles.Success).Render(versionLabel),
		lipgloss.NewStyle().Foreground(styles.TextMuted).Render(deploymentUpdatedLabel(dep.UpdateTime)),
	)
}

func deploymentUpdatedLabel(t time.Time) string {
	if t.IsZero() || t.Year() < 2000 {
		return "automatic"
	}
	return formatTimeAgo(t)
}

func renderEntryPoint(entry deployment.EntryPoint) string {
	parts := []string{lipgloss.NewStyle().Foreground(styles.Info).Render(string(entry.Type))}
	if entry.WebApp != nil && entry.WebApp.URL != "" {
		parts = append(parts, entry.WebApp.URL)
	}
	if entry.ExecutionAPI != nil {
		parts = append(parts, string(entry.ExecutionAPI.Access))
	}
	return strings.Join(parts, "  ")
}

func padDeploymentsView(lines []string, width, height int) string {
	padY := 1
	padX := 2
	if height <= 2 {
		padY = 0
	}
	if width <= 4 {
		padX = 0
	}

	innerHeight, innerWidth := deploymentsInnerSize(width, height)

	if len(lines) > innerHeight {
		lines = lines[:innerHeight]
	}
	for i, line := range lines {
		if ansi.StringWidth(line) > innerWidth {
			lines[i] = ansi.Truncate(line, innerWidth, "…")
		}
	}
	emptyLine := strings.Repeat(" ", innerWidth)
	for len(lines) < innerHeight {
		lines = append(lines, emptyLine)
	}
	return lipgloss.NewStyle().
		Padding(padY, padX).
		MaxWidth(width).
		MaxHeight(height).
		Render(strings.Join(lines, "\n"))
}

func (v *DeploymentsView) loadDeployments() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		deployments, err := v.depRepo.List(ctx, v.proj.ID)
		if err != nil {
			return deploymentsErrorMsg{err: err}
		}
		versions, err := v.verRepo.List(ctx, v.proj.ID)
		if err != nil {
			return deploymentsErrorMsg{err: err}
		}
		return deploymentsLoadedMsg{deployments: deployments, versions: versions}
	}
}

func (v *DeploymentsView) deployHead(deploymentID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		description := fmt.Sprintf("Yukti TUI deploy %s", time.Now().UTC().Format("2006-01-02 15:04 MST"))
		ver, err := v.verRepo.Create(ctx, v.proj.ID, version.CreateRequest{Description: description})
		if err != nil {
			return deploymentsErrorMsg{err: err}
		}
		var dep *deployment.Deployment
		if deploymentID == "" {
			dep, err = v.depRepo.Create(ctx, v.proj.ID, deployment.CreateRequest{
				VersionNumber: ver.VersionNumber,
				Description:   description,
			})
		} else {
			dep, err = v.depRepo.Update(ctx, v.proj.ID, deploymentID, deployment.UpdateRequest{
				VersionNumber: ver.VersionNumber,
				Description:   description,
			})
		}
		if err != nil {
			return deploymentsErrorMsg{err: err}
		}
		return deploymentCreatedMsg{versionNumber: ver.VersionNumber, deploymentID: dep.ID, updated: deploymentID != ""}
	}
}

func (v *DeploymentsView) selectedDeployment() (deployment.Deployment, bool) {
	if len(v.deployments) == 0 || v.cursor < 0 || v.cursor >= len(v.deployments) {
		return deployment.Deployment{}, false
	}
	return v.deployments[v.cursor], true
}

func (v *DeploymentsView) nextVersionNumber() int {
	maxVersion := 0
	for _, ver := range v.versions {
		if ver.VersionNumber > maxVersion {
			maxVersion = ver.VersionNumber
		}
	}
	for _, dep := range v.deployments {
		if dep.Config.VersionID > maxVersion {
			maxVersion = dep.Config.VersionID
		}
		if dep.Version != nil && dep.Version.VersionNumber > maxVersion {
			maxVersion = dep.Version.VersionNumber
		}
	}
	return maxVersion + 1
}

func (v *DeploymentsView) normalizeCursor() {
	if len(v.deployments) == 0 {
		v.cursor = 0
		v.scroll = 0
		return
	}
	if v.cursor >= len(v.deployments) {
		v.cursor = len(v.deployments) - 1
	}
	if v.cursor < 0 {
		v.cursor = 0
	}
	v.ensureCursorVisible()
}

func (v *DeploymentsView) ensureCursorVisible() {
	maxVisible := v.maxVisibleDeployments()
	if v.cursor < v.scroll {
		v.scroll = v.cursor
	}
	if v.cursor >= v.scroll+maxVisible {
		v.scroll = v.cursor - maxVisible + 1
	}
	if v.scroll < 0 {
		v.scroll = 0
	}
}

func (v *DeploymentsView) visibleDeployments() []deployment.Deployment {
	if len(v.deployments) == 0 {
		return nil
	}
	v.normalizeCursor()
	maxVisible := v.maxVisibleDeployments()
	end := v.scroll + maxVisible
	if end > len(v.deployments) {
		end = len(v.deployments)
	}
	return v.deployments[v.scroll:end]
}

func (v *DeploymentsView) maxVisibleDeployments() int {
	innerHeight, _ := deploymentsInnerSize(v.width, v.height)
	maxVisible := (innerHeight - 5) / 4
	if maxVisible < 1 {
		return 1
	}
	return maxVisible
}

func deploymentsInnerSize(width, height int) (innerHeight, innerWidth int) {
	padY := 1
	padX := 2
	if height <= 2 {
		padY = 0
	}
	if width <= 4 {
		padX = 0
	}
	innerHeight = height - (padY * 2)
	innerWidth = width - (padX * 2)
	if innerHeight < 0 {
		innerHeight = 0
	}
	if innerWidth < 0 {
		innerWidth = 0
	}
	return innerHeight, innerWidth
}

type deploymentsLoadedMsg struct {
	deployments []deployment.Deployment
	versions    []version.Version
}

type deploymentCreatedMsg struct {
	versionNumber int
	deploymentID  string
	updated       bool
}

type deploymentsErrorMsg struct {
	err error
}
