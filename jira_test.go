package jiragoli

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestNewJIRA(t *testing.T) {
	server, _ := runJIRATestServer()
	defer server.Close()

	// Server that always sends 404 Not Found
	badServer := httptest.NewServer(&BADJIRATestServer{})
	defer badServer.Close()

	// Server that always sends invalid JSON
	faultyServer := httptest.NewServer(&FaultyJIRATestServer{})
	defer faultyServer.Close()

	for _, test := range []struct {
		serverURL string
		hasError  bool
	}{
		// Ensure that url Parse fails here for blank server scheme
		{"://jira.com", true},

		// Ensure that auth fails here
		{"htp://jira.com", true},

		// This URL should always generate non-200 status code
		{badServer.URL, true},

		// This should always send a faulty JSON back
		{faultyServer.URL, true},

		// This should work perfectly fine and thus no error is
		// expected here
		{server.URL, false},
	} {

		jira, err := NewJIRA(nil, test.serverURL, nil)
		switch {
		case test.hasError:
			switch err {
			case nil:
				t.Fatalf("expected error here for server: %s", test.serverURL)
			default:
				// This is expected behavior here
				t.Logf("expected error here: %s for server: %s", err, test.serverURL)
				continue
			}
		case !test.hasError && err != nil:
			t.Fatalf("initiating a new JIRA client failed: %s", err)
		}
		// Ensure the project list is sorted by name
		var prevProjectName string
		for _, project := range jira.Projects() {
			if prevProjectName == "" {
				prevProjectName = project.Name
				continue
			}

			if prevProjectName > project.Name {
				t.Errorf("expected %q to be after %q in the project order",
					prevProjectName,
					project.Name)
			}

			prevProjectName = project.Name
		}
	}
}

func TestIssues(t *testing.T) {
	server, _ := runJIRATestServer()
	defer server.Close()

	jira, err := NewJIRA(nil, server.URL, nil)
	if err != nil {
		t.Fatalf("initiating a new JIRA client failed: %s", err)
	}

	// Server that always sends 404 Not Found
	badServer := httptest.NewServer(&BADJIRATestServer{})
	defer badServer.Close()

	// Server that always sends invalid JSON
	faultyServer := httptest.NewServer(&FaultyJIRATestServer{})
	defer faultyServer.Close()

	for _, test := range []struct {
		serverURL string
		idStr     interface{}
		lenIssues int
		hasError  bool
	}{
		// Test for the case where the project exists in JIRA and has at least
		// one issue for it
		{server.URL, 101, 1, false},

		// Test for the case where the project exists in JIRA and has at least
		// one issue for it and we pass in the project name instead of id
		{server.URL, "projectA", 1, false},

		// Test for the case where the project exists in JIRA but has no issues
		// assigned to it
		{server.URL, 102, 0, true},

		// Test for the case where the project id sent is not a string or an int
		{server.URL, true, 0, true},

		// Test for the case where response doesn't return a 200OK, it should throw
		// an error
		{badServer.URL, 101, 0, true},

		// Test when client.Get itself fails
		{"htp://jira.com", 101, 0, true},

		// Test for when json encoded data fails on you
		{faultyServer.URL, 101, 0, true},
	} {
		jira.endpoint, _ = url.Parse(test.serverURL)

		issues, err := jira.Issues(test.idStr)
		switch {
		case test.hasError:
			switch err {
			case nil:
				t.Errorf("expected error for issues for id: %v", test.idStr)
			default:
				t.Logf("expected error here: %q for server: %q", err, test.serverURL)
				continue
			}
		case !test.hasError && err != nil:
			t.Errorf("unexpected error %q for id: %v", err, test.idStr)
		}

		if len(issues) != test.lenIssues {
			t.Errorf("expected %d issues, actual issues: %d",
				test.lenIssues,
				len(issues))
		}
	}
}

