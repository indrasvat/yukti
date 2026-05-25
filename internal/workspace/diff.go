package workspace

import (
	"fmt"
	"sort"
)

// ChangeKind is the local workspace state relative to the last remote snapshot.
type ChangeKind string

const (
	ChangeAdded     ChangeKind = "added"
	ChangeModified  ChangeKind = "modified"
	ChangeDeleted   ChangeKind = "deleted"
	ChangeUnchanged ChangeKind = "unchanged"
)

// Change describes one local file status entry.
type Change struct {
	Kind ChangeKind
	Path string
}

// Diff compares local files with a manifest snapshot.
func Diff(files []LocalFile, manifest *Manifest) []Change {
	local := make(map[string]LocalFile, len(files))
	for _, file := range files {
		local[file.Path] = file
	}

	paths := make(map[string]struct{}, len(local)+len(manifest.Files))
	for path := range local {
		paths[path] = struct{}{}
	}
	for path := range manifest.Files {
		paths[path] = struct{}{}
	}

	ordered := make([]string, 0, len(paths))
	for path := range paths {
		ordered = append(ordered, path)
	}
	sort.Strings(ordered)

	changes := make([]Change, 0, len(ordered))
	for _, path := range ordered {
		file, localOK := local[path]
		state, manifestOK := manifest.Files[path]

		switch {
		case localOK && !manifestOK:
			changes = append(changes, Change{Kind: ChangeAdded, Path: path})
		case !localOK && manifestOK:
			changes = append(changes, Change{Kind: ChangeDeleted, Path: path})
		case file.Hash != state.Hash:
			changes = append(changes, Change{Kind: ChangeModified, Path: path})
		default:
			changes = append(changes, Change{Kind: ChangeUnchanged, Path: path})
		}
	}
	return changes
}

// Dirty returns true when any change is not unchanged.
func Dirty(changes []Change) bool {
	for _, change := range changes {
		if change.Kind != ChangeUnchanged {
			return true
		}
	}
	return false
}

// Summary returns compact counts by change kind.
func Summary(changes []Change) string {
	counts := map[ChangeKind]int{}
	for _, change := range changes {
		counts[change.Kind]++
	}
	return fmt.Sprintf("%d added, %d modified, %d deleted, %d unchanged",
		counts[ChangeAdded],
		counts[ChangeModified],
		counts[ChangeDeleted],
		counts[ChangeUnchanged],
	)
}
