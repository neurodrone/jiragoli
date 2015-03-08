package jiragoli

import (
	"sort"
	"testing"
)

func TestProjectsSort(t *testing.T) {
	expectedOrder := []string{
		"100",
		"101",
		"102",
		"103",
		"104",
	}

	projects := make(JIRAProjects, 0, len(expectedOrder))

	for _, project := range []struct {
		id, name string
	}{
		{"100", "AProject"},
		{"104", "EProject"},
		{"102", "CProject"},
		{"103", "DProject"},
		{"101", "BProject"},

		// The above is a shuffled order from the expected order
	} {
		projects = append(projects, &JIRAProject{ID: project.id, Name: project.name})
	}

	sort.Sort(projects)

	for i, id := range expectedOrder {
		if id != projects[i].ID {
			t.Errorf("expected ID: %v, actual: %v",
				id,
				projects[i].ID)
		}
	}
}
