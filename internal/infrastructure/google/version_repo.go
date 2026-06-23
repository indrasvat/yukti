package google

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"yukti/internal/domain/version"
)

// VersionRepository implements version.Repository using the Apps Script API.
type VersionRepository struct {
	client *Client
}

// NewVersionRepository creates a new version repository.
func NewVersionRepository(client *Client) *VersionRepository {
	return &VersionRepository{client: client}
}

type (
	versionResponse struct {
		ScriptID      string `json:"scriptId"`
		VersionNumber int    `json:"versionNumber"`
		Description   string `json:"description"`
		CreateTime    string `json:"createTime"`
	}

	versionsResponse struct {
		Versions      []versionResponse `json:"versions"`
		NextPageToken string            `json:"nextPageToken"`
	}
)

// List returns all versions for a project.
func (r *VersionRepository) List(ctx context.Context, scriptID string) ([]version.Version, error) {
	var all []version.Version
	pageToken := ""

	for {
		path := fmt.Sprintf("/projects/%s/versions", url.PathEscape(scriptID))
		params := url.Values{}
		params.Set("pageSize", "50")
		if pageToken != "" {
			params.Set("pageToken", pageToken)
		}
		path += "?" + params.Encode()

		var resp versionsResponse
		if err := r.client.Get(ctx, path, &resp); err != nil {
			return nil, fmt.Errorf("listing versions: %w", err)
		}
		for _, item := range resp.Versions {
			all = append(all, toVersion(item))
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return all, nil
}

// Get returns a specific version.
func (r *VersionRepository) Get(ctx context.Context, scriptID string, versionNumber int) (*version.Version, error) {
	path := fmt.Sprintf("/projects/%s/versions/%s", url.PathEscape(scriptID), url.PathEscape(strconv.Itoa(versionNumber)))

	var resp versionResponse
	if err := r.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("getting version: %w", err)
	}

	ver := toVersion(resp)
	return &ver, nil
}

// Create creates a new immutable version from the current project HEAD.
func (r *VersionRepository) Create(ctx context.Context, scriptID string, req version.CreateRequest) (*version.Version, error) {
	path := fmt.Sprintf("/projects/%s/versions", url.PathEscape(scriptID))
	body := map[string]string{
		"description": req.Description,
	}

	var resp versionResponse
	if err := r.client.Post(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("creating version: %w", err)
	}

	ver := toVersion(resp)
	return &ver, nil
}

func toVersion(resp versionResponse) version.Version {
	createTime, _ := time.Parse(time.RFC3339, resp.CreateTime)
	return version.Version{
		ScriptID:      resp.ScriptID,
		VersionNumber: resp.VersionNumber,
		Description:   resp.Description,
		CreateTime:    createTime,
	}
}
