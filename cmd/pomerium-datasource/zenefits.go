package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/internal/zenefits"
)

type zenefitsCmd struct {
	APIKey string `validate:"required"`

	Address string `validate:"required"`

	TimeZone string

	Debug bool

	cobra.Command  `validate:"-"`
	zerolog.Logger `validate:"-"`
}

func zenefitsCommand(log zerolog.Logger) *cobra.Command {
	cmd := &zenefitsCmd{
		Command: cobra.Command{
			Use:   "zenefits",
			Short: "run Zenefits connector",
		},
		Logger: log,
	}
	cmd.RunE = cmd.exec

	cmd.setupFlags()
	return &cmd.Command
}

func (cmd *zenefitsCmd) setupFlags() {
	flags := cmd.Flags()
	flags.StringVar(&cmd.APIKey, "zenefits-api-key", "", "Bearer API token https://developers.zenefits.com/v1.0/docs/auth")
	flags.StringVar(&cmd.Address, "address", "localhost:8080", "tcp address to listen on")
	flags.StringVar(&cmd.TimeZone, "time-zone", "UTC", "timezone, required for vacation data")
	flags.BoolVar(&cmd.Debug, "debug", false, "turns debug mode on that would dump requests and responses")
}

func (cmd *zenefitsCmd) exec(c *cobra.Command, _ []string) error {
	if err := validator.New().Struct(cmd); err != nil {
		return err
	}

	srv, err := cmd.newServer()
	if err != nil {
		return fmt.Errorf("prep server: %w", err)
	}

	log := zerolog.New(os.Stdout)
	log.Info().Str("address", cmd.Address).Msg("ready")

	return server.RunHTTPServer(c.Context(), cmd.Address, srv)
}

func (cmd *zenefitsCmd) newServer() (http.Handler, error) {
	client := server.NewBearerTokenClient(http.DefaultClient, cmd.APIKey)
	if cmd.Debug {
		client = server.NewDebugClient(client, cmd.Logger)
	}

	location, err := time.LoadLocation(cmd.TimeZone)
	if err != nil {
		return nil, fmt.Errorf("time zone %s: %w", cmd.TimeZone, err)
	}

	srv := zenefits.NewServer(zenefits.PeopleRequest{}, client,
		zenefits.WithLogger(cmd.Logger),
		zenefits.WithRemoveOnVacation(location))
	return srv, nil
}
