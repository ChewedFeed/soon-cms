package retro

import (
	"testing"
)

func TestGetProgressColor(t *testing.T) {
	tests := []struct {
		name     string
		total    int
		complete int
		expected float32
	}{
		{"no tasks", 0, 0, 0},
		{"all complete", 4, 4, 100},
		{"half complete", 4, 2, 50},
		{"one of three", 3, 1, 33.333336},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.total == 0 {
				if tt.expected != 0 {
					t.Error("expected 0 for no tasks")
				}
				return
			}
			result := (float32(tt.complete) / float32(tt.total)) * 100
			if result != tt.expected {
				t.Errorf("expected %f, got %f", tt.expected, result)
			}
		})
	}
}

func TestProjectLinkType(t *testing.T) {
	link := ProjectLink{
		LinkType: "dashboard",
		URL:      "https://dashboard.flags.gg",
		Label:    "Dashboard",
	}

	if link.LinkType != "dashboard" {
		t.Errorf("expected 'dashboard', got '%s'", link.LinkType)
	}
	if link.URL != "https://dashboard.flags.gg" {
		t.Errorf("expected 'https://dashboard.flags.gg', got '%s'", link.URL)
	}
}

func TestRoadmapItemType(t *testing.T) {
	target := "2026-06-01"
	item := RoadmapItem{
		Name:       "Beta release",
		TargetDate: &target,
		Completed:  false,
		SortOrder:  1,
	}

	if item.Name != "Beta release" {
		t.Errorf("expected 'Beta release', got '%s'", item.Name)
	}
	if item.Completed {
		t.Error("expected not completed")
	}
	if *item.TargetDate != "2026-06-01" {
		t.Errorf("expected '2026-06-01', got '%s'", *item.TargetDate)
	}
}

func TestServiceStruct(t *testing.T) {
	svc := Service{
		ID:          1,
		Name:        "Flags.gg",
		Description: "Feature flags",
		URL:         "https://flags.gg",
		LaunchDate:  LaunchDate{Year: 2025, Month: 6, Day: 15},
		Progress:    75,
		Icon:        "solid fa-flag",
		Launched:    true,
		Links: []ProjectLink{
			{LinkType: "main", URL: "https://flags.gg", Label: "Flags.gg"},
			{LinkType: "dashboard", URL: "https://dashboard.flags.gg", Label: "Dashboard"},
			{LinkType: "docs", URL: "https://docs.flags.gg", Label: "Docs"},
		},
		Roadmap: []RoadmapItem{
			{Name: "Alpha", Completed: true, SortOrder: 0},
			{Name: "Beta", Completed: false, SortOrder: 1},
		},
	}

	if len(svc.Links) != 3 {
		t.Errorf("expected 3 links, got %d", len(svc.Links))
	}
	if len(svc.Roadmap) != 2 {
		t.Errorf("expected 2 roadmap items, got %d", len(svc.Roadmap))
	}
	if svc.Links[1].LinkType != "dashboard" {
		t.Errorf("expected 'dashboard', got '%s'", svc.Links[1].LinkType)
	}
	if !svc.Roadmap[0].Completed {
		t.Error("expected first roadmap item to be completed")
	}
}

func TestLaunchDate(t *testing.T) {
	ld := LaunchDate{Year: 2026, Month: 3, Day: 20}
	if ld.Year != 2026 || ld.Month != 3 || ld.Day != 20 {
		t.Errorf("unexpected launch date: %+v", ld)
	}
}
