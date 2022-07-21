package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/bamboohr"
	"github.com/pomerium/datasource/internal/server"
)

type bambooCmd struct {
	BambooAPIKey    string `validate:"required"`
	BambooSubdomain string `validate:"required,hostname,excludesrune=."`
	BambooTimeZone  string `validate:"required"`
	Address         string `validate:"required"`
	Debug           bool
	cobra.Command   `validate:"-"`
	zerolog.Logger  `validate:"-"`
}

func bambooCommand(log zerolog.Logger) *cobra.Command {
	cmd := &bambooCmd{
		Command: cobra.Command{
			Use:   "bamboohr",
			Short: "run BambooHR connector",
		},
		Logger: log,
	}
	cmd.RunE = cmd.exec

	cmd.setupFlags()
	return &cmd.Command
}

func (cmd *bambooCmd) setupFlags() {
	flags := cmd.Flags()
	flags.StringVar(&cmd.BambooSubdomain, "bamboohr-subdomain", "", "BambooHR subdomain, i.e. if your instance is corp.bamboohr.com, then subdomain is corp")
	flags.StringVar(&cmd.BambooAPIKey, "bamboohr-api-key", "", "api key, see https://documentation.bamboohr.com/docs#section-authentication")
	flags.StringVar(&cmd.BambooTimeZone, "bamboohr-time-zone", "UTC", "BambooHR global time zone, see Settings > Account > General Settings > Time Zone ")
	flags.BoolVar(&cmd.Debug, "debug", false, "turns debug mode on that would dump requests and responses")
	flags.StringVar(&cmd.Address, "address", ":8080", "tcp address to listen to")
}

func (cmd *bambooCmd) exec(c *cobra.Command, _ []string) error {
	if err := validator.New().Struct(cmd); err != nil {
		return err
	}

	srv, err := cmd.newServer()
	if err != nil {
		return fmt.Errorf("prep server: %w", err)
	}

	log := zerolog.New(os.Stdout)
	log.Info().Msg("ready")

	return server.RunHTTPServer(c.Context(), cmd.Address, srv)
}

func (cmd *bambooCmd) newServer() (http.Handler, error) {
	auth := bamboohr.Auth{
		APIKey:    cmd.BambooAPIKey,
		Subdomain: cmd.BambooSubdomain,
	}

	location, err := time.LoadLocation(cmd.BambooTimeZone)
	if err != nil {
		return nil, fmt.Errorf("time zone %s: %w", cmd.BambooTimeZone, err)
	}

	emplReq := bamboohr.EmployeeRequest{
		Auth:     auth,
		Location: location,
	}
	client := http.DefaultClient
	if cmd.Debug {
		client = server.NewDebugClient(http.DefaultClient, cmd.Logger)
	}
	srv := bamboohr.NewServer(emplReq, client, cmd.Logger)
	return srv, nil
}
