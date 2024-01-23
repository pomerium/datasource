package main

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/pkg/directory"
	"github.com/pomerium/datasource/pkg/directory/auth0"
	"github.com/pomerium/datasource/pkg/directory/azure"
	"github.com/pomerium/datasource/pkg/directory/cognito"
	"github.com/pomerium/datasource/pkg/directory/github"
	"github.com/pomerium/datasource/pkg/directory/gitlab"
	"github.com/pomerium/datasource/pkg/directory/google"
	"github.com/pomerium/datasource/pkg/directory/okta"
	"github.com/pomerium/datasource/pkg/directory/onelogin"
	"github.com/pomerium/datasource/pkg/directory/ping"
)

func directoryCommand(logger zerolog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "directory",
		Short: "runs the directory server",
	}

	cmd.AddCommand(
		directorySubCommand(logger, "auth0",
			func(flags *pflag.FlagSet) func() directory.Provider {
				clientID := requiredStringFlag(flags, "client-id", "client id")
				clientSecret := requiredStringFlag(flags, "client-secret", "client secret")
				domain := requiredStringFlag(flags, "domain", "domain")
				return func() directory.Provider {
					return auth0.New(
						auth0.WithClientID(*clientID),
						auth0.WithClientSecret(*clientSecret),
						auth0.WithDomain(*domain),
						auth0.WithLogger(logger),
					)
				}
			}),
		directorySubCommand(logger, "azure",
			func(flags *pflag.FlagSet) func() directory.Provider {
				clientID := requiredStringFlag(flags, "client-id", "client id")
				clientSecret := requiredStringFlag(flags, "client-secret", "client secret")
				directoryID := requiredStringFlag(flags, "directory-id", "directory id")
				return func() directory.Provider {
					return azure.New(
						azure.WithClientID(*clientID),
						azure.WithClientSecret(*clientSecret),
						azure.WithDirectoryID(*directoryID),
						azure.WithLogger(logger),
					)
				}
			}),
		directorySubCommand(logger, "cognito",
			func(flags *pflag.FlagSet) func() directory.Provider {
				userPoolID := optionalStringFlag(flags, "user-pool-id", "user pool id")
				accessKeyID := optionalStringFlag(flags, "access-key-id", "access key id")
				secretAccessKey := optionalStringFlag(flags, "secret-access-key", "secret access key")
				sessionToken := optionalStringFlag(flags, "session-token", "session token")
				return func() directory.Provider {
					return cognito.New(
						cognito.WithAccessKeyID(*accessKeyID),
						cognito.WithSecretAccessKey(*secretAccessKey),
						cognito.WithSessionToken(*sessionToken),
						cognito.WithUserPoolID(*userPoolID),
					)
				}
			}),
		directorySubCommand(logger, "github",
			func(flags *pflag.FlagSet) func() directory.Provider {
				personalAccessToken := requiredStringFlag(flags, "personal-access-token", "personal access token")
				username := requiredStringFlag(flags, "username", "username")
				return func() directory.Provider {
					return github.New(
						github.WithLogger(logger),
						github.WithPersonalAccessToken(*personalAccessToken),
						github.WithUsername(*username),
					)
				}
			}),
		directorySubCommand(logger, "gitlab",
			func(flags *pflag.FlagSet) func() directory.Provider {
				privateToken := requiredStringFlag(flags, "private-token", "private token")
				return func() directory.Provider {
					return gitlab.New(
						gitlab.WithLogger(logger),
						gitlab.WithPrivateToken(*privateToken),
					)
				}
			}),
		directorySubCommand(logger, "google",
			func(flags *pflag.FlagSet) func() directory.Provider {
				impersonateUser := requiredStringFlag(flags, "impersonate-user", "impersonate user")
				jsonKey := optionalBytesFlag(flags, "json-key", "json key (base64)")
				jsonKeyFile := optionalStringFlag(flags, "json-key-file", "json key file")
				return func() directory.Provider {
					return google.New(
						google.WithImpersonateUser(*impersonateUser),
						google.WithJSONKey(*jsonKey),
						google.WithJSONKeyFile(*jsonKeyFile),
						google.WithLogger(logger),
					)
				}
			}),
		directorySubCommand(logger, "okta",
			func(flags *pflag.FlagSet) func() directory.Provider {
				apiKey := requiredStringFlag(flags, "api-key", "api key")
				return func() directory.Provider {
					return okta.New(
						okta.WithAPIKey(*apiKey),
						okta.WithLogger(logger),
					)
				}
			}),
		directorySubCommand(logger, "onelogin",
			func(flags *pflag.FlagSet) func() directory.Provider {
				clientID := requiredStringFlag(flags, "client-id", "client id")
				clientSecret := requiredStringFlag(flags, "client-secret", "client secret")
				return func() directory.Provider {
					return onelogin.New(
						onelogin.WithClientID(*clientID),
						onelogin.WithClientSecret(*clientSecret),
						onelogin.WithLogger(logger),
					)
				}
			}),
		directorySubCommand(logger, "ping",
			func(flags *pflag.FlagSet) func() directory.Provider {
				return func() directory.Provider {
					clientID := requiredStringFlag(flags, "client-id", "client id")
					clientSecret := requiredStringFlag(flags, "client-secret", "client secret")
					environmentID := requiredStringFlag(flags, "environment-id", "environment id")
					return ping.New(
						ping.WithClientID(*clientID),
						ping.WithClientSecret(*clientSecret),
						ping.WithEnvironmentID(*environmentID),
						ping.WithLogger(logger),
					)
				}
			}),
	)
	return cmd
}

func directorySubCommand(
	logger zerolog.Logger,
	provider string,
	setupFlags func(flags *pflag.FlagSet) func() directory.Provider,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   provider,
		Short: "runs the " + provider + " directory server",
	}
	addr := ":8080"
	debug := false
	cmd.Flags().StringVar(&addr, "address", ":8080", "tcp address to listen to")
	cmd.Flags().BoolVar(&debug, "debug", false, "debug mode")
	newProvider := setupFlags(cmd.Flags())
	cmd.Run = func(cmd *cobra.Command, args []string) {
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		provider := newProvider()
		err := server.RunHTTPServer(cmd.Context(), addr, directory.NewHandler(provider))
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
	}
	return cmd
}

func optionalBytesFlag(flags *pflag.FlagSet, name, usage string) *[]byte {
	ptr := new([]byte)
	flags.BytesBase64Var(ptr, name, nil, usage)
	return ptr
}

func optionalStringFlag(flags *pflag.FlagSet, name, usage string) *string {
	ptr := new(string)
	flags.StringVar(ptr, name, "", usage)
	return ptr
}

func requiredStringFlag(flags *pflag.FlagSet, name, usage string) *string {
	ptr := new(string)
	flags.StringVar(ptr, name, "", usage)
	if err := cobra.MarkFlagRequired(flags, name); err != nil {
		panic(err)
	}
	return ptr
}
