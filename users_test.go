package jiragoli

import "testing"

func TestUsersString(t *testing.T) {
	for _, test := range []struct {
		name, email string
		output      string
	}{
		{"userA", "userA@jira.com", "UserA (userA@jira.com)"},
		{"", "userB@jira.com", "(userB@jira.com)"},
		{"userC", "", "UserC ()"},
	} {
		user := JIRAUser{test.name, test.email, ""}

		if test.output != user.String() {
			t.Errorf("expected user string: %q, actual: %q",
				test.output,
				user.String())
		}
	}
}
