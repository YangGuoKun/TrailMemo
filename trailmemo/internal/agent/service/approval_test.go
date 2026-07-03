package service

import "testing"

func TestCanCommitArtifactStatus(t *testing.T) {
	tests := []struct {
		name          string
		status        string
		needsApproval bool
		want          bool
	}{
		{name: "write requires approved", status: "approved", needsApproval: true, want: true},
		{name: "write rejects draft", status: "draft", needsApproval: true, want: false},
		{name: "read artifact allows draft", status: "draft", needsApproval: false, want: true},
		{name: "committed is not commit-ready", status: "committed", needsApproval: true, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canCommitArtifactStatus(tt.status, tt.needsApproval); got != tt.want {
				t.Fatalf("expected %v, got %v", tt.want, got)
			}
		})
	}
}
