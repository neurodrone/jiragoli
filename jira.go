package jiragoli

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"
)

var (
	ErrProjectIDType = errors.New("projectID has to be a string or int")
)

type JIRA struct {
	client   *http.Client
	endpoint *url.URL

	projects JIRAProjects
}

// NewJIRA creates a *JIRA which we will use to pull in information
// about individual projects, issue types, issues, users among other things.
// It doesn't support OAuth in its first version and uses basic-auth. The
// user/pass for basic path are passed in as the first parameter.
// It also allows the caller to pass in a custom *http.Client. If no such
// client is passed then http.DefaultClient is used.
//
// Before *JIRA is created, this constructor will try to authenticate
// the user. If there is an issue authenticating or talking to remote
// endpoint, the function will error out.
func NewJIRA(info *url.Userinfo, jiraURL string, client *http.Client) (*JIRA, error) {
	if client == nil {
		client = http.DefaultClient
	}

	u, err := url.Parse(jiraURL)
	if err != nil {
		return nil, fmt.Errorf("error parsing jira url: %s", err)
	}

	// add the basic-auth creds to the url
	// TODO(neurodrone): Use OAuth instead once we get the API usage sorted
	// out.
	u.User = info

	// perform an initial authentication to make sure the user creds
	// have correct perms by loading all the relevant projects.
	// We have to make sure we don't modify the original URL object here
	// as we will need to store that in our JIRA struct.
	tempURL := &url.URL{}
	*tempURL = *u
	tempURL.Path += "/project"

	resp, err := client.Get(tempURL.String())
	if err != nil {
		return nil, fmt.Errorf("request to jira for auth failed: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response: %q", resp.Status)
	}

	projects := JIRAProjects{}

	if err := json.NewDecoder(resp.Body).Decode(&projects); err != nil {
		return nil, fmt.Errorf("error parsing projects: %s", err)
	}

	sort.Sort(projects)

	return &JIRA{
		client:   client,
		endpoint: u,
		projects: projects,
	}, nil
}

// Issues will get all the issues associated with a project. If the project
// doesn't exist we might receive an error.
//
// The project identifier (projectID) can both be an int or a string. If it's
// of neither int or string type then an ErrProjectIDType is returned signi-
// fying that the passed in parameter is of unrecognized type.
func (j *JIRA) Issues(projectID interface{}) (JIRAIssues, error) {
	tempURL := &url.URL{}
	*tempURL = *j.endpoint

	switch queryString := j.getQueryString(projectID); {
	case queryString == "":
		return nil, ErrProjectIDType
	default:
		tempURL.RawQuery = queryString
	}

	// There is no good way to get the issues affiliated to a project
	// from the /project endpoint. The only way this is
	// exposed is querying via JQL.
	//
	// This code is written for v6.3.15.
	tempURL.Path += "/search"

	resp, err := j.client.Get(tempURL.String())
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// If the request didn't succeed we don't continue.
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("error response: %q", resp.Status)
	}

	s := struct {
		Issues []struct {
			Key       string     `json:"key"`
			JIRAIssue *JIRAIssue `json:"fields"`
		} `json:"issues"`
	}{}

	if err := json.NewDecoder(resp.Body).Decode(&s); err != nil {
		return nil, err
	}

	// We know exactly how many issues are going to be present
	// for a particular project from the above request. Allocate
	// space for only those issues.
	jiraIssues := make(JIRAIssues, 0, len(s.Issues))

	for _, issue := range s.Issues {
		ji := issue.JIRAIssue

		ji.IssueURL, _ = url.Parse(fmt.Sprintf("%s://%s/browse/%s",
			j.endpoint.Scheme,
			j.endpoint.Host,
			issue.Key))

		ji.Key = issue.Key
		jiraIssues = append(jiraIssues, ji)
	}

	sort.Sort(jiraIssues)

	return jiraIssues, nil
}

// Projects will return list of all the active projects in JIRA that
// the user has access to.
func (j *JIRA) Projects() JIRAProjects {
	return j.projects
}

// getQueryString should return the query-string using the given
// project ID. A blank query-string is returned if (a) no such
// project exists or (b) the type of projectID is neither an int
// or a string.
func (j *JIRA) getQueryString(projectID interface{}) string {
	values := url.Values{}

	switch p := projectID.(type) {
	case int:
		values.Set("jql", fmt.Sprintf("project=%d", p))
	case string:
		for _, project := range j.projects {
			if strings.HasPrefix(strings.ToLower(project.Name), strings.ToLower(p)) {
				values.Set("jql", fmt.Sprintf("project=%s", project.ID))
				break
			}
		}
	default:
		return ""
	}

	return values.Encode()
}
