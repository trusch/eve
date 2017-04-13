package etcd

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/trusch/eve/config"
	lbRule "github.com/trusch/eve/loadbalancer/rule"
	mwRule "github.com/trusch/eve/middleware/rule"
)

// Client is a etcd backed config.Stream
type Client struct {
	v3         *clientv3.Client
	leaseID    clientv3.LeaseID
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
	resp, err := cli.Grant(ctx, 5)
	if err != nil {
		log.Fatal(err)
	}
	client.leaseID = resp.ID
	if _, err := cli.KeepAlive(ctx, resp.ID); err != nil {
		log.Fatal(err)
	}
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
