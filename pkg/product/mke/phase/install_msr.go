package phase

import (
	"fmt"

	"github.com/Mirantis/mcc/pkg/msr"
	"github.com/Mirantis/mcc/pkg/phase"
	common "github.com/Mirantis/mcc/pkg/product/common/api"
	"github.com/Mirantis/mcc/pkg/product/mke/api"
	"github.com/alessio/shellescape"
	"github.com/k0sproject/rig/exec"
	log "github.com/sirupsen/logrus"
)

// InstallMSR is the phase implementation for running the actual MSR installer
// bootstrap.
type InstallMSR struct {
	phase.Analytics
	phase.CleanupDisabling
	MSRPhase

	leader *api.Host
}

// Title prints the phase title.
func (p *InstallMSR) Title() string {
	return "Install MSR components"
}

// ShouldRun should return true only when there is an installation to be performed.
func (p *InstallMSR) ShouldRun() bool {
	p.leader = p.Config.Spec.MSRLeader()
	return p.Config.Spec.ContainsMSR() && (p.leader.MSRMetadata == nil || !p.leader.MSRMetadata.Installed)
}

// Run the installer container.
func (p *InstallMSR) Run() error {
	h := p.leader

	if h.MSRMetadata == nil {
		h.MSRMetadata = &api.MSRMetadata{}
	}

	managers := p.Config.Spec.Managers()

	err := p.Config.Spec.CheckMKEHealthRemote(managers)
	if err != nil {
		return fmt.Errorf("%s: failed to health check mke, try to set `--ucp-url` installFlag and check connectivity: %w", h, err)
	}

	p.EventProperties = map[string]interface{}{
		"msr_version": p.Config.Spec.MSR.Version,
	}

	image := p.Config.Spec.MSR.GetBootstrapperImage()

	runFlags := common.Flags{"-i"}
	if !p.CleanupDisabled() {
		runFlags.Add("--rm")
	}

	if h.Configurer.SELinuxEnabled(h) {
		runFlags.Add("--security-opt label=disable")
	}
	installFlags := p.Config.Spec.MSR.InstallFlags
	redacts := []string{installFlags.GetValue("--ucp-username"), installFlags.GetValue("--ucp-password")}

	// Configure the mkeFlags from existing MKEConfig
	mkeFlags := msr.BuildMKEFlags(p.Config)
	// Conduct the install passing the --ucp-node flag for the host provided in
	// msrLeader.
	mkeFlags.AddOrReplace(fmt.Sprintf("--ucp-node %s", h.Metadata.Hostname))

	installFlags.Merge(mkeFlags)

	if p.Config.Spec.MSR.CACertData != "" {
		escaped := shellescape.Quote(p.Config.Spec.MSR.CACertData)
		installFlags.AddOrReplace(fmt.Sprintf("--dtr-ca %s", escaped))
		redacts = append(redacts, escaped)
	}

	if p.Config.Spec.MSR.CertData != "" {
		escaped := shellescape.Quote(p.Config.Spec.MSR.CertData)
		installFlags.AddOrReplace(fmt.Sprintf("--dtr-cert %s", escaped))
		redacts = append(redacts, escaped)
	}

	if p.Config.Spec.MSR.KeyData != "" {
		escaped := shellescape.Quote(p.Config.Spec.MSR.KeyData)
		installFlags.AddOrReplace(fmt.Sprintf("--dtr-key %s", escaped))
		redacts = append(redacts, escaped)
	}

	if p.Config.Spec.MKE.CACertData != "" {
		escaped := shellescape.Quote(p.Config.Spec.MKE.CACertData)
		installFlags.AddOrReplace(fmt.Sprintf("--ucp-ca %s", escaped))
		redacts = append(redacts, escaped)
	}

	if h.MSRMetadata.ReplicaID != "" {
		log.Infof("%s: installing MSR with replica id %s", h, h.MSRMetadata.ReplicaID)
		installFlags.AddOrReplace(fmt.Sprintf("--replica-id %s", h.MSRMetadata.ReplicaID))
	} else {
		log.Infof("%s: installing MSR version %s", h, p.Config.Spec.MSR.Version)
	}

	installCmd := h.Configurer.DockerCommandf("run %s %s install %s", runFlags.Join(), image, installFlags.Join())
	err = h.Exec(installCmd, exec.StreamOutput(), exec.RedactString(redacts...))
	if err != nil {
		return fmt.Errorf("%s: failed to run MSR installer: %w", h, err)
	}

	msrMeta, err := msr.CollectFacts(h)
	if err != nil {
		return fmt.Errorf("%s: failed to collect existing MSR details: %w", h, err)
	}
	h.MSRMetadata = msrMeta
	return nil
}

// CleanUp removes remnants of MSR after a failed installation.
func (p *InstallMSR) CleanUp() {
	log.Infof("Cleaning up for '%s'", p.Title())
	if err := msr.Destroy(p.leader, p.Config); err != nil {
		log.Warnf("Error while cleaning-up resources for '%s': %s", p.Title(), err.Error())
	}
}
