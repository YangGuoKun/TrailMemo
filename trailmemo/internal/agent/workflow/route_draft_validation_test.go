package workflow

import "testing"

func TestParseRouteDraftArtifactExtractsJSONFromMarkdown(t *testing.T) {
	raw := "好的，路线如下：\n```json\n{\"title\":\"杭州美食线\",\"summary\":\"轻松逛吃\",\"start_city\":\"杭州\",\"end_city\":\"杭州\",\"estimated_budget\":800,\"estimated_hours\":8,\"checkpoints\":[{\"name\":\"西湖\",\"city\":\"杭州\",\"address\":\"西湖区\",\"sequence\":9,\"arrive_time\":\"Day1 09:00\",\"stay_duration\":90,\"description\":\"散步\"}]}\n```\n祝你玩得开心"

	artifact, err := parseRouteDraftArtifact(raw)
	if err != nil {
		t.Fatalf("expected parse success, got %v", err)
	}
	if artifact.Title != "杭州美食线" {
		t.Fatalf("unexpected title: %s", artifact.Title)
	}
	if len(artifact.Checkpoints) != 1 {
		t.Fatalf("expected 1 checkpoint, got %d", len(artifact.Checkpoints))
	}
	if artifact.Checkpoints[0].Sequence != 1 {
		t.Fatalf("expected normalized sequence 1, got %d", artifact.Checkpoints[0].Sequence)
	}
}

func TestParseRouteDraftArtifactRejectsMissingCheckpoints(t *testing.T) {
	_, err := parseRouteDraftArtifact(`{"title":"空路线","summary":"无","start_city":"杭州","end_city":"杭州"}`)
	if err == nil {
		t.Fatal("expected missing checkpoints error")
	}
}
