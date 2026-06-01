package githubactions

type Workflow struct {
	File     string
	Triggers []string
	Jobs     []Job
}

type Job struct {
	ID          string
	Line        int
	Permissions PermissionSummary
	Steps       []Step
}

type Step struct {
	Name string
	ID   string
	Uses string
	Run  string
	Line int
}
