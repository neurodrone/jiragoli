jiragoli
========

### A command-line app for interaction with JIRA.

You can filter by projects, assignees, reporters or by statuses.

Filtering by projects can be done by:

```bash
$ jirago issues -project librarian
```

Filtering projects can be done by just mentioning the suffixes of project name if you are too lazy like me:

```bash
$ jirago issues -project lib
```

Filtering by assignee can be done by:

```bash
$ jirago issues -project lib -assignee alee
```

Filtering by reporter can be done by:

```bash
$ jirago issues -project lib -reporter dcelis
```

You can mix and match those options, so the following is valid:

```bash
$ jirago issues -project lib -reporter vaibhav -assignee alee 
```

You can also filter by issue status:

```bash
$ jirago issues -project lib -status done 
```

And the combination of all above flags.

```bash
$ jirago issues -project lib -reporter vaibhav -assignee alee -status done
```
--------

### Test coverage:

```bash
$ go test -cover
PASS
coverage: 100.0% of statements
ok  	github.com/neurodrone/jiragoli	0.021s
```
