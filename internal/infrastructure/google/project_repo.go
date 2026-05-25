package google

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"yukti/internal/domain/project"
)

// DriveBaseURL is the Google Drive API base URL for listing script projects.
// The Apps Script API doesn't have a list endpoint, so we use Drive API.
const DriveBaseURL = "https://www.googleapis.com/drive/v3"

// AppsScriptMIME is the MIME type for Google Apps Script files.
const AppsScriptMIME = "application/vnd.google-apps.script"

// ProjectRepository implements project.Repository using the Google APIs.
// Uses Drive API for listing (Apps Script API lacks list endpoint)
// and Apps Script API for content/metrics operations.
type ProjectRepository struct {
	client *Client
}

// NewProjectRepository creates a new project repository.
func NewProjectRepository(client *Client) *ProjectRepository {
	return &ProjectRepository{client: client}
}

// API response types for JSON unmarshaling.
type (
	// Drive API response for listing files
	driveFilesResponse struct {
		Files         []driveFileResponse `json:"files"`
		NextPageToken string              `json:"nextPageToken"`
	}

	driveFileResponse struct {
		ID             string   `json:"id"`
		Name           string   `json:"name"`
		CreatedTime    string   `json:"createdTime"`
		ModifiedTime   string   `json:"modifiedTime"`
		Owners         []owner  `json:"owners,omitempty"`
		LastModifyUser owner    `json:"lastModifyingUser,omitempty"`
		Parents        []string `json:"parents,omitempty"`
	}

	owner struct {
		DisplayName  string `json:"displayName,omitempty"`
		EmailAddress string `json:"emailAddress,omitempty"`
		PhotoLink    string `json:"photoLink,omitempty"`
	}

	// Apps Script API response for project details
	projectResponse struct {
		ScriptID       string       `json:"scriptId"`
		Title          string       `json:"title"`
		ParentID       string       `json:"parentId,omitempty"`
		CreateTime     string       `json:"createTime"`
		UpdateTime     string       `json:"updateTime"`
		Creator        userResponse `json:"creator,omitempty"`
		LastModifyUser userResponse `json:"lastModifyUser,omitempty"`
	}

	userResponse struct {
		Domain   string `json:"domain,omitempty"`
		Email    string `json:"email,omitempty"`
		Name     string `json:"name,omitempty"`
		PhotoURL string `json:"photoUrl,omitempty"`
	}

	contentResponse struct {
		ScriptID string         `json:"scriptId"`
		Files    []fileResponse `json:"files"`
	}

	fileResponse struct {
		Name        string              `json:"name"`
		Type        string              `json:"type"`
		Source      string              `json:"source"`
		CreateTime  string              `json:"createTime,omitempty"`
		UpdateTime  string              `json:"updateTime,omitempty"`
		FunctionSet functionSetResponse `json:"functionSet,omitempty"`
	}

	functionSetResponse struct {
		Values []functionResponse `json:"values,omitempty"`
	}

	functionResponse struct {
		Name       string   `json:"name"`
		Parameters []string `json:"parameters,omitempty"`
	}
)

// List retrieves all Apps Script projects using the Drive API.
// The Apps Script API doesn't have a list endpoint, so we query Drive for
// files with mimeType='application/vnd.google-apps.script'.
func (r *ProjectRepository) List(ctx context.Context, opts project.ListOptions) (*project.ListResult, error) {
	// Build Drive API query
	params := url.Values{}
	params.Set("q", fmt.Sprintf("mimeType='%s' and trashed=false", AppsScriptMIME))
	params.Set("fields", "nextPageToken,files(id,name,createdTime,modifiedTime,owners,lastModifyingUser,parents)")
	params.Set("orderBy", "modifiedTime desc")

	if opts.PageSize > 0 {
		params.Set("pageSize", strconv.Itoa(opts.PageSize))
	} else {
		params.Set("pageSize", "50") // Default page size
	}
	if opts.PageToken != "" {
		params.Set("pageToken", opts.PageToken)
	}

	// Use Drive API endpoint
	driveURL := DriveBaseURL + "/files?" + params.Encode()

	var resp driveFilesResponse
	if err := r.client.GetAbsolute(ctx, driveURL, &resp); err != nil {
		return nil, fmt.Errorf("listing projects: %w", err)
	}

	result := &project.ListResult{
		Projects:      make([]project.Project, 0, len(resp.Files)),
		NextPageToken: resp.NextPageToken,
	}

	for i := range resp.Files {
		proj := driveFileToProject(resp.Files[i])
		result.Projects = append(result.Projects, proj)
	}

	return result, nil
}

// driveFileToProject converts a Drive file response to a Project domain object.
func driveFileToProject(f driveFileResponse) project.Project {
	createTime, _ := time.Parse(time.RFC3339, f.CreatedTime)
	modifiedTime, _ := time.Parse(time.RFC3339, f.ModifiedTime)

	var parentID string
	if len(f.Parents) > 0 {
		parentID = f.Parents[0]
	}

	var creator project.User
	if len(f.Owners) > 0 {
		creator = project.User{
			Name:     f.Owners[0].DisplayName,
			Email:    f.Owners[0].EmailAddress,
			PhotoURL: f.Owners[0].PhotoLink,
		}
	}

	lastModifier := project.User{
		Name:     f.LastModifyUser.DisplayName,
		Email:    f.LastModifyUser.EmailAddress,
		PhotoURL: f.LastModifyUser.PhotoLink,
	}

	return project.Project{
		ID:           f.ID,
		Title:        f.Name,
		ParentID:     parentID,
		CreateTime:   createTime,
		UpdateTime:   modifiedTime,
		Creator:      creator,
		LastModifier: lastModifier,
	}
}

