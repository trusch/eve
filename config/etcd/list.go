package etcd

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/coreos/etcd/clientv3"

	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/trusch/eve/config"
	lbRule "github.com/trusch/eve/loadbalancer/rule"
	mwRule "github.com/trusch/eve/middleware/rule"
)

// GetLoadbalancerRules returns a slice of all loadbalancer rules
func (client *Client) GetLoadbalancerRules() ([]*lbRule.Rule, error) {
	resp, err := client.v3.Get(client.ctx, "/eve/lbrules", clientv3.WithPrefix())
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
	resp, err := client.v3.Get(client.ctx, "/eve/mwrules", clientv3.WithPrefix())
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
	resp, err := client.v3.Get(client.ctx, "/eve/loadbalancer", clientv3.WithPrefix())
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
	resp, err := client.v3.Get(client.ctx, "/eve/certs", clientv3.WithPrefix())
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
	rule.ID = string(kv.Key[len("/eve/lbrules/"):])
	return rule, nil
}

func (client *Client) parseMwRule(kv *mvccpb.KeyValue) (*mwRule.Rule, error) {
	rule := &mwRule.Rule{}
	err := json.Unmarshal(kv.Value, rule)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing MW-rule: %v", err)
	}
	rule.ID = string(kv.Key[len("/eve/mwrules/"):])
	return rule, nil
}

func (client *Client) parseCertConfig(kv *mvccpb.KeyValue) (*config.CertConfig, error) {
	cfg := &config.CertConfig{}
	err := json.Unmarshal(kv.Value, cfg)
	if err != nil {
		return nil, fmt.Errorf("Error while parsing certificate: %v", err)
	}
	cfg.ID = string(kv.Key[len("/eve/certs/"):])
	return cfg, nil
}

func (client *Client) parseHostConfig(kv *mvccpb.KeyValue) (*config.HostConfig, error) {
	// /eve/loadbalancer/example-lb/hosts/foobar http://123.123.123.123:8080
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
