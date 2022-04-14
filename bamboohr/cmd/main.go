package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/pomerium/datasource/bamboohr/internal"
	"github.com/rs/zerolog"
)

func main() {
	log := makeLogger()
	log.Info().Str("version", internal.FullVersion()).Msg("starting")
	cmd := serveCommand(log)
	if err := cmd.ExecuteContext(signalContext(log)); err != nil {
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
