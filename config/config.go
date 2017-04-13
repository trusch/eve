package config

import (
	lbRule "github.com/trusch/bobbyd/loadbalancer/rule"
	mwRule "github.com/trusch/bobbyd/middleware/rule"
)

// Stream is the interface used by the application to get config value updates
type Stream interface {
	GetChannel() chan *Action
}

// A Action is a change of configuration
type Action struct {
	Type       ActionType
	LbRule     *lbRule.Rule
	MwRule     *mwRule.Rule
	HostConfig *HostConfig
	CertConfig *CertConfig
}

// HostConfig represents the registration of a host
type HostConfig struct {
	ID           string
	Loadbalancer string
	URL          string
}

// CertConfig represents a certificate
type CertConfig struct {
	ID      string
	CertPem string
	KeyPem  string
}

// ActionType is the type of a Action
type ActionType int

const (
	// UpsertLbRule represents the request to upsert a loadbalancer rule
	UpsertLbRule ActionType = iota
	// UpsertMwRule represents the request to upsert a middleware rule
	UpsertMwRule
	// UpsertCert represents the request to upsert a certificate
	UpsertCert
	// UpsertHost represents the request to upsert a host on a loadbalancer
	UpsertHost
	// DeleteLbRule represents the request to delete a loadbalancer rule
	DeleteLbRule
	// DeleteMwRule represents the request to delete a middleware rule
	DeleteMwRule
	// DeleteCert represents the request to delete a certificate
	DeleteCert
	// DeleteHost represents the request to delete a host from a loadbalancer
	DeleteHost
)

// Encrypt seals the cert config with a password
func (cfg *CertConfig) Encrypt(password string) error {
	cfg.CertPem = encrypt(cfg.CertPem, password)
	cfg.KeyPem = encrypt(cfg.KeyPem, password)
	return nil
}

// Decrypt decrypts a sealed cert config
func (cfg *CertConfig) Decrypt(password string) error {
	cfg.CertPem = decrypt(cfg.CertPem, password)
	cfg.KeyPem = decrypt(cfg.KeyPem, password)
	return nil
}
