package phase

import (
	"strings"

	"github.com/Mirantis/mcc/pkg/config"

	log "github.com/sirupsen/logrus"
)

// GatherUcpFacts collects facts about possibly existing UCP setup
type GatherUcpFacts struct{}

// Title for the phase
func (p *GatherUcpFacts) Title() string {
	return "Gather UCP facts"
}

// Run collects the facts from swarm leader
func (p *GatherUcpFacts) Run(conf *config.ClusterConfig) error {
	swarmLeader := conf.Controllers()[0]
	exists, existingVersion, err := ucpExists(swarmLeader)

	if err != nil {
		return NewError("Failed to check existing UCP version")
	}

	conf.Ucp.Metadata = &config.UcpMetadata{
		Installed:        exists,
		InstalledVersion: existingVersion,
	}

	log.Debugf("Found UCP facts: %+v", conf.Ucp.Metadata)

	return nil
}

// checks whether UCP is already running. If it is also returns the current version.
func ucpExists(swarmLeader *config.Host) (bool, string, error) {
	output, err := swarmLeader.ExecWithOutput(`sudo docker inspect --format '{{ index .Config.Labels "com.docker.ucp.version"}}' ucp-proxy`)
	if err != nil {
		// We need to check the output to check if the container does not exist
		if strings.Contains(output, "No such object") {
			return false, "", nil
		}
		return false, "", err
	}
	return true, output, nil
}
