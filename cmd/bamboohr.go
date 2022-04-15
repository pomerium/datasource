package main

import (
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/bamboohr"
	"github.com/pomerium/datasource/internal/server"
)

type bambooCmd struct {
	BambooAPIKey             string   `validate:"required"`
	BambooSubdomain          string   `validate:"required"`
	BambooEmployeeFields     []string `validate:"required"`
	BambooEmployeeFieldRemap []string `validate:"required"`
	Address                  string   `validate:"required,tcp_addr"`
	AccessToken              string   `validate:"required"`
	cobra.Command            `validate:"-"`
	zerolog.Logger           `validate:"-"`
}

var (
	// DefaultBambooEmployeeFields
	DefaultBambooEmployeeFields = []string{"email"}
	// RestrictedBambooEmployeeFields prohibit certain sensitive fields from being exposed
	RestrictedBambooEmployeeFields = []string{"national_id", "sin", "ssn", "nin", "id"}
	// DefaultBambooEmployeeFieldsRemap
	DefaultBambooEmployeeFieldsRemap = []string{"email=id"}
)

func bambooCommand(log zerolog.Logger) *cobra.Command {
	cmd := &bambooCmd{
		Command: cobra.Command{
			Use:   "serve",
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
	flags.StringVar(&cmd.AccessToken, "access-token", "", "all requests must contain Token header matching this token")
}

func (cmd *bambooCmd) exec(c *cobra.Command, _ []string) error {
	if err := validator.New().Struct(cmd); err != nil {
		return err
	}

	if err := checkFieldsNotInList(cmd.BambooEmployeeFields, RestrictedBambooEmployeeFields); err != nil {
		return fmt.Errorf("field check: %w", err)
	}
	srv, err := cmd.newServer()
	if err != nil {
		return fmt.Errorf("prep server: %w", err)
	}

	l, err := net.Listen("tcp", cmd.Address)
	if err != nil {
		return fmt.Errorf("listen %s: %w", cmd.Address, err)
	}

	log := zerolog.New(os.Stdout)
	log.Info().Msg("ready")
	return http.Serve(l, srv)
}

func (cmd *bambooCmd) newServer() (http.Handler, error) {
	auth := bamboohr.Auth{
		APIKey:    cmd.BambooAPIKey,
		Subdomain: cmd.BambooSubdomain,
	}

	remap, err := keyRemap(cmd.BambooEmployeeFieldRemap)
	if err != nil {
		return nil, fmt.Errorf("field remap: %w", err)
	}

	if err := checkFieldsInList(cmd.BambooEmployeeFields, remap); err != nil {
		return nil, fmt.Errorf("remap fields ")
	}
	emplReq := bamboohr.EmployeeRequest{
		Auth:        auth,
		CurrentOnly: true,
		Fields:      cmd.BambooEmployeeFields,
		Remap:       remap,
	}
	srv := bamboohr.NewServer(emplReq, cmd.Logger)
	srv.Use(server.TokenMiddleware(cmd.AccessToken))
	return srv, nil
}

func keyRemap(m []string) (map[string]string, error) {
	dst := make(map[string]string, len(m))
	for _, r := range m {
		pair := strings.Split(r, "=")
		if len(pair) != 2 {
			return nil, fmt.Errorf("%s: expect key=newKey format", r)
		}
		key, newKey := pair[0], pair[1]
		if key == "" || newKey == "" {
			return nil, fmt.Errorf("%s: expect key=newKey format", r)
		}
		if _, there := dst[newKey]; there {
			return nil, fmt.Errorf("%s: key %s was already used", r, newKey)
		}
		dst[key] = newKey
	}
	return dst, nil
}

func checkFieldsInList(fields []string, remap map[string]string) error {
	fm := make(map[string]struct{}, len(fields))
	for _, fld := range fields {
		fm[fld] = struct{}{}
	}

	for fld := range remap {
		if _, there := fm[fld]; !there {
			return fmt.Errorf("field %s in field remap must be present in the field list", fld)
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
