package phase

import (
	"github.com/Mirantis/mcc/pkg/api"
	retry "github.com/avast/retry-go"
	log "github.com/sirupsen/logrus"
)

// Connect connects to each of the hosts
type Connect struct {
	Analytics
	BasicPhase
}

// Title for the phase
func (p *Connect) Title() string {
	return "Open Remote Connection"
}

// Run connects to all the hosts in parallel
func (p *Connect) Run() error {
	return runParallelOnHosts(p.config.Spec.Hosts, p.config, p.connectHost)
}

func (p *Connect) connectHost(host *api.Host, c *api.ClusterConfig) error {
	proto := "SSH"

	if host.Localhost {
		proto = "Local"
	} else if host.WinRM != nil {
		proto = "WinRM"
	}

	err := retry.Do(
		func() error {
			log.Infof("%s: opening %s connection", host.Address, proto)
			err := host.Connect()
			if err != nil {
				log.Errorf("%s: failed to connect -> %s", host.Address, err.Error())
			}
			return err
		},
		retry.Attempts(6),
	)
	if err != nil {
		log.Errorf("%s: failed to open connection", host.Address)
		return err
	}

	log.Printf("%s: %s connection opened", host.Address, proto)
	return p.testConnection(host)
}

func (p *Connect) testConnection(h *api.Host) error {
	log.Infof("%s: testing connection", h.Address)

	if err := h.ExecCmd("echo", "", false, false); err != nil {
		return err
	}

	return nil
}
