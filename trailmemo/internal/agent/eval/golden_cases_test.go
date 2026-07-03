package eval

import "testing"

func TestBootstrapCasesHasPhase6Coverage(t *testing.T) {
	cases := BootstrapCases()
	if len(cases) < 20 || len(cases) > 30 {
		t.Fatalf("expected 20-30 golden cases, got %d", len(cases))
	}

	requiredCategories := []string{
		"recommend",
		"route",
		"route_remix",
		"travel_note",
		"safety",
		"fallback",
		"personalize",
		"tool",
	}
	seen := make(map[string]bool)
	names := make(map[string]bool)
	for _, c := range cases {
		if c.Name == "" {
			t.Fatal("case name must not be empty")
		}
		if names[c.Name] {
			t.Fatalf("duplicate case name: %s", c.Name)
		}
		names[c.Name] = true
		if c.Description == "" {
			t.Fatalf("case %s description must not be empty", c.Name)
		}
		if len(c.Input.Messages) == 0 {
			t.Fatalf("case %s must include input messages", c.Name)
		}
		if len(c.Checks) == 0 {
			t.Fatalf("case %s must include checks", c.Name)
		}
		seen[c.Category] = true
	}
	for _, category := range requiredCategories {
		if !seen[category] {
			t.Fatalf("expected golden cases to include category %s", category)
		}
	}
}
