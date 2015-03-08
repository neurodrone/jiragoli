package jiragoli

import (
	"sort"
	"testing"
	"time"
)

func TestIssuesSort(t *testing.T) {
	expectedOrder := []string{
		"PRJ-3",
		"PRJ-8",
		"PRJ-6",
		"PRJ-1",
		"PRJ-2",
	}

	issues := make(JIRAIssues, 0, len(expectedOrder))

	for _, issue := range []struct {
		key       string
		createdAt string
	}{
		{"PRJ-1", "2015-03-05T16:54:07.000+0000"},
		{"PRJ-3", "2015-03-01T16:54:07.000+0000"},
		{"PRJ-8", "2015-03-02T16:54:07.000+0000"},
		{"PRJ-2", "2015-03-06T16:54:07.000+0000"},
		{"PRJ-6", "2015-03-04T16:54:07.000+0000"},

		// The above is a shuffled order from the expected order
	} {
		tm, err := time.Parse("2006-01-02T15:04:05.000-0700", issue.createdAt)
		if err != nil {
			t.Fatalf("unexpected error in parsing time-string: %s", err)
		}

		issues = append(issues, &JIRAIssue{Key: issue.key, CreatedAt: tm})
	}

	sort.Sort(issues)

	for i, key := range expectedOrder {
		if key != issues[i].Key {
			t.Errorf("expected key: %q, actual: %q",
				key,
				issues[i].Key)
		}
	}
}
