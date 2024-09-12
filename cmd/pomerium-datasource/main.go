package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal/version"
)

func main() {
	logger := makeLogger()

	rootCmd := &cobra.Command{
		Use:     "pomerium-datasource",
		Version: version.FullVersion(),
	}
	rootCmd.AddCommand(
		bambooCommand(logger),
		directoryCommand(logger),
		zenefitsCommand(logger),
		ip2LocationCmd,
		wellKnownIPsCmd,
		fleetDMCommand(logger),
	)
	if err := rootCmd.ExecuteContext(signalContext(logger)); err != nil {
		logger.Fatal().Err(err).Msg("exit")
	}
}

func makeLogger() zerolog.Logger {
	logger := zerolog.New(zerolog.NewConsoleWriter())
	log.Logger = logger
	return logger
}

func signalContext(log zerolog.Logger) context.Context {
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		sig := <-sigs
		log.Error().Str("signal", sig.String()).Msg("caught signal, quitting...")
		cancel()
		time.Sleep(time.Second * 2)
		log.Error().Msg("did not shut down gracefully, exit")
		os.Exit(1)
	}()
	return ctx
}
