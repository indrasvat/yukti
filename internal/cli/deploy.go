package cli

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"yukti/internal/domain/deployment"
	"yukti/internal/domain/version"
	googleinfra "yukti/internal/infrastructure/google"
	"yukti/internal/workspace"
)

var (
	deployScriptID      string
	deployDescription   string
	deployVersionNumber int
	deployYes           bool
	releaseDeploymentID string
	releaseNoPush       bool
)

var versionsCmd = &cobra.Command{
	Use:   "versions",
	Short: "Manage Apps Script immutable versions",
}

var versionsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List versions for a script",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		scriptID, err := resolveScriptID(deployScriptID, syncDir)
		if err != nil {
			return err
		}
		repo, err := authenticatedVersionRepo(ctx)
		if err != nil {
			return err
		}
		versions, err := repo.List(ctx, scriptID)
		if err != nil {
			return err
		}
		printVersions(scriptID, versions)
		return nil
	},
}

var versionsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create an immutable version from the current remote HEAD",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		scriptID, err := resolveScriptID(deployScriptID, syncDir)
		if err != nil {
			return err
		}
		repo, err := authenticatedVersionRepo(ctx)
		if err != nil {
			return err
		}
		ver, err := repo.Create(ctx, scriptID, version.CreateRequest{Description: deployDescription})
		if err != nil {
			return err
		}
		printVersion("Created version", ver)
		return nil
	},
}

var deploymentsCmd = &cobra.Command{
	Use:   "deployments",
	Short: "Manage Apps Script deployments",
}

var deploymentsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List deployments for a script",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		scriptID, err := resolveScriptID(deployScriptID, syncDir)
		if err != nil {
			return err
		}
		repo, err := authenticatedDeploymentRepo(ctx)
		if err != nil {
			return err
		}
		deployments, err := repo.List(ctx, scriptID)
		if err != nil {
			return err
		}
		printDeployments(scriptID, deployments)
		return nil
	},
}

var deploymentsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a deployment for a script version",
	RunE: func(cmd *cobra.Command, args []string) error {
		if deployVersionNumber <= 0 {
			return errors.New("--version must be a positive version number")
		}
		ctx := context.Background()
		scriptID, err := resolveScriptID(deployScriptID, syncDir)
		if err != nil {
			return err
		}
		repo, err := authenticatedDeploymentRepo(ctx)
		if err != nil {
			return err
		}
		dep, err := repo.Create(ctx, scriptID, deployment.CreateRequest{
			VersionNumber: deployVersionNumber,
			Description:   deployDescription,
		})
		if err != nil {
			return err
		}
		printDeployment("Created deployment", dep)
		return nil
	},
}

var deploymentsUpdateCmd = &cobra.Command{
	Use:   "update <deployment-id>",
	Short: "Update an existing deployment to a new version",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if deployVersionNumber <= 0 {
			return errors.New("--version must be a positive version number")
		}
		ctx := context.Background()
		scriptID, err := resolveScriptID(deployScriptID, syncDir)
		if err != nil {
			return err
		}
		repo, err := authenticatedDeploymentRepo(ctx)
		if err != nil {
			return err
		}
		dep, err := repo.Update(ctx, scriptID, args[0], deployment.UpdateRequest{
			VersionNumber: deployVersionNumber,
			Description:   deployDescription,
		})
		if err != nil {
			return err
		}
		printDeployment("Updated deployment", dep)
		return nil
	},
}

var deploymentsDeleteCmd = &cobra.Command{
	Use:   "delete <deployment-id>",
	Short: "Delete a deployment",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if !deployYes {
			return errors.New("refusing to delete without --yes")
		}
		ctx := context.Background()
		scriptID, err := resolveScriptID(deployScriptID, syncDir)
		if err != nil {
			return err
		}
		repo, err := authenticatedDeploymentRepo(ctx)
		if err != nil {
			return err
		}
		if err := repo.Delete(ctx, scriptID, args[0]); err != nil {
			return err
		}
		fmt.Printf("Deleted deployment %s for script %s\n", args[0], scriptID)
		return nil
	},
}

