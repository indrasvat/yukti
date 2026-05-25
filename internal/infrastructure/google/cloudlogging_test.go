package google

import "testing"

func TestMatchesFunctionNameKeepsUnlabeledLogs(t *testing.T) {
	t.Parallel()

	entry := LogEntry{Message: "unlabeled"}
	if !matchesFunctionName(entry, "main") {
		t.Fatal("unlabeled entries should be kept because Apps Script labels are unreliable")
	}
}

func TestMatchesFunctionNameFiltersMismatchedLabels(t *testing.T) {
	t.Parallel()

	entry := LogEntry{FunctionName: "other", Message: "wrong function"}
	if matchesFunctionName(entry, "main") {
		t.Fatal("labeled entries from other functions should be filtered")
	}
}

func TestMatchesFunctionNameKeepsMatchingLabels(t *testing.T) {
	t.Parallel()

	entry := LogEntry{FunctionName: "main", Message: "right function"}
	if !matchesFunctionName(entry, "main") {
		t.Fatal("matching labeled entry should be kept")
	}
}
