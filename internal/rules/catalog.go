package rules

import "github.com/AstralDrift/codex-action-guard/internal/guard"

type Doc = guard.RuleDoc

func Catalog() []Doc {
	return guard.AllRuleDocs()
}

func Get(id string) (Doc, bool) {
	return guard.GetRuleDoc(id)
}
