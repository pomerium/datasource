package main

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/server"
	"github.com/pomerium/datasource/internal/wellknownips"
)

var wellKnownIPsArgs struct {
	address   string
	ip2asnURL string
}

var wellKnownIPsCmd = &cobra.Command{
	Use:   "well-known-ips",
	Short: "runs the well known ips server",
	Run: func(cmd *cobra.Command, _ []string) {
		log.Info().
			Str("address", wellKnownIPsArgs.address).
			Str("ip2asn-url", wellKnownIPsArgs.ip2asnURL).
			Msg("starting well-known-ips http server")
		srv := wellknownips.NewServer(wellknownips.WithIP2ASNURL(wellKnownIPsArgs.ip2asnURL))
		err := server.RunHTTPServer(cmd.Context(), wellKnownIPsArgs.address, srv)
		if err != nil {
			log.Fatal().Err(err).Send()
		}
	},
}

func init() {
	wellKnownIPsCmd.Flags().StringVar(&wellKnownIPsArgs.address, "address", ":8080",
		"the tcp address to listen on")
	wellKnownIPsCmd.Flags().StringVar(&wellKnownIPsArgs.ip2asnURL, "ip2asn-url", wellknownips.DefaultIP2ASNURL,
		"the URL for the ip2asn database")
}
