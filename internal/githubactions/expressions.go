package githubactions

import "regexp"

var untrustedExpressionPatterns = []*regexp.Regexp{
	regexp.MustCompile(`github\.event\.pull_request\.(title|body|head\.(ref|label|sha|repo\.full_name))`),
	regexp.MustCompile(`github\.event\.issue\.(title|body)`),
	regexp.MustCompile(`github\.event\.comment\.body`),
	regexp.MustCompile(`github\.event\.discussion(_comment)?\.(title|body)`),
	regexp.MustCompile(`github\.(head_ref|ref_name)`),
	regexp.MustCompile(`github\.event\.(head_commit\.message|commits)`),
	regexp.MustCompile(`github\.event\.workflow_run\.(display_title|head_branch|head_sha|name)`),
}

func ContainsUntrustedExpression(text string) bool {
	for _, pattern := range untrustedExpressionPatterns {
		if pattern.MatchString(text) {
			return true
		}
	}
	return false
}
