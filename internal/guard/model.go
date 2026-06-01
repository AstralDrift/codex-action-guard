package guard

import (
	"github.com/AstralDrift/codex-action-guard/internal/githubactions"
	"gopkg.in/yaml.v3"
)

type workflowInfo struct {
	repoRoot string
	absFile  string
	relFile  string
	raw      string
	lines    []string
	doc      *yaml.Node
	triggers map[string]bool
}

type jobInfo struct {
	id             string
	node           *yaml.Node
	line           int
	permissions    permissionContext
	envSecrets     []envSecret
	steps          []stepInfo
	hasCheckout    bool
	repoControlled bool
	hasGate        bool
	hasEnvironment bool
	secretUse      bool
}

type stepInfo struct {
	index int
	node  *yaml.Node
	line  int
	id    string
	name  string
	uses  string
	run   string
	with  map[string]*yaml.Node
}

type codexInvocation struct {
	summary          Invocation
	job              *jobInfo
	step             stepInfo
	promptSources    []promptSource
	codexArgs        string
	outputSchemaText string
	hasSchema        bool
}

type promptSource struct {
	File         string
	Text         string
	Raw          string
	Lines        []string
	FallbackLine int
	Boundary     string
	Description  string
}

type permissionContext = githubactions.PermissionSummary

type envSecret struct {
	Key  string
	Line int
}

type untrustedMatch struct {
	Source      string
	File        string
	Line        int
	Text        string
	Lines       []string
	Boundary    string
	Description string
}

type sinkInfo struct {
	Line          int
	Kind          string
	Detail        string
	Snippet       string
	Posting       bool
	HasConstraint bool
}
