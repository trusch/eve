package rule

import (
	"errors"
	"net/http"

	"github.com/vulcand/route"
)

// A Rule represents one loadbalacer rule
type Rule struct {
	ID     string
	Route  string
	Target string
}

// A Set is a set of rules which match requests to loadbalancers
type Set struct {
	rules  map[string]*Rule
	router route.Router
}

// UpsertRule upserts a rule
func (rs *Set) UpsertRule(rule *Rule) error {
	rs.rules[rule.ID] = rule
	return rs.router.UpsertRoute(rule.Route, rule.Target)
}

// RemoveRule removes a rule
func (rs *Set) RemoveRule(id string) error {
	rule, ok := rs.rules[id]
	if !ok {
		return errors.New("rule not found")
	}
	delete(rs.rules, id)
	return rs.router.RemoveRoute(rule.Route)
}

// GetTarget returns the target loadbalancer ID for a request
func (rs *Set) GetTarget(req *http.Request) (string, error) {
	target, err := rs.router.Route(req)
	if err != nil {
		return "", err
	}
	if target == nil {
		return "", errors.New("no matching loadbalancer rule")
	}
	return target.(string), nil
}

// New returns a new rule object
func New(id, route, target string) *Rule {
	return &Rule{id, route, target}
}

// NewSet returns a new, empty rule set
func NewSet() *Set {
	return &Set{
		rules:  make(map[string]*Rule),
		router: route.New(),
	}
}
