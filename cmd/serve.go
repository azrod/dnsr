/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/azrod/dnsr/internal/cache"
	"github.com/azrod/dnsr/internal/config"
	"github.com/azrod/dnsr/internal/server"
	"github.com/miekg/dns"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

// serveCmd represents the serve command
var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {

		var (
			sigs   = make(chan os.Signal, 1)
			ticker *time.Ticker
			done   = make(chan bool, 1)
		)

		signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

		if err := config.ReadConfig(cfgFile); err != nil {
			log.Error().Err(err).Msg("Error reading the configuration file")
			return
		}

		config.WatchConfigFile(cfgFile, done)

		if config.Cfg.ExternalUpstreamsInterval > 0 {
			ticker = time.NewTicker(time.Duration(config.Cfg.ExternalUpstreamsInterval) * time.Minute)
		} else {
			ticker = time.NewTicker(5 * time.Minute)
		}
		go func() {
			for {
				select {
				case <-ticker.C:
					config.LoadExternalUpstreams()
				case <-done:
					return
				}
			}
		}()

		srv := &dns.Server{Addr: config.Cfg.Server.GetListenAddress(), Net: "udp"}
		srv.Handler = &server.DNSHandler{}

		if config.Cfg.Cache.Enabled {
			ca, err := cache.New()
			if err != nil {
				log.Error().Err(err).Msg("Error creating cache")
			}
			srv.Handler = &server.DNSHandler{Cache: ca}
		}

		log.Info().Msgf("Server listening on %s", config.Cfg.Server.GetListenAddress())
		go func() {
			if err := srv.ListenAndServe(); err != nil {
				log.Error().Err(err).Msg("Error starting the server")
				// send signal to sigs channel with error
				sigs <- syscall.SIGINT
			}
		}()

		<-sigs
		log.Info().Msg("Received signal to shut down the server")

		// send signal to done channel
		done <- true

		if err := srv.Shutdown(); err != nil {
			log.Error().Err(err).Msg("Error shutting down the server")
		}

		if config.Cfg.Cache.Enabled {
			if err := cache.PersistCache(srv.Handler.(*server.DNSHandler).Cache); err != nil {
				log.Error().Err(err).Msg("Error persisting cache before shutting down")
			}
		}

	},
}

func init() {
	rootCmd.AddCommand(serveCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// serveCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// serveCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
