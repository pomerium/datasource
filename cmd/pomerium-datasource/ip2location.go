package main

import (
	"fmt"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/ip2location"
	"github.com/pomerium/datasource/internal/server"
)

var ip2LocationArgs struct {
	address string
	file    string
}

var ip2LocationCmd = &cobra.Command{
	Use:   "ip2location <file>",
	Short: "runs the IP2Location server",
	Args:  cobra.ExactArgs(1),
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 0 {
			return fmt.Errorf("file is required")
		}
		ip2LocationArgs.file = args[0]
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		log.Info().
			Str("address", ip2LocationArgs.address).
			Str("file", ip2LocationArgs.file).
			Msg("starting ip2location http server")
		srv := ip2location.NewServer(ip2location.WithFile(ip2LocationArgs.file))
		err := server.RunHTTPServer(cmd.Context(), ip2LocationArgs.address, srv)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	ip2LocationCmd.Flags().StringVar(&ip2LocationArgs.address, "address", ":8080",
		"the tcp address to listen on")
}
