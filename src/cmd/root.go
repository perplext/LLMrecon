package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "LLMrecon",
	Short: "LLMreconing Tool for security assessment of Large Language Models",
	Long: `LLMreconing Tool is a security assessment platform designed to evaluate
the robustness and safety of Large Language Models (LLMs) by identifying
vulnerabilities and undesirable behaviors.

This tool provides a template-based, automated approach to red teaming LLMs,
inspired by Nuclei's approach to web application security. It enables security
professionals, AI developers, and organizations to ensure their LLMs are secure,
compliant, and resilient against adversarial attacks.`,

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}

func init() {
	cobra.OnInitialize(initConfig)

	// Here you will define your flags and configuration settings.
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.LLMrecon.yaml)")

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := os.UserHomeDir()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".LLMrecon" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigType("yaml")
		viper.SetConfigName(".LLMrecon")
	}

	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
