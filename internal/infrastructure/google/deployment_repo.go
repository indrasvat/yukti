package google

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"yukti/internal/domain/deployment"
)

// DeploymentRepository implements deployment.Repository using the Apps Script API.
type DeploymentRepository struct {
	client *Client
}

// NewDeploymentRepository creates a new deployment repository.
func NewDeploymentRepository(client *Client) *DeploymentRepository {
	return &DeploymentRepository{client: client}
}

type (
	deploymentResponse struct {
		ID          string                     `json:"deploymentId"`
		Config      deploymentConfigResponse   `json:"deploymentConfig"`
		UpdateTime  string                     `json:"updateTime"`
		EntryPoints []deploymentEntryPointResp `json:"entryPoints"`
	}

	deploymentsResponse struct {
		Deployments   []deploymentResponse `json:"deployments"`
		NextPageToken string               `json:"nextPageToken"`
	}

	deploymentConfigResponse struct {
		ScriptID         string `json:"scriptId"`
		VersionNumber    int    `json:"versionNumber"`
		Description      string `json:"description"`
		ManifestFileName string `json:"manifestFileName"`
	}

	deploymentEntryPointResp struct {
		Type         string                      `json:"entryPointType"`
		WebApp       *webAppEntryPointResp       `json:"webApp,omitempty"`
		ExecutionAPI *executionAPIEntryPointResp `json:"executionApi,omitempty"`
		AddOn        *addOnEntryPointResp        `json:"addOn,omitempty"`
	}

	webAppEntryPointResp struct {
		URL    string `json:"url"`
		Config struct {
			Access string `json:"access"`
		} `json:"entryPointConfig"`
	}

	executionAPIEntryPointResp struct {
		Config struct {
			Access string `json:"access"`
		} `json:"entryPointConfig"`
	}

	addOnEntryPointResp struct {
		AddOnType  string `json:"addOnType"`
		ReportName string `json:"reportName"`
	}
)

// List returns all deployments for a project.
func (r *DeploymentRepository) List(ctx context.Context, scriptID string) ([]deployment.Deployment, error) {
	var all []deployment.Deployment
	pageToken := ""

	for {
		path := fmt.Sprintf("/projects/%s/deployments", url.PathEscape(scriptID))
		params := url.Values{}
		params.Set("pageSize", "50")
		if pageToken != "" {
			params.Set("pageToken", pageToken)
		}
		path += "?" + params.Encode()

		var resp deploymentsResponse
		if err := r.client.Get(ctx, path, &resp); err != nil {
			return nil, fmt.Errorf("listing deployments: %w", err)
		}
		for _, item := range resp.Deployments {
			all = append(all, toDeployment(item))
		}
		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return all, nil
}

// Get returns a specific deployment.
func (r *DeploymentRepository) Get(ctx context.Context, scriptID, deploymentID string) (*deployment.Deployment, error) {
	path := fmt.Sprintf("/projects/%s/deployments/%s", url.PathEscape(scriptID), url.PathEscape(deploymentID))

	var resp deploymentResponse
	if err := r.client.Get(ctx, path, &resp); err != nil {
		return nil, fmt.Errorf("getting deployment: %w", err)
	}

	dep := toDeployment(resp)
	return &dep, nil
}

// Create creates a deployment for a project version.
func (r *DeploymentRepository) Create(ctx context.Context, scriptID string, req deployment.CreateRequest) (*deployment.Deployment, error) {
	path := fmt.Sprintf("/projects/%s/deployments", url.PathEscape(scriptID))

	var resp deploymentResponse
	if err := r.client.Post(ctx, path, deploymentConfigBody(scriptID, req.VersionNumber, req.Description), &resp); err != nil {
		return nil, fmt.Errorf("creating deployment: %w", err)
	}

	dep := toDeployment(resp)
	return &dep, nil
}

// Update points an existing deployment at a new version or description.
func (r *DeploymentRepository) Update(ctx context.Context, scriptID, deploymentID string, req deployment.UpdateRequest) (*deployment.Deployment, error) {
	path := fmt.Sprintf("/projects/%s/deployments/%s", url.PathEscape(scriptID), url.PathEscape(deploymentID))
	body := map[string]any{
		"deploymentConfig": deploymentConfigBody(scriptID, req.VersionNumber, req.Description),
	}

	var resp deploymentResponse
	if err := r.client.Put(ctx, path, body, &resp); err != nil {
		return nil, fmt.Errorf("updating deployment: %w", err)
	}

	dep := toDeployment(resp)
	return &dep, nil
}

// Delete undeploys a deployment.
func (r *DeploymentRepository) Delete(ctx context.Context, scriptID, deploymentID string) error {
	path := fmt.Sprintf("/projects/%s/deployments/%s", url.PathEscape(scriptID), url.PathEscape(deploymentID))
	if err := r.client.Delete(ctx, path); err != nil {
		return fmt.Errorf("deleting deployment: %w", err)
	}
	return nil
}

func deploymentConfigBody(scriptID string, versionNumber int, description string) map[string]any {
	return map[string]any{
		"scriptId":         scriptID,
		"versionNumber":    versionNumber,
		"manifestFileName": "appsscript",
		"description":      description,
	}
}

func toDeployment(resp deploymentResponse) deployment.Deployment {
	updateTime, _ := time.Parse(time.RFC3339, resp.UpdateTime)
	depVersion := &deployment.Version{
		VersionNumber: resp.Config.VersionNumber,
		Description:   resp.Config.Description,
	}

	return deployment.Deployment{
		ID:      resp.ID,
		Version: depVersion,
		Config: deployment.Config{
			ScriptID:    resp.Config.ScriptID,
			VersionID:   resp.Config.VersionNumber,
			Description: resp.Config.Description,
		},
		UpdateTime:  updateTime,
		EntryPoints: toEntryPoints(resp.EntryPoints),
	}
}

func toEntryPoints(resp []deploymentEntryPointResp) []deployment.EntryPoint {
	entryPoints := make([]deployment.EntryPoint, 0, len(resp))
	for _, item := range resp {
		entryPoint := deployment.EntryPoint{
			Type: deployment.EntryPointType(item.Type),
		}
		if item.WebApp != nil {
			entryPoint.WebApp = &deployment.WebAppConfig{
				URL:    item.WebApp.URL,
				Access: deployment.WebAppAccess(item.WebApp.Config.Access),
			}
		}
		if item.ExecutionAPI != nil {
			entryPoint.ExecutionAPI = &deployment.ExecutionAPIConfig{
				Access: deployment.ExecutionAPIAccess(item.ExecutionAPI.Config.Access),
			}
		}
		if item.AddOn != nil {
			entryPoint.AddOn = &deployment.AddOnConfig{
				AddOnType:  item.AddOn.AddOnType,
				ReportName: item.AddOn.ReportName,
			}
		}
		entryPoints = append(entryPoints, entryPoint)
	}
	return entryPoints
}
