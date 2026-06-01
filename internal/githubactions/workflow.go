package githubactions

import "gopkg.in/yaml.v3"

type Workflow struct {
	File     string
	Root     *yaml.Node
	Triggers []string
	Jobs     []Job
}

type Job struct {
	ID          string
	Node        *yaml.Node
	Line        int
	Permissions PermissionSummary
	Steps       []Step
}

type Step struct {
	Index int
	Node  *yaml.Node
	Name  string
	ID    string
	Uses  string
	Run   string
	With  map[string]*yaml.Node
	Line  int
}
