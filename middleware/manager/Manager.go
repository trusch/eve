package manager

import (
	"net/http"

	"github.com/trusch/yap/middleware/registry"
	"github.com/trusch/yap/middleware/rule"
)

// Manager manages middlewares
type Manager struct {
	ruleset *rule.Set
}

// New returns a new Manager
func New() *Manager {
	return &Manager{rule.NewSet()}
}

// UpsertRule upserts a loadbalancer rule
func (mgr *Manager) UpsertRule(rule *rule.Rule) error {
	return mgr.ruleset.UpsertRule(rule)
}

// RemoveRule removes a rule from the current rule-set
func (mgr *Manager) RemoveRule(id string) error {
	return mgr.ruleset.RemoveRule(id)
}

// BuildChain returns a middleware chain which is finalized by the given next handler
func (mgr *Manager) BuildChain(req *http.Request, next http.Handler) (http.Handler, error) {
	cfgs, err := mgr.ruleset.GetMiddlewares(req)
	if err != nil {
		return nil, err
	}
	for i := len(cfgs) - 1; i >= 0; i-- {
		cfg := cfgs[i]
		h, err := registry.Create(cfg.ID, next, cfg.Opts)
		if err != nil {
			return nil, err
		}
		next = h
	}
	return next, nil
}
