package main

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/ianschenck/envflag"
	"github.com/neurodrone/jiragoli"
)

func main() {
	var (
		jiraUser = envflag.String("JIRAUSER", "", "username for jira")
		jiraPass = envflag.String("JIRAPASS", "", "password for jira")
		jiraURL  = envflag.String("JIRAURL", "", "url for jira")
	)
	envflag.Parse()

	if *jiraUser == "" || *jiraPass == "" {
		log.Fatalln("both JIRA user and password should be set to auth")
	}

	if *jiraURL == "" {
		log.Fatalln("URL for JIRA endpoint needs to be provided")
	}

	info := url.UserPassword(*jiraUser, *jiraPass)

	jira, err := jiragoli.NewJIRA(info, *jiraURL, nil)
	if err != nil {
		log.Fatalln(err)
	}

	app := cli.NewApp()
	app.Name = "JiraGoli"
	app.Usage = "Munch them sweet issues"
	app.Commands = []cli.Command{
		{
			Name:  "issues",
			Usage: "display or update JIRA issues",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "project,p",
					Value: "",
					Usage: "list issues specific to a project",
				},
				cli.StringFlag{
					Name:  "assignee",
					Value: "",
					Usage: "list issues assigned to specific person",
				},
				cli.StringFlag{
					Name:  "reporter",
					Value: "",
					Usage: "list issues reported by specific person",
				},
				cli.StringFlag{
					Name:  "status",
					Value: "",
					Usage: "list issues that have a specific status",
				},
			},
			Action: func(c *cli.Context) {
				assignee := c.String("assignee")
				reporter := c.String("reporter")
				status := c.String("status")

				switch project := c.String("project"); {
				case project == "":
					log.Println("value of -project cannot be empty")
					return
				default:
					issues, err := jira.Issues(project)
					if err != nil {
						log.Println("error:", err)
						return
					}

					issueCount := 0

					for _, issue := range issues {
						if assignee != "" && !strings.Contains(strings.ToLower(issue.Assignee.Name), assignee) {
							continue
						}

						if reporter != "" && !strings.Contains(strings.ToLower(issue.Reporter.Name), reporter) {
							continue
						}

						if status != "" && !strings.Contains(strings.ToLower(issue.JIRAStatus.Name), status) {
							continue
						}

						log.Printf("[%s] %s\n", issue.Key, issue.Summary)
						log.Printf("Status: %s\n", strings.ToUpper(issue.JIRAStatus.Name))
						log.Printf("Reported by: %s\n", issue.Reporter)

						if issue.Assignee.Name != "" {
							log.Printf("Assigned to: %s\n", issue.Assignee)
						}

						log.Printf("Labels: ['%s']\n", strings.Join(issue.Labels, "', '"))
						log.Printf("Permalink: %q\n", issue.IssueURL)
						log.Println()

						issueCount++
					}

					log.Printf("Total matching issues found: %d\n", issueCount)
				}
			},
		},
	}

	app.Run(os.Args)
}
