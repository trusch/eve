package etcd

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/coreos/etcd/clientv3"
	"github.com/trusch/bobbyd/config"
	lbRule "github.com/trusch/bobbyd/loadbalancer/rule"
	mwRule "github.com/trusch/bobbyd/middleware/rule"
)

// PutLbRule sets a loadbalancer rule
func (client *Client) PutLbRule(rule *lbRule.Rule, persistent bool) error {
	key := fmt.Sprintf("/bobbyd/lbrules/%v", rule.ID)
	bs, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	val := string(bs)
	return client.put(key, val, persistent)
}

// PutMwRule sets a middleware rule
func (client *Client) PutMwRule(rule *mwRule.Rule, persistent bool) error {
	key := fmt.Sprintf("/bobbyd/mwrules/%v", rule.ID)
	bs, err := json.Marshal(rule)
	if err != nil {
		return err
	}
	val := string(bs)
	return client.put(key, val, persistent)
}

// PutHostConfig sets a hostconfig
func (client *Client) PutHostConfig(cfg *config.HostConfig, persistent bool) error {
	key := fmt.Sprintf("/bobbyd/loadbalancer/%v/hosts/%v", cfg.Loadbalancer, cfg.ID)
	val := cfg.URL
	return client.put(key, val, persistent)

}

// PutCertConfig sets a cert config
func (client *Client) PutCertConfig(cfg *config.CertConfig, persistent bool) error {
	key := fmt.Sprintf("/bobbyd/certs/%v", cfg.ID)
	bs, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	val := string(bs)
	return client.put(key, val, persistent)
}

// DelLbRule deletes a loadbalancer rule
func (client *Client) DelLbRule(id string) error {
	key := fmt.Sprintf("/bobbyd/lbrules/%v", id)
	return client.del(key)
}

// DelMwRule deletes a middleware rule
func (client *Client) DelMwRule(id string) error {
	key := fmt.Sprintf("/bobbyd/mwrules/%v", id)
	return client.del(key)
}

// DelHostConfig deletes a host config
func (client *Client) DelHostConfig(loadbalancer, hostID string) error {
	key := fmt.Sprintf("/bobbyd/loadbalancer/%v/hosts/%v", loadbalancer, hostID)
	return client.del(key)
}

// DelCertConfig deletes a cert config
func (client *Client) DelCertConfig(id string) error {
	key := fmt.Sprintf("/bobbyd/certs/%v", id)
	return client.del(key)
}

func (client *Client) put(key, val string, persistent bool) error {
	if persistent {
		_, err := client.v3.Put(client.ctx, key, val)
		return err
	}
	_, err := client.v3.Put(client.ctx, key, val, clientv3.WithLease(client.leaseID))
	return err
}

func (client *Client) del(key string) error {
	resp, err := client.v3.Delete(client.ctx, key)
	if resp.Deleted == 0 {
		return errors.New("entity not found")
	}
	return err
}
