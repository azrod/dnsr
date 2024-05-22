/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"os"
	"time"

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
		if err := config.ReadConfig(cfgFile); err != nil {
			log.Error().Err(err).Msg("Error reading the configuration file")
			return
		}

		var (
			done   = make(chan bool)
			ticker *time.Ticker
		)

		go config.WatchConfigFile(cfgFile, done)

		if config.Cfg.ExternalUpstreamsInternal > 0 {
			ticker = time.NewTicker(time.Duration(config.Cfg.ExternalUpstreamsInternal) * time.Minute)
		} else {
			ticker = time.NewTicker(5 * time.Minute)
		}
		go func() {
			for {
				select {
				case <-done:
					return
				case <-ticker.C:
					config.LoadExternalUpstreams()
				}
			}
		}()

		srv := &dns.Server{Addr: config.Cfg.Server.GetListenAddress(), Net: "udp"}
		srv.Handler = &server.DNSHandler{}
		log.Info().Msgf("Server listening on %s", config.Cfg.Server.GetListenAddress())

		if err := srv.ListenAndServe(); err != nil {
			log.Error().Err(err).Msg("Error starting the server")
			// done chan is used to signal the watcher to stop
			done <- true
			os.Exit(1)
		}
		// nolint: errcheck
		defer srv.Shutdown()
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
