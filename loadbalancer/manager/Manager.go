package manager

import (
	"errors"
	"net/http"
	"net/url"

	"github.com/trusch/eve/config"
	"github.com/trusch/eve/loadbalancer/rule"
	"github.com/vulcand/oxy/forward"
	"github.com/vulcand/oxy/roundrobin"
)

// Manager manages available loadbalancers
type Manager struct {
	loadbalancers map[string]*roundrobin.Rebalancer
	ruleset       *rule.Set
	hosts         map[string]*config.HostConfig
}

// NewLoadbalancer returns a new Loadbalancer. In facts its a chain: rebalancer -> roundrobin -> forward
func NewLoadbalancer() *roundrobin.Rebalancer {
	fwd, _ := forward.New()
	lb, _ := roundrobin.New(fwd)
	rb, _ := roundrobin.NewRebalancer(lb)
	return rb
}

// New returns a new LB Manager
func New() *Manager {
	return &Manager{
		loadbalancers: make(map[string]*roundrobin.Rebalancer),
		ruleset:       rule.NewSet(),
		hosts:         make(map[string]*config.HostConfig),
	}
}

// UpsertServer upserts a server at a specific loadbalancer
// if the lb doesn't exist, it is created
func (mgr *Manager) UpsertServer(cfg *config.HostConfig) error {
	if oldCfg, ok := mgr.hosts[cfg.ID]; ok {
		mgr.RemoveServer(oldCfg)
	}
	mgr.hosts[cfg.ID] = cfg
	lb, ok := mgr.loadbalancers[cfg.Loadbalancer]
	if !ok {
		lb = NewLoadbalancer()
		mgr.loadbalancers[cfg.Loadbalancer] = lb
	}
	url, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}
	return lb.UpsertServer(url)
}

// RemoveServer removes a server from a specific loadbalancer
func (mgr *Manager) RemoveServer(cfg *config.HostConfig) error {
	lb, ok := mgr.loadbalancers[cfg.Loadbalancer]
	if !ok {
		return errors.New("loadbalancer doesn't exist")
	}
	url, err := url.Parse(cfg.URL)
	if err != nil {
		return err
	}
	return lb.RemoveServer(url)
}

// UpsertRule upserts a loadbalancer rule
func (mgr *Manager) UpsertRule(rule *rule.Rule) error {
	return mgr.ruleset.UpsertRule(rule)
}

// RemoveRule removes a rule from the current rule-set
func (mgr *Manager) RemoveRule(id string) error {
	return mgr.ruleset.RemoveRule(id)
}

// ServeHTTP serves HTTP requests by finding the correct loadbalancer and calling it
func (mgr *Manager) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	target, err := mgr.ruleset.GetTarget(req)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(err.Error()))
		return
	}
	lb := mgr.loadbalancers[target]
	lb.ServeHTTP(w, req)
}