var releaseCmd = &cobra.Command{
	Use:   "release",
	Short: "Push the current workspace, create a version, and deploy it",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		projectRepo, err := authenticatedProjectRepo(ctx)
		if err != nil {
			return err
		}
		versionRepo, err := authenticatedVersionRepo(ctx)
		if err != nil {
			return err
		}
		deploymentRepo, err := authenticatedDeploymentRepo(ctx)
		if err != nil {
			return err
		}

		root, manifest, err := resolveReleaseWorkspace(deployScriptID, syncDir, releaseNoPush)
		if err != nil {
			return err
		}
		scriptID := manifest.ScriptID
		title := manifest.Title
		targetDeploymentID, err := resolveReleaseDeploymentID(ctx, deploymentRepo, scriptID, root, releaseDeploymentID, manifest.DeploymentID)
		if err != nil {
			return err
		}

		if !releaseNoPush {
			result, pushErr := workspace.NewService(projectRepo).Push(ctx, workspace.PushOptions{
				Dir:   syncDir,
				Force: syncForce,
			})
			if pushErr != nil {
				return fmt.Errorf("pushing workspace before release: %w", pushErr)
			}
			title = result.Title
			scriptID = result.ScriptID
			printChanges("Pushed", result.Changes)
		}

		description := deployDescription
		if description == "" {
			description = fmt.Sprintf("Yukti release %s", time.Now().UTC().Format("2006-01-02 15:04 MST"))
		}

		ver, err := versionRepo.Create(ctx, scriptID, version.CreateRequest{Description: description})
		if err != nil {
			return err
		}

		var dep *deployment.Deployment
		if targetDeploymentID != "" {
			dep, err = deploymentRepo.Update(ctx, scriptID, targetDeploymentID, deployment.UpdateRequest{
				VersionNumber: ver.VersionNumber,
				Description:   description,
			})
		} else {
			dep, err = deploymentRepo.Create(ctx, scriptID, deployment.CreateRequest{
				VersionNumber: ver.VersionNumber,
				Description:   description,
			})
		}
		if err != nil {
			return err
		}

		if saveErr := saveWorkspaceDeploymentID(root, dep.ID); saveErr != nil {
			return fmt.Errorf("saving deployment ID to workspace manifest: %w", saveErr)
		}

		printReleaseResult(title, ver, dep)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(versionsCmd, deploymentsCmd, releaseCmd)
	versionsCmd.AddCommand(versionsListCmd, versionsCreateCmd)
	deploymentsCmd.AddCommand(deploymentsListCmd, deploymentsCreateCmd, deploymentsUpdateCmd, deploymentsDeleteCmd)

	for _, cmd := range []*cobra.Command{versionsListCmd, versionsCreateCmd, deploymentsListCmd, deploymentsCreateCmd, deploymentsUpdateCmd, deploymentsDeleteCmd, releaseCmd} {
		cmd.Flags().StringVar(&deployScriptID, "script-id", "", "Apps Script project ID (defaults to current Yukti workspace)")
		cmd.Flags().StringVarP(&syncDir, "dir", "d", "", "Workspace directory (defaults to current directory)")
	}
	for _, cmd := range []*cobra.Command{versionsCreateCmd, deploymentsCreateCmd, deploymentsUpdateCmd, releaseCmd} {
		cmd.Flags().StringVar(&deployDescription, "description", "", "Version or deployment description")
	}
	for _, cmd := range []*cobra.Command{deploymentsCreateCmd, deploymentsUpdateCmd} {
		cmd.Flags().IntVar(&deployVersionNumber, "version", 0, "Version number to deploy")
	}
	deploymentsDeleteCmd.Flags().BoolVar(&deployYes, "yes", false, "Confirm deployment deletion")
	releaseCmd.Flags().BoolVarP(&syncForce, "force", "f", false, "Overwrite push safety checks before release")
	releaseCmd.Flags().StringVar(&releaseDeploymentID, "deployment-id", "", "Existing deployment ID to update instead of creating a new deployment")
	releaseCmd.Flags().BoolVar(&releaseNoPush, "no-push", false, "Create and deploy a version from the current remote HEAD without pushing local files")
}

func resolveScriptID(explicit, dir string) (string, error) {
	scriptID, _, err := resolveWorkspaceInfo(explicit, dir)
	return scriptID, err
}

func resolveWorkspaceInfo(explicit, dir string) (scriptID, title string, err error) {
	if explicit != "" {
		return explicit, explicit, nil
	}
	if dir == "" {
		dir = "."
	}
	root, err := workspace.FindRoot(dir)
	if err != nil {
		return "", "", fmt.Errorf("not in a Yukti workspace; pass --script-id or run from a directory with %s", workspace.ManifestName)
	}
	manifest, err := workspace.LoadManifest(root)
	if err != nil {
		return "", "", err
	}
	return manifest.ScriptID, manifest.Title, nil
}

func resolveReleaseWorkspace(explicit, dir string, noPush bool) (string, *workspace.Manifest, error) {
	if dir == "" {
		dir = "."
	}

	root, err := workspace.FindRoot(dir)
	if err != nil {
		if explicit != "" && noPush {
			return "", &workspace.Manifest{ScriptID: explicit, Title: explicit}, nil
		}
		return "", nil, fmt.Errorf("not in a Yukti workspace; pass --script-id with --no-push or run from a directory with %s", workspace.ManifestName)
	}

	manifest, err := workspace.LoadManifest(root)
	if err != nil {
		return "", nil, err
	}
	if explicit != "" && noPush && explicit != manifest.ScriptID {
		return "", &workspace.Manifest{ScriptID: explicit, Title: explicit}, nil
	}
	if explicit != "" && explicit != manifest.ScriptID {
		return "", nil, fmt.Errorf("--script-id %s does not match workspace script_id %s; use --no-push to release remote HEAD directly", explicit, manifest.ScriptID)
	}
	return root, manifest, nil
}

