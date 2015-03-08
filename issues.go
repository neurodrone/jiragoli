package jiragoli

import (
	"net/url"
	"time"
)

// JIRAIssues are sortable and can be ordered based on their created
// dates.
type JIRAIssues []*JIRAIssue

func (jis JIRAIssues) Len() int      { return len(jis) }
func (jis JIRAIssues) Swap(i, j int) { jis[i], jis[j] = jis[j], jis[i] }
func (jis JIRAIssues) Less(i, j int) bool {
	return jis[i].CreatedAt.Before(jis[j].CreatedAt)
}

type JIRAIssue struct {
	Key                  string
	Summary              string   `json:"summary"`
	Description          string   `json:"description"`
	Labels               []string `json:"labels"`
	Assignee             JIRAUser `json:"assignee"`
	Reporter             JIRAUser `json:"reporter"`
	IssueURL             *url.URL
	CreatedAt, UpdatedAt time.Time
	JIRAStatus           JIRAStatus `json:"status"`
}
