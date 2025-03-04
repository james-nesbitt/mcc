package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/Mirantis/mcc/pkg/analytics"
	"github.com/Mirantis/mcc/pkg/config"
	log "github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

// NewClientConfigCommand creates a download bundle command to be called via the CLI.
func NewClientConfigCommand() *cli.Command {
	return &cli.Command{
		Name:    "client-config",
		Aliases: []string{"download-bundle"},
		Usage:   "Get cluster client configuration",
		Flags: append(GlobalFlags, []cli.Flag{
			configFlag,
			confirmFlag,
			redactFlag,
			&cli.StringFlag{
				Name:   "username",
				Usage:  "Username",
				Hidden: true,
			},
			&cli.StringFlag{
				Name:   "Password",
				Usage:  "Password",
				Hidden: true,
			},
		}...),
		Before: actions(initLogger, startUpgradeCheck, initAnalytics, checkLicense, initExec, deprecateUserPass),
		After:  actions(closeAnalytics, upgradeCheckResult),
		Action: func(ctx *cli.Context) error {
			product, err := config.ProductFromFile(ctx.String("config"))
			if err != nil {
				return fmt.Errorf("failed to read product configuration: %w", err)
			}

			if err := product.ClientConfig(); err != nil {
				analytics.TrackEvent("Client configuration download Failed", nil)
				return fmt.Errorf("failed to download client configuration: %w", err)
			}

			analytics.TrackEvent("Client configuration download Completed", nil)

			return nil
		},
	}
}

func deprecateUserPass(ctx *cli.Context) error {
	if strings.Contains(strings.Join(os.Args, " "), "download-bundle") {
		log.Warn()
		log.Warn("[DEPRECATED] The 'download-bundle' subcommand is now 'client-config'.")
		log.Warn()
	}
	if ctx.String("username") != "" || ctx.String("password") != "" {
		log.Warn("[DEPRECATED] The --username and --password flags are ignored and are now read from the configuration file")
		log.Warn()
	}
	return nil
}
