package main

import (
	"context"
	"fmt"
	"io"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"gocloud.dev/gcerrors"

	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/pkg/blob"
	"github.com/pomerium/datasource/pkg/directory"
	"github.com/pomerium/datasource/pkg/directory/auth0"
	"github.com/pomerium/datasource/pkg/directory/azure"
	"github.com/pomerium/datasource/pkg/directory/cognito"
	"github.com/pomerium/datasource/pkg/directory/github"
	"github.com/pomerium/datasource/pkg/directory/gitlab"
	"github.com/pomerium/datasource/pkg/directory/google"
	"github.com/pomerium/datasource/pkg/directory/keycloak"
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
				accessKeyID := optionalStringFlag(flags, "access-key-id", "access key id")
				region := optionalStringFlag(flags, "region", "aws region")
				secretAccessKey := optionalStringFlag(flags, "secret-access-key", "secret access key")
				sessionToken := optionalStringFlag(flags, "session-token", "session token")
				userPoolID := optionalStringFlag(flags, "user-pool-id", "user pool id")
				return func() directory.Provider {
					return cognito.New(
						cognito.WithAccessKeyID(*accessKeyID),
						cognito.WithRegion(*region),
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
		directorySubCommand(logger, "keycloak",
			func(flags *pflag.FlagSet) func() directory.Provider {
				clientID := requiredStringFlag(flags, "client-id", "client id")
				clientSecret := requiredStringFlag(flags, "client-secret", "client secret")
				realm := requiredStringFlag(flags, "realm", "realm name")
				url := requiredStringFlag(flags, "url", "url")
				return func() directory.Provider {
					return keycloak.New(
						keycloak.WithClientID(*clientID),
						keycloak.WithClientSecret(*clientSecret),
						keycloak.WithRealm(*realm),
						keycloak.WithLogger(logger),
						keycloak.WithURL(*url),
					)
				}
			}),
		directorySubCommand(logger, "okta",
			func(flags *pflag.FlagSet) func() directory.Provider {
				apiKey := requiredStringFlag(flags, "api-key", "api key")
				url := requiredStringFlag(flags, "url", "url")
				return func() directory.Provider {
					return okta.New(
						okta.WithAPIKey(*apiKey),
						okta.WithLogger(logger),
						okta.WithURL(*url),
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
	cmd.Run = func(cmd *cobra.Command, _ []string) {
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		provider := newProvider()
		err := server.RunHTTPServer(cmd.Context(), addr, directory.NewHandler(provider))
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
	}

	cmd.AddCommand(directoryUploadSubCommand(logger, setupFlags))

	return cmd
}

func directoryUploadSubCommand(
	logger zerolog.Logger,
	setupFlags func(flags *pflag.FlagSet) func() directory.Provider,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "upload",
		Short: "upload directory data to blob storage",
	}
	debug := false
	cmd.Flags().BoolVar(&debug, "debug", false, "debug mode")
	destination := requiredStringFlag(cmd.Flags(), "destination", "blob url to upload files to")
	newProvider := setupFlags(cmd.Flags())
	cmd.Run = func(cmd *cobra.Command, _ []string) {
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}
		provider := newProvider()

		err := uploadDirectoryBundleToBlob(cmd.Context(), provider, *destination)
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

func uploadDirectoryBundleToBlob(ctx context.Context, provider directory.Provider, urlstr string) error {
	if provider, ok := provider.(directory.PersistentProvider); ok {
		err := downloadDirectoryStateFromBlob(ctx, provider, urlstr)
		if err != nil {
			return err
		}
	}

	groups, users, err := provider.GetDirectory(ctx)
	if err != nil {
		return fmt.Errorf("error retrieving directory data: %w", err)
	}

	err = blob.UploadBundle(ctx, urlstr, map[string]any{
		directory.GroupRecordType: groups,
		directory.UserRecordType:  users,
	})
	if err != nil {
		return fmt.Errorf("error uploading directory data: %w", err)
	}

	if provider, ok := provider.(directory.PersistentProvider); ok {
		err := uploadDirectoryStateToBlob(ctx, provider, urlstr)
		if err != nil {
			return err
		}
	}

	return nil
}

func downloadDirectoryStateFromBlob(ctx context.Context, provider directory.PersistentProvider, urlstr string) error {
	err := blob.DownloadState(ctx, urlstr, func(src io.Reader) error {
		return provider.LoadDirectoryState(ctx, src)
	})
	if gcerrors.Code(err) == gcerrors.NotFound {
		return nil
	} else if err != nil {
		return fmt.Errorf("error downloading directory state from blob: %w", err)
	}

	return nil
}

func uploadDirectoryStateToBlob(ctx context.Context, provider directory.PersistentProvider, urlstr string) error {
	err := blob.UploadState(ctx, urlstr, func(dst io.Writer) error {
		return provider.SaveDirectoryState(ctx, dst)
	})
	if err != nil {
		return fmt.Errorf("error uploading directory state to blob: %w", err)
	}

	return nil
}
