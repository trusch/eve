package etcd

import (
	"log"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/trusch/eve/config"
	lbRule "github.com/trusch/eve/loadbalancer/rule"
	mwRule "github.com/trusch/eve/middleware/rule"
)

func (client *Client) watchLbRules() {
	rch := client.v3.Watch(client.ctx, "/eve/lbrules", clientv3.WithPrefix())
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
				id := string(ev.Kv.Key[len("/eve/lbrules/"):])
				client.feedDeleteLbRuleToChannel(&lbRule.Rule{ID: id})
			}
		}
	}
}

func (client *Client) watchMwRules() {
	rch := client.v3.Watch(client.ctx, "/eve/mwrules", clientv3.WithPrefix())
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
				id := string(ev.Kv.Key[len("/eve/mwrules/"):])
				client.feedDeleteMwRuleToChannel(&mwRule.Rule{ID: id})
			}
		}
	}
}

func (client *Client) watchHosts() {
	rch := client.v3.Watch(client.ctx, "/eve/loadbalancer", clientv3.WithPrefix())
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
	rch := client.v3.Watch(client.ctx, "/eve/certs", clientv3.WithPrefix())
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
				cfg := &config.CertConfig{}
				cfg.ID = string(ev.Kv.Key[len("/eve/certs/"):])
				client.feedDeleteCertToChannel(cfg)
			}
		}
	}
}
