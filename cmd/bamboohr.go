package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/bamboohr"
	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/internal/util"
)

type bambooCmd struct {
	BambooAPIKey             string   `validate:"required"`
	BambooSubdomain          string   `validate:"required"`
	BambooEmployeeFields     []string `validate:"required"`
	BambooEmployeeFieldRemap []string `validate:"required"`
	Address                  string   `validate:"required"`
	BearerToken              string   `validate:"required"`
	Debug                    bool
	cobra.Command            `validate:"-"`
	zerolog.Logger           `validate:"-"`
}

var (
	// DefaultBambooEmployeeFields
	// note `id` is always returned anyway by the API
	DefaultBambooEmployeeFields = []string{"workEmail", "department", "firstName", "lastName"}
	// RestrictedBambooEmployeeFields prohibit certain sensitive fields from being exposed
	RestrictedBambooEmployeeFields = []string{"national_id", "sin", "ssn", "nin"}
	// DefaultBambooEmployeeFieldsRemap
	DefaultBambooEmployeeFieldsRemap = []string{"id=bamboohr_id", "workEmail=id"}
)

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
	flags.StringVar(&cmd.BambooSubdomain, "bamboohr-subdomain", "", "subdomain")
	flags.StringSliceVar(&cmd.BambooEmployeeFields, "bamboohr-employee-fields", DefaultBambooEmployeeFields, "employee fields")
	flags.StringSliceVar(&cmd.BambooEmployeeFieldRemap, "bamboohr-employee-field-remap", DefaultBambooEmployeeFieldsRemap, "list of key=newKey to rename keys")
	flags.StringVar(&cmd.BambooAPIKey, "bamboohr-api-key", "", "api key, see https://documentation.bamboohr.com/docs#section-authentication")
	flags.StringVar(&cmd.BearerToken, "bearer-token", "", "all requests must contain Authorization: Bearer header matching this token")
	flags.BoolVar(&cmd.Debug, "debug", false, "turns debug mode on that would dump requests and responses")
	flags.StringVar(&cmd.Address, "address", ":8080", "tcp address to listen to")
}

func (cmd *bambooCmd) exec(c *cobra.Command, _ []string) error {
	if err := validator.New().Struct(cmd); err != nil {
		return err
	}

	remap, err := cmd.getFieldsRemap()
	if err != nil {
		return fmt.Errorf("field remap: %w", err)
	}

	srv, err := cmd.newServer(remap)
	if err != nil {
		return fmt.Errorf("prep server: %w", err)
	}

	log := zerolog.New(os.Stdout)
	log.Info().Msg("ready")

	return server.RunHTTPServer(c.Context(), cmd.Address, srv)
}

func (cmd *bambooCmd) getFieldsRemap() ([]util.FieldRemap, error) {
	// `id`` is always returned by the API, thus we assume it is present
	cmd.BambooEmployeeFields = append(cmd.BambooEmployeeFields, "id")
	if err := checkFieldsNotInList(cmd.BambooEmployeeFields, RestrictedBambooEmployeeFields); err != nil {
		return nil, fmt.Errorf("field check: %w", err)
	}

	remap, err := util.NewRemapFromPairs(cmd.BambooEmployeeFieldRemap)
	if err != nil {
		return nil, fmt.Errorf("field remap: %w", err)
	}

	if err := checkFieldsInList(cmd.BambooEmployeeFields, remap); err != nil {
		return nil, fmt.Errorf("remap fields: %w", err)
	}

	return remap, nil
}

func (cmd *bambooCmd) newServer(remap []util.FieldRemap) (http.Handler, error) {
	auth := bamboohr.Auth{
		APIKey:    cmd.BambooAPIKey,
		Subdomain: cmd.BambooSubdomain,
	}

	emplReq := bamboohr.EmployeeRequest{
		Auth:        auth,
		CurrentOnly: true,
		Fields:      cmd.BambooEmployeeFields,
		Remap:       remap,
	}
	client := server.NewDebugClient(http.DefaultClient, cmd.Logger)
	srv := bamboohr.NewServer(emplReq, client, cmd.Logger)
	srv.Use(server.AuthorizationBearerMiddleware(cmd.BearerToken))
	return srv, nil
}

func checkFieldsInList(fields []string, remap []util.FieldRemap) error {
	fm := make(map[string]struct{}, len(fields))
	for _, fld := range fields {
		fm[fld] = struct{}{}
	}

	for _, fld := range remap {
		if _, there := fm[fld.From]; !there {
			return fmt.Errorf("field %s in field remap must be present in the field list", fld.From)
		}
	}
	return nil
}

func checkFieldsNotInList(fields []string, denyList []string) error {
	denySet := make(map[string]struct{}, len(denyList))
	for _, key := range denyList {
		denySet[key] = struct{}{}
	}

	for _, key := range fields {
		if _, there := denySet[key]; there {
			return fmt.Errorf("%s is restricted", key)
		}
	}

	return nil
}
