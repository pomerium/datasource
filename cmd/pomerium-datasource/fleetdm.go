package main

import (
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/pomerium/datasource/internal/fleetdm"
	"github.com/pomerium/datasource/internal/server"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
)

type fleetDMCmd struct {
	APIToken    string `validate:"required"`
	APIURL      string `validate:"required,url"`
	Address     string `validate:"required"`
	CertQueryID uint   `validate:"required"`

	cobra.Command  `validate:"-"`
	zerolog.Logger `validate:"-"`
}

func fleetDMCommand(log zerolog.Logger) *cobra.Command {
	cmd := &fleetDMCmd{
		Command: cobra.Command{
			Use:   "fleetdm",
			Short: "run FleetDM connector",
		},
		Logger: log,
	}
	cmd.RunE = cmd.exec

	cmd.setupFlags()
	return &cmd.Command
}

func (cmd *fleetDMCmd) setupFlags() {
	flags := cmd.Flags()
	flags.StringVar(&cmd.APIToken, "api-token", "", "FleetDM API token")
	flags.StringVar(&cmd.APIURL, "api-url", "", "FleetDM API URL")
	flags.UintVar(&cmd.CertQueryID, "cert-query-id", 0, "FleetDM certificate query ID")
	flags.StringVar(&cmd.Address, "address", ":8080", "tcp address to listen to")
}

func (cmd *fleetDMCmd) exec(c *cobra.Command, _ []string) error {
	if err := validator.New().Struct(cmd); err != nil {
		return err
	}

	srv, err := cmd.newServer()
	if err != nil {
		return err
	}

	return server.RunHTTPServer(c.Context(), cmd.Address, srv)
}

func (cmd *fleetDMCmd) newServer() (http.Handler, error) {
	srv, err := fleetdm.NewServer(
		fleetdm.WithAPIToken(cmd.APIToken),
		fleetdm.WithAPIURL(cmd.APIURL),
		fleetdm.WithCertificateQueryID(cmd.CertQueryID),
	)
	if err != nil {
		return nil, err
	}

	return srv, nil
}
