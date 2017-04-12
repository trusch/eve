package etcd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/trusch/yap/config"
	lbRule "github.com/trusch/yap/loadbalancer/rule"
	mwRule "github.com/trusch/yap/middleware/rule"
)

// Client is a etcd backed config.Stream
type Client struct {
	v3         *clientv3.Client
	ctx        context.Context
	cancelFunc context.CancelFunc
	output     chan *config.Action
}

// NewClient returns a new etcd client
func NewClient(etcdAddr string) (*Client, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{etcdAddr},
		DialTimeout: 3 * time.Second,
	})
	if err != nil {
		return nil, errors.New("can not connect to etcd")
	}
	ctx, cancelFunc := context.WithCancel(context.Background())
	client := &Client{v3: cli, output: make(chan *config.Action, 32)}
	client.ctx = ctx
	client.cancelFunc = cancelFunc
	go client.backend()
	return client, nil
}

// GetChannel returns the output channel
func (client *Client) GetChannel() chan *config.Action {
	return client.output
}

// Close closes the client
func (client *Client) Close() {
	client.cancelFunc()
	client.v3.Close()
}

func (client *Client) backend() {
	lbRules, err := client.GetLoadbalancerRules()
	if err != nil {
		log.Print(err)
	}
	for _, rule := range lbRules {
		client.feedUpsertLbRuleToChannel(rule)
	}
	mwRules, err := client.GetMiddlewareRules()
	if err != nil {
		log.Print(err)
	}
	for _, rule := range mwRules {
		client.feedUpsertMwRuleToChannel(rule)
	}
	hostCfgs, err := client.GetHostConfigs()
	if err != nil {
		log.Print(err)
	}
	for _, cfg := range hostCfgs {
		client.feedUpsertHostToChannel(cfg)
	}
	certCfgs, err := client.GetCertConfigs()
	if err != nil {
		log.Print(err)
	}
	for _, cfg := range certCfgs {
		client.feedUpsertCertToChannel(cfg)
	}

	go client.watchLbRules()
	go client.watchMwRules()
	go client.watchHosts()
	go client.watchCerts()

}

func (client *Client) watchLbRules() {
	rch := client.v3.Watch(client.ctx, "/yap/lbrules", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT {
				rule, err := client.parseLbRule(ev.Kv)
				if err != nil {
					log.Print(err)
					continue
				}
				client.feedUpsertLbRuleToChannel(rule)
			} else {
				id := string(ev.Kv.Key[len("/yap/lbrules/"):])
				client.feedDeleteLbRuleToChannel(&lbRule.Rule{ID: id})
			}
		}
	}
}

func (client *Client) watchMwRules() {
	rch := client.v3.Watch(client.ctx, "/yap/mwrules", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT {
				rule, err := client.parseMwRule(ev.Kv)
				if err != nil {
					log.Print(err)
					continue
				}
				client.feedUpsertMwRuleToChannel(rule)
			} else {
				id := string(ev.Kv.Key[len("/yap/mwrules/"):])
				client.feedDeleteMwRuleToChannel(&mwRule.Rule{ID: id})
			}
		}
	}
}

func (client *Client) watchHosts() {
	rch := client.v3.Watch(client.ctx, "/yap/loadbalancer", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT {
				cfg, err := client.parseHostConfig(ev.Kv)
				if err != nil {
					log.Print(err)
					continue
				}
				client.feedUpsertHostToChannel(cfg)
			} else {
				cfg, err := client.parseHostConfig(ev.Kv)
				if err != nil {
					log.Print(err)
					continue
				}
				client.feedDeleteHostToChannel(cfg)
			}
		}
	}
}

func (client *Client) watchCerts() {
	rch := client.v3.Watch(client.ctx, "/yap/certs", clientv3.WithPrefix())
	for wresp := range rch {
		for _, ev := range wresp.Events {
			if ev.Type == mvccpb.PUT {
				cfg, err := client.parseCertConfig(ev.Kv)
				if err != nil {
					log.Print(err)
					continue
				}
				client.feedUpsertCertToChannel(cfg)
			} else {
				cfg, err := client.parseCertConfig(ev.Kv)
				if err != nil {
					log.Print(err)
					continue
				}
				client.feedDeleteCertToChannel(cfg)
			}
		}
	}
}

