/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/azrod/dnsr/internal/config"
)

// checkConfigCmd represents the checkConfig command.
var checkConfigCmd = &cobra.Command{
	Use:   "checkConfig",
	Short: "Check the configuration file",
	Long:  `Check the configuration file to ensure it is valid and can be used by the application.`,
	Run: func(_ *cobra.Command, _ []string) {
		if err := config.ReadConfig(cfgFile); err != nil {
			log.Error().Err(err).Msg("Error reading the configuration file")
			return
		}

		log.Info().Msg("Configuration file is valid")
	},
}

func init() {
	rootCmd.AddCommand(checkConfigCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// checkConfigCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// checkConfigCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