// Get retrieves a single project by ID.
func (r *ProjectRepository) Get(ctx context.Context, id string) (*project.Project, error) {
	path := fmt.Sprintf("/projects/%s", url.PathEscape(id))

	var resp projectResponse
	if err := r.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("getting project: %w", err)
	}

	return toProject(resp)
}

// GetContent retrieves the content (files) of a project.
func (r *ProjectRepository) GetContent(ctx context.Context, id string) (*project.Content, error) {
	path := fmt.Sprintf("/projects/%s/content", url.PathEscape(id))

	var resp contentResponse
	if err := r.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("getting project content: %w", err)
	}

	return toContent(resp), nil
}

// GetMetrics retrieves project usage metrics.
func (r *ProjectRepository) GetMetrics(ctx context.Context, id string, opts project.MetricsOptions) (*project.Metrics, error) {
	path := fmt.Sprintf("/projects/%s/metrics", url.PathEscape(id))
	params := url.Values{}

	// Format times as RFC3339
	if !opts.StartTime.IsZero() {
		params.Set("metricsFilter.startTime", opts.StartTime.Format(time.RFC3339))
	}
	if !opts.EndTime.IsZero() {
		params.Set("metricsFilter.endTime", opts.EndTime.Format(time.RFC3339))
	}

	if len(params) > 0 {
		path += "?" + params.Encode()
	}

	// The metrics API returns a different structure - simplified for now
	var resp struct {
		ActiveUsers   []metricValueResponse `json:"activeUsers"`
		TotalUsers    []metricValueResponse `json:"totalUsers"`
		Executions    []metricValueResponse `json:"executions"`
		FailedPercent []metricValueResponse `json:"failedExecutions"`
	}

	if err := r.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("getting project metrics: %w", err)
	}

	return &project.Metrics{
		ActiveUsers:   toMetricValues(resp.ActiveUsers),
		TotalUsers:    toMetricValues(resp.TotalUsers),
		Executions:    toMetricValues(resp.Executions),
		FailedPercent: toMetricValues(resp.FailedPercent),
	}, nil
}

type metricValueResponse struct {
	StartTime string `json:"startTime"`
	Value     string `json:"value"`
}

// Create creates a new Apps Script project.
func (r *ProjectRepository) Create(ctx context.Context, req project.CreateRequest) (*project.Project, error) {
	body := map[string]string{
		"title": req.Title,
	}
	if req.ParentID != "" {
		body["parentId"] = req.ParentID
	}

	var resp projectResponse
	if err := r.client.Post(ctx, "/projects", body, &resp); err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	return toProject(resp)
}

// UpdateContent updates the content (files) of a project.
func (r *ProjectRepository) UpdateContent(ctx context.Context, id string, content *project.Content) error {
	path := fmt.Sprintf("/projects/%s/content", url.PathEscape(id))

	// Convert to API format
	files := make([]map[string]any, 0, len(content.Files))
	for _, f := range content.Files {
		file := map[string]any{
			"name":   f.Name,
			"type":   string(f.Type),
			"source": f.Source,
		}
		files = append(files, file)
	}

	body := map[string]any{
		"files": files,
	}

	if err := r.client.Put(ctx, path, body, nil); err != nil {
		return fmt.Errorf("updating project content: %w", err)
	}

	return nil
}

// Helper functions to convert API responses to domain entities.

func toProject(resp projectResponse) (*project.Project, error) {
	createTime, _ := time.Parse(time.RFC3339, resp.CreateTime)
	updateTime, _ := time.Parse(time.RFC3339, resp.UpdateTime)

	return &project.Project{
		ID:           resp.ScriptID,
		Title:        resp.Title,
		ParentID:     resp.ParentID,
		CreateTime:   createTime,
		UpdateTime:   updateTime,
		Creator:      toUser(resp.Creator),
		LastModifier: toUser(resp.LastModifyUser),
	}, nil
}

func toUser(resp userResponse) project.User {
	return project.User{
		Domain:   resp.Domain,
		Email:    resp.Email,
		Name:     resp.Name,
		PhotoURL: resp.PhotoURL,
	}
}

func toContent(resp contentResponse) *project.Content {
	files := make([]project.File, 0, len(resp.Files))
	for _, f := range resp.Files {
		updateTime, _ := time.Parse(time.RFC3339, f.UpdateTime)

		file := project.File{
			Name:         f.Name,
			Type:         project.FileType(f.Type),
			Source:       f.Source,
			LastModified: updateTime,
		}

		// Convert functions
		if len(f.FunctionSet.Values) > 0 {
			functions := make([]project.Function, 0, len(f.FunctionSet.Values))
			for _, fn := range f.FunctionSet.Values {
				functions = append(functions, project.Function{
					Name:       fn.Name,
					Parameters: fn.Parameters,
				})
			}
			file.FunctionSet = &project.FunctionSet{Functions: functions}
		}

		files = append(files, file)
	}

	return &project.Content{
		ScriptID: resp.ScriptID,
		Files:    files,
	}
}

func toMetricValues(values []metricValueResponse) []project.MetricValue {
	result := make([]project.MetricValue, 0, len(values))
	for _, v := range values {
		t, _ := time.Parse(time.RFC3339, v.StartTime)
		val, _ := strconv.ParseInt(v.Value, 10, 64)
		result = append(result, project.MetricValue{
			Time:  t,
			Value: val,
		})
	}
	return result
}
