package podman

import (
	"context"
	"fmt"
	"log/slog"
	"proxy-manager/pkg/caddy"
	"strconv"

	"github.com/containers/podman/v4/pkg/bindings"
	"github.com/containers/podman/v4/pkg/bindings/containers"
	"github.com/containers/podman/v4/pkg/domain/entities"
)

type LabelManager struct {
	conn   context.Context
	domain string
}

func NewLabelManager(ctx context.Context, uri string, domain string) (*LabelManager, error) {
	conn, err := bindings.NewConnection(ctx, uri)
	if err != nil {
		return nil, err
	}

	return &LabelManager{conn: conn, domain: domain}, nil
}

func (l LabelManager) getContainers() ([]entities.ListContainer, error) {
	containersList, err := containers.List(l.conn, &containers.ListOptions{})
	if err != nil {
		return nil, err
	}

	return containersList, nil
}

func (l LabelManager) GetContainerProxies() ([]caddy.Proxy, error) {
	containers, err := l.getContainers()
	if err != nil {
		return nil, err
	}

	var proxies []caddy.Proxy
	for _, container := range containers {
		labels := container.Labels

		enabled, ok := labels["proxy-manager.enable"]
		if !ok || enabled != "true" {
			continue
		}

		name, ok := labels["proxy-manager.name"]
		if !ok {
			name = container.Names[0]
		}

		var port int
		portString, ok := labels["proxy-manager.port"]
		if !ok {
			if len(container.Ports) == 0 {
				slog.Warn("Container has no ports exposed", "container", name)
				continue
			} else {
				port = int(container.Ports[0].HostPort)
			}
		} else {
			port, err = strconv.Atoi(portString)
			if err != nil {
				slog.Warn("Invalid port number", "container", name, "port", portString)
				continue
			}
		}

		proxies = append(proxies, caddy.Proxy{
			Upstream: fmt.Sprintf("localhost:%d", port),
			Match:    fmt.Sprintf("%s.%s", name, l.domain),
		})
	}

	return proxies, nil
}