// GetLoadbalancerRules returns a slice of all loadbalancer rules
func (client *Client) GetLoadbalancerRules() ([]*lbRule.Rule, error) {
	resp, err := client.v3.Get(client.ctx, "/yap/lbrules", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	rules := make([]*lbRule.Rule, 0, resp.Count)
	for _, kv := range resp.Kvs {
		rule, err := client.parseLbRule(kv)
		if err != nil {
			log.Print("Error: ", err)
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// GetMiddlewareRules returns a slice of all middleware rules
func (client *Client) GetMiddlewareRules() ([]*mwRule.Rule, error) {
	resp, err := client.v3.Get(client.ctx, "/yap/mwrules", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	rules := make([]*mwRule.Rule, 0, resp.Count)
	for _, kv := range resp.Kvs {
		rule, err := client.parseMwRule(kv)
		if err != nil {
			log.Print("Error: ", err)
			continue
		}
		rules = append(rules, rule)
	}
	return rules, nil
}

// GetHostConfigs returns a slice of all host configs
func (client *Client) GetHostConfigs() ([]*config.HostConfig, error) {
	resp, err := client.v3.Get(client.ctx, "/yap/loadbalancer", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	cfgs := make([]*config.HostConfig, 0, resp.Count)
	for _, kv := range resp.Kvs {
		cfg, err := client.parseHostConfig(kv)
		if err != nil {
			log.Print("Error: ", err)
			continue
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs, nil
}

// GetCertConfigs returns a slice of all cert configs
func (client *Client) GetCertConfigs() ([]*config.CertConfig, error) {
	resp, err := client.v3.Get(client.ctx, "/yap/certs", clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}
	cfgs := make([]*config.CertConfig, 0, resp.Count)
	for _, kv := range resp.Kvs {
		cfg, err := client.parseCertConfig(kv)
		if err != nil {
			log.Print("Error: ", err)
			continue
		}
		cfgs = append(cfgs, cfg)
	}
	return cfgs, nil
}

func (client *Client) parseLbRule(kv *mvccpb.KeyValue) (*lbRule.Rule, error) {
	rule := &lbRule.Rule{}
	err := json.Unmarshal(kv.Value, rule)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing LB-rule: %v", err)
	}
	rule.ID = string(kv.Key[len("/yap/lbrules/"):])
	return rule, nil
}

func (client *Client) parseMwRule(kv *mvccpb.KeyValue) (*mwRule.Rule, error) {
	rule := &mwRule.Rule{}
	err := json.Unmarshal(kv.Value, rule)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing MW-rule: %v", err)
	}
	rule.ID = string(kv.Key[len("/yap/mwrules/"):])
	return rule, nil
}

func (client *Client) parseCertConfig(kv *mvccpb.KeyValue) (*config.CertConfig, error) {
	cfg := &config.CertConfig{}
	err := json.Unmarshal(kv.Value, cfg)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing certificate: %v", err)
	}
	cfg.ID = string(kv.Key[len("/yap/certs/"):])
	return cfg, nil
}

func (client *Client) parseHostConfig(kv *mvccpb.KeyValue) (*config.HostConfig, error) {
	// /yap/loadbalancer/example-lb/hosts/foobar http://123.123.123.123:8080
	parts := strings.Split(string(kv.Key), "/")
	if len(parts) != 6 {
		return nil, errors.New("malformed key")
	}
	return &config.HostConfig{
		ID:           parts[5],
		Loadbalancer: parts[3],
		URL:          string(kv.Value),
	}, nil
}

func (client *Client) feedUpsertLbRuleToChannel(rule *lbRule.Rule) {
	client.output <- &config.Action{
		Type:   config.UpsertLbRule,
		LbRule: rule,
	}
}

func (client *Client) feedUpsertMwRuleToChannel(rule *mwRule.Rule) {
	client.output <- &config.Action{
		Type:   config.UpsertMwRule,
		MwRule: rule,
	}
}

func (client *Client) feedUpsertHostToChannel(cfg *config.HostConfig) {
	client.output <- &config.Action{
		Type:       config.UpsertHost,
		HostConfig: cfg,
	}
}

func (client *Client) feedUpsertCertToChannel(cfg *config.CertConfig) {
	client.output <- &config.Action{
		Type:       config.UpsertCert,
		CertConfig: cfg,
	}
}

func (client *Client) feedDeleteLbRuleToChannel(rule *lbRule.Rule) {
	client.output <- &config.Action{
		Type:   config.DeleteLbRule,
		LbRule: rule,
	}
}

func (client *Client) feedDeleteMwRuleToChannel(rule *mwRule.Rule) {
	client.output <- &config.Action{
		Type:   config.DeleteMwRule,
		MwRule: rule,
	}
}

func (client *Client) feedDeleteHostToChannel(cfg *config.HostConfig) {
	client.output <- &config.Action{
		Type:       config.DeleteHost,
		HostConfig: cfg,
	}
}

func (client *Client) feedDeleteCertToChannel(cfg *config.CertConfig) {
	client.output <- &config.Action{
		Type:       config.DeleteCert,
		CertConfig: cfg,
	}
}