func saveWorkspaceDeploymentID(root, deploymentID string) error {
	if root == "" {
		return nil
	}
	manifest, err := workspace.LoadManifest(root)
	if err != nil {
		return err
	}
	if manifest.DeploymentID == deploymentID {
		return nil
	}
	manifest.DeploymentID = deploymentID
	return manifest.Save(root)
}

func resolveReleaseDeploymentID(ctx context.Context, repo deployment.Repository, scriptID, root, explicitID, savedID string) (string, error) {
	targetID := explicitID
	fromManifest := false
	if targetID == "" {
		targetID = savedID
		fromManifest = targetID != ""
	}
	if targetID == "" {
		return "", nil
	}

	existingDeployment, err := repo.Get(ctx, scriptID, targetID)
	if err != nil {
		if fromManifest && errors.Is(err, googleinfra.ErrNotFound) {
			fmt.Printf("  %sCached deployment %s no longer exists; creating a new deployment.%s\n", colorDim, targetID, colorReset)
			if clearErr := saveWorkspaceDeploymentID(root, ""); clearErr != nil {
				return "", fmt.Errorf("clearing stale deployment ID from workspace manifest: %w", clearErr)
			}
			return "", nil
		}
		return "", fmt.Errorf("validating deployment before release: %w", err)
	}
	if existingDeployment.Config.VersionID == 0 {
		return "", fmt.Errorf("deployment %s is the automatic HEAD deployment; choose a versioned deployment ID or omit --deployment-id", targetID)
	}
	return targetID, nil
}

func printVersions(scriptID string, versions []version.Version) {
	fmt.Printf("\n  %s%sVersions%s for %s\n\n", colorBold, colorBlue, colorReset, scriptID)
	if len(versions) == 0 {
		fmt.Println("  No versions found")
		return
	}
	for _, ver := range versions {
		printVersion("Version", &ver)
	}
}

func printVersion(prefix string, ver *version.Version) {
	description := ver.Description
	if description == "" {
		description = colorDim + "(no description)" + colorReset
	}
	fmt.Printf("  %s%s%s  v%d  %s\n", colorBold, prefix, colorReset, ver.VersionNumber, description)
	if !ver.CreateTime.IsZero() {
		fmt.Printf("     %screated %s%s\n", colorDim, ver.CreateTime.Format(time.RFC3339), colorReset)
	}
}

func printDeployments(scriptID string, deployments []deployment.Deployment) {
	fmt.Printf("\n  %s%sDeployments%s for %s\n\n", colorBold, colorBlue, colorReset, scriptID)
	if len(deployments) == 0 {
		fmt.Println("  No deployments found")
		return
	}
	for _, dep := range deployments {
		printDeployment("Deployment", &dep)
	}
}

func printDeployment(prefix string, dep *deployment.Deployment) {
	versionLabel := "HEAD"
	if dep.Config.VersionID > 0 {
		versionLabel = "v" + strconv.Itoa(dep.Config.VersionID)
	}
	fmt.Printf("  %s%s%s  %s  %s\n", colorBold, prefix, colorReset, dep.ID, versionLabel)
	if dep.Config.Description != "" {
		fmt.Printf("     %s%s%s\n", colorDim, dep.Config.Description, colorReset)
	}
	for _, entryPoint := range dep.EntryPoints {
		fmt.Printf("     %s%s%s", colorCyan, entryPoint.Type, colorReset)
		if entryPoint.WebApp != nil && entryPoint.WebApp.URL != "" {
			fmt.Printf("  %s", entryPoint.WebApp.URL)
		}
		fmt.Println()
	}
	if meaningfulTime(dep.UpdateTime) {
		fmt.Printf("     %supdated %s%s\n", colorDim, dep.UpdateTime.Format(time.RFC3339), colorReset)
	}
}

func printReleaseResult(title string, ver *version.Version, dep *deployment.Deployment) {
	fmt.Printf("\n  %s%sReleased%s %s\n\n", colorBold, colorGreen, colorReset, title)
	printVersion("Created version", ver)
	printDeployment("Deployment", dep)
	if url := dep.WebAppURL(); url != "" {
		fmt.Printf("\n  %sWeb app%s  %s\n", colorBold, colorReset, url)
	}
}

func meaningfulTime(t time.Time) bool {
	return !t.IsZero() && t.Year() >= 2000
}
