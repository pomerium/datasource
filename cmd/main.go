package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/pomerium/datasource/internal"
)

func main() {
	log := makeLogger()

	rootCmd := &cobra.Command{
		Use:     "pomerium-datasource",
		Version: internal.FullVersion(),
	}
	rootCmd.AddCommand(bambooCommand(log), zenefitsCommand(log))
	if err := rootCmd.ExecuteContext(signalContext(log)); err != nil {
		log.Fatal().Err(err).Msg("exit")
	}
}

func makeLogger() zerolog.Logger {
	return zerolog.New(zerolog.NewConsoleWriter())
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