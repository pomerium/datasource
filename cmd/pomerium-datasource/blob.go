package main

import (
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/pkg/blob"
)

func blobCommand(logger zerolog.Logger) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "blob",
		Short: "runs the blob server",
	}
	addr := ":8080"
	debug := false
	source := requiredStringFlag(cmd.Flags(), "source", "blob url to serve files from")
	cmd.Flags().StringVar(&addr, "address", ":8080", "tcp address to listen to")
	cmd.Flags().BoolVar(&debug, "debug", false, "debug mode")
	cmd.Run = func(cmd *cobra.Command, _ []string) {
		if debug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		err := server.RunHTTPServer(cmd.Context(), addr, blob.NewHandler(*source))
		if err != nil {
			logger.Fatal().Err(err).Send()
		}
	}
	return cmd
}
