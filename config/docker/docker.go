package docker

import (
	"context"
	"fmt"
	"log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/trusch/bobbyd/config"
	"github.com/trusch/bobbyd/loadbalancer/rule"
)

// ConfigSource reads docker labels and updates lb rules and hosts accordingly
type ConfigSource struct {
	cli    *client.Client
	output chan *config.Action
}

// New creates a new ConfigSource
func New() (*ConfigSource, error) {
	c, err := client.NewEnvClient()
	if err != nil {
		return nil, err
	}
	res := &ConfigSource{
		cli:    c,
		output: make(chan *config.Action, 32),
	}
	go res.backend()
	return res, nil
}

// GetChannel returns the action channel
func (src *ConfigSource) GetChannel() chan *config.Action {
	return src.output
}

func (src *ConfigSource) backend() {
	containers, err := src.cli.ContainerList(context.Background(), types.ContainerListOptions{})
	if err != nil {
		log.Print(err)
	}
	for _, container := range containers {
		if host := checkForBobbydHostLabel(container.Labels); host != "" {
			src.handleStart(container.ID, host)
		}
	}

	events, _ := src.cli.Events(context.Background(), types.EventsOptions{})
	for event := range events {
		if event.Action == "start" {
			if host := checkForBobbydHostLabel(event.Actor.Attributes); host != "" {
				src.handleStart(event.Actor.ID, host)
			}
		} else if event.Action == "die" {
			if host := checkForBobbydHostLabel(event.Actor.Attributes); host != "" {
				src.handleStop(event.Actor.ID, host)
			}
		}
	}

}

func checkForBobbydHostLabel(labels map[string]string) string {
	for key, val := range labels {
		if key == "bobbyd.host" {
			return val
		}
	}
	return ""
}

func (src *ConfigSource) getIP(id string) (string, error) {
	info, err := src.cli.ContainerInspect(context.Background(), id)
	if err != nil {
		return "", err
	}
	return info.NetworkSettings.IPAddress, nil
}

func (src *ConfigSource) handleStart(id string, host string) {
	ip, err := src.getIP(id)
	if err != nil {
		log.Print(err)
		return
	}
	src.output <- &config.Action{
		Type: config.UpsertHost,
		HostConfig: &config.HostConfig{
			ID:           id,
			Loadbalancer: host,
			URL:          "http://" + ip,
		},
	}
	src.output <- &config.Action{
		Type: config.UpsertLbRule,
		LbRule: &rule.Rule{
			ID:     id,
			Target: host,
			Route:  fmt.Sprintf(`Host("%v")`, host),
		},
	}
}

func (src *ConfigSource) handleStop(id string, host string) {
	src.output <- &config.Action{
		Type: config.DeleteHost,
		HostConfig: &config.HostConfig{
			ID: id,
		},
	}
}