// runJIRATestServer is responsible for creating a stub of JIRA
// API service and help test the following endpoints:
//
// 1. "/projects"
// 2. "/search"
//
// It returns a closer func that can be used to close the test
// server once the test run completes using it.
func runJIRATestServer() (*httptest.Server, *JIRATestServer) {
	jiraTestServer := &JIRATestServer{
		[]*JIRAProject{
			// {id, key,  name} //
			{"100", "PA", "ProjectA"},
			{"101", "PB", "ProjectB"},
			{"102", "PC", "ProjectC"},
		},
		map[string]*JIRAIssue{
			"100": {
				Key:         "PA",
				Summary:     "Some ProjectA issue",
				Description: "Some ProjectA description",
				Labels:      []string{"labelA"},
				Assignee: JIRAUser{
					Name:    "UserA",
					Email:   "userA@jira.com",
					UserURL: "https://jira.com/userA",
				},
				Reporter: JIRAUser{
					Name:    "UserB",
					Email:   "userB@jira.com",
					UserURL: "https://jira.com/userB",
				},
				JIRAStatus: JIRAStatus{
					Name:        "Open",
					Description: "Open Ticket",
				},
			},
			"101": {
				Key:         "PB",
				Summary:     "Some ProjectB issue",
				Description: "Some ProjectB description",
				Labels:      []string{"labelB1", "labelB2"},
				Assignee: JIRAUser{
					Name:    "UserB",
					Email:   "userB@jira.com",
					UserURL: "https://jira.com/userB",
				},
				Reporter: JIRAUser{
					Name:    "UserA",
					Email:   "userA@jira.com",
					UserURL: "https://jira.com/userA",
				},
				JIRAStatus: JIRAStatus{
					Name:        "In Progress",
					Description: "Ticket is being reviewed",
				},
			},
		},
	}

	return httptest.NewServer(jiraTestServer), jiraTestServer
}

type JIRATestServer struct {
	projects JIRAProjects
	issues   map[string]*JIRAIssue
}

func (j *JIRATestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		path := r.URL.Path

		switch {
		case strings.HasSuffix(path, "/project"):
			j.serveJIRAProjects(w, r)
		case strings.HasSuffix(path, "/search"):
			j.serveJIRASearch(w, r)
		}
	default:
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// serveJIRAProjects is a dummy /projects endpoint that replies
// with a list of projects that a list of issues would be associated
// with.
func (j *JIRATestServer) serveJIRAProjects(w http.ResponseWriter, r *http.Request) {
	_ = json.NewEncoder(w).Encode(j.projects)
}

func (j *JIRATestServer) serveJIRASearch(w http.ResponseWriter, r *http.Request) {
	var projectID string

	val := r.FormValue("jql")
	switch n, err := fmt.Sscanf(val, "project=%s", &projectID); {
	case err != nil:
		http.Error(w, fmt.Sprintf("%s: input:%q", err, val), http.StatusBadRequest)
	case n != 1:
		http.Error(w, fmt.Sprintf("expected 1 item parsed, actual parsed: %d", n), http.StatusBadRequest)
	}

	issue, ok := j.issues[projectID]
	if !ok {
		http.Error(w, "no such project found", http.StatusNotFound)
		return
	}

	type Issues struct {
		Key       string     `json:"key"`
		JIRAIssue *JIRAIssue `json:"fields"`
	}

	issuesResponse := struct {
		Issues []Issues `json:"issues"`
	}{
		[]Issues{
			{issue.Key, issue},
		},
	}

	_ = json.NewEncoder(w).Encode(issuesResponse)
}

type BADJIRATestServer struct {
	// nothing here yet
}

func (j *BADJIRATestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "not found", http.StatusNotFound)
}

type FaultyJIRATestServer struct {
	// nothing here yet
}

func (j *FaultyJIRATestServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// This should send a broken json for every request
	fmt.Fprintf(w, "{")

	w.Header().Set("Content-type", "application/json")
}
