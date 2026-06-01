package guard

import "regexp"

type untrustedPattern struct {
	name    string
	pattern *regexp.Regexp
}

type sinkPattern struct {
	pattern string
	kind    string
}

var codexExecPattern = regexp.MustCompile(`(?m)(^|[;&|[:space:]])codex[[:space:]]+exec\b`)

var untrustedPatterns = []untrustedPattern{
	{"pull request title", regexp.MustCompile(`github\.event\.pull_request\.title`)},
	{"pull request body", regexp.MustCompile(`github\.event\.pull_request\.body`)},
	{"pull request head ref", regexp.MustCompile(`github\.event\.pull_request\.head\.(ref|label|sha|repo\.full_name)`)},
	{"issue title", regexp.MustCompile(`github\.event\.issue\.title`)},
	{"issue body", regexp.MustCompile(`github\.event\.issue\.body`)},
	{"comment body", regexp.MustCompile(`github\.event\.comment\.body`)},
	{"discussion content", regexp.MustCompile(`github\.event\.discussion(_comment)?\.(title|body)`)},
	{"branch name", regexp.MustCompile(`github\.(head_ref|ref_name)`)},
	{"commit message", regexp.MustCompile(`github\.event\.(head_commit\.message|commits)`)},
	{"workflow run text", regexp.MustCompile(`github\.event\.workflow_run\.(display_title|head_branch|head_sha|name)`)},
	{"artifact-derived text", regexp.MustCompile(`(?i)(download-artifact|artifact.*(body|comment|prompt|message|summary))`)},
}

var repoControlledRunPatterns = []string{
	"go test",
	"npm ",
	"pnpm ",
	"yarn ",
	"make",
	"./",
	"bash scripts",
	"sh scripts",
	"python ",
	"pip install",
	"bundle ",
	"cargo ",
	"mvn ",
	"gradle ",
}

var gatePatterns = []string{
	"github.actor",
	"allow-users",
	"allow-bots",
	"author_association",
	"maintainer",
	"labels.*.name",
	"contains(github.event.pull_request.labels",
	"environment:",
}

var stepGatePatterns = []string{
	"allow-users",
	"github.actor",
	"maintainer",
	"labels",
}

var promptFileWriteMarkers = []string{">", "tee ", "cat <<", "printf ", "echo "}

var untrustedTriggerNames = []string{
	"pull_request",
	"pull_request_target",
	"issue_comment",
	"issues",
	"discussion",
	"discussion_comment",
	"workflow_run",
}

var trustedOnlyTriggerNames = map[string]bool{
	"workflow_dispatch": true,
	"schedule":          true,
}

var untrustedCheckoutRefs = []string{
	"github.event.pull_request.head",
	"github.head_ref",
	"refs/pull/",
	"github.event.workflow_run.head_sha",
	"github.event.workflow_run.head_branch",
}

var sensitiveSinkPatterns = []sinkPattern{
	{"actions/github-script", "github-script automation"},
	{"gh pr", "GitHub CLI PR automation"},
	{"gh issue", "GitHub CLI issue automation"},
	{"gh release", "GitHub CLI release automation"},
	{"git push", "git push"},
	{"npm publish", "package publish"},
	{"pypi-publish", "package publish"},
	{"docker push", "container publish"},
	{"kubectl", "deployment command"},
	{"deploy", "deployment action or command"},
	{"createcomment", "GitHub comment automation"},
	{"issues.createcomment", "GitHub issue comment automation"},
	{"pulls.merge", "GitHub merge automation"},
	{"issues.addlabels", "GitHub label automation"},
	{"softprops/action-gh-release", "GitHub release automation"},
	{"peter-evans/create-pull-request", "pull request write automation"},
}

var postingSinkPatterns = []string{
	"comment",
	"createcomment",
	"github_step_summary",
	"release",
	"body:",
	"body =",
	"pr comment",
	"issue comment",
}

var outputConstraintPatterns = []string{
	"output-schema",
	"jq",
	"ajv",
	"jsonschema",
	"head -c",
	"truncate",
	"wc -c",
	"substring",
	"substr",
	"slice(",
	"escape",
	"redact",
	"::add-mask::",
	"mask",
}
