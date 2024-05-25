/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"

	"github.com/azrod/dnsr/internal/config"
	"github.com/spf13/cobra"
)

// checkConfigCmd represents the checkConfig command
var checkConfigCmd = &cobra.Command{
	Use:   "checkConfig",
	Short: "Check the configuration file",
	Long:  `Check the configuration file to ensure it is valid and can be used by the application.`,
	Run: func(cmd *cobra.Command, args []string) {
		if err := config.ReadConfig(cfgFile); err != nil {
			fmt.Println("Error reading the configuration file:", err)
			return
		}

		fmt.Printf("Configuration file located at %s is valid\n", cfgFile)
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
