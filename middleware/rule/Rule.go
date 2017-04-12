package rule

import (
	"errors"
	"net/http"

	"github.com/vulcand/route"
)

// A Rule represents one loadbalacer rule
type Rule struct {
	ID          string
	Route       string
	Middlewares []*Config
}

// Config is the config for one middleware
type Config struct {
	ID   string
	Opts interface{}
}

// A Set is a set of rules which match requests to middleware configs
type Set struct {
	rules  map[string]*Rule
	router route.Router
}

// UpsertRule upserts a rule
func (rs *Set) UpsertRule(rule *Rule) error {
	rs.rules[rule.ID] = rule
	return rs.router.UpsertRoute(rule.Route, rule.Middlewares)
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

// GetMiddlewares returns the target middleware config array for a request
func (rs *Set) GetMiddlewares(req *http.Request) ([]*Config, error) {
	target, err := rs.router.Route(req)
	if err != nil {
		return nil, err
	}
	if target == nil {
		return nil, nil
	}
	return target.([]*Config), nil
}

// New returns a new rule object
func New(id, route string, middlewares []*Config) *Rule {
	return &Rule{id, route, middlewares}
}

// NewSet returns a new, empty rule set
func NewSet() *Set {
	return &Set{
		rules:  make(map[string]*Rule),
		router: route.New(),
	}
}
