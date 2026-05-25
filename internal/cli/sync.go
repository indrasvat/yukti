package cli

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"yukti/internal/workspace"
)

var (
	syncDir      string
	syncForce    bool
	syncParentID string
)

var newCmd = &cobra.Command{
	Use:   "new <title>",
	Short: "Create a new Apps Script project and local workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		repo, err := authenticatedProjectRepo(ctx)
		if err != nil {
			return err
		}

		result, err := workspace.NewService(repo).Create(ctx, workspace.CreateOptions{
			Title:    args[0],
			Dir:      syncDir,
			ParentID: syncParentID,
			Force:    syncForce,
		})
		if err != nil {
			return err
		}
		printSyncResult("Created", result)
		return nil
	},
}

var cloneCmd = &cobra.Command{
	Use:   "clone <script-id>",
	Short: "Clone an Apps Script project into a local workspace",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		repo, err := authenticatedProjectRepo(ctx)
		if err != nil {
			return err
		}

		result, err := workspace.NewService(repo).Clone(ctx, workspace.CloneOptions{
			ScriptID: args[0],
			Dir:      syncDir,
			Force:    syncForce,
		})
		if err != nil {
			return err
		}
		printSyncResult("Cloned", result)
		return nil
	},
}

var pullCmd = &cobra.Command{
	Use:   "pull",
	Short: "Pull remote HEAD files into the current Yukti workspace",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		repo, err := authenticatedProjectRepo(ctx)
		if err != nil {
			return err
		}

		result, err := workspace.NewService(repo).Pull(ctx, workspace.PullOptions{
			Dir:   syncDir,
			Force: syncForce,
		})
		if err != nil {
			return err
		}
		printSyncResult("Pulled", result)
		return nil
	},
}

var pushCmd = &cobra.Command{
	Use:   "push",
	Short: "Push local Yukti workspace files to remote HEAD",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		repo, err := authenticatedProjectRepo(ctx)
		if err != nil {
			return err
		}

		result, err := workspace.NewService(repo).Push(ctx, workspace.PushOptions{
			Dir:   syncDir,
			Force: syncForce,
		})
		if err != nil {
			return err
		}
		printChanges("Pushed", result.Changes)
		printSyncResult("Synced", result)
		return nil
	},
}

var diffCmd = &cobra.Command{
	Use:   "diff",
	Short: "Show local workspace changes since the last pull or push",
	RunE: func(cmd *cobra.Command, args []string) error {
		result, err := workspace.NewService(nil).Status(syncDir)
		if err != nil {
			return err
		}

		printWorkspaceHeader(result)
		printChanges("Local changes", result.Changes)
		if !workspace.Dirty(result.Changes) {
			fmt.Println("  clean")
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(newCmd, cloneCmd, pullCmd, pushCmd, diffCmd)

	for _, cmd := range []*cobra.Command{newCmd, cloneCmd, pullCmd, pushCmd, diffCmd} {
		cmd.Flags().StringVarP(&syncDir, "dir", "d", "", "Workspace directory (defaults to current directory where applicable)")
		cmd.Flags().BoolVarP(&syncForce, "force", "f", false, "Overwrite safety checks for this operation")
	}
	newCmd.Flags().StringVar(&syncParentID, "parent-id", "", "Drive file ID to create a bound Apps Script project")
}

func printSyncResult(action string, result *workspace.Result) {
	printWorkspaceHeader(result)
	fmt.Printf("  %s %d files in %s\n", action, len(result.Files), displayPath(result.Dir))
}

func printWorkspaceHeader(result *workspace.Result) {
	fmt.Printf("\n  %s%s%s\n", colorBold, result.Title, colorReset)
	fmt.Printf("  %sScript ID%s  %s\n", colorDim, colorReset, result.ScriptID)
}

func printChanges(title string, changes []workspace.Change) {
	fmt.Printf("\n  %s%s%s\n", colorBold, title, colorReset)
	for _, change := range changes {
		if change.Kind == workspace.ChangeUnchanged {
			continue
		}
		fmt.Printf("  %s  %s\n", changeMarker(change.Kind), change.Path)
	}
	fmt.Printf("  %s%s%s\n", colorDim, workspace.Summary(changes), colorReset)
}

func changeMarker(kind workspace.ChangeKind) string {
	switch kind {
	case workspace.ChangeAdded:
		return colorGreen + "A" + colorReset
	case workspace.ChangeModified:
		return colorYellow + "M" + colorReset
	case workspace.ChangeDeleted:
		return colorRed + "D" + colorReset
	default:
		return colorDim + "-" + colorReset
	}
}

func displayPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return path
	}
	home, err := os.UserHomeDir()
	if err == nil && strings.HasPrefix(abs, home) {
		return "~" + strings.TrimPrefix(abs, home)
	}
	return abs
}
