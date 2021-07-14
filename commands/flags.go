package commands

import (
	"os"
	"path"

	"github.com/spf13/cobra"
)

type commonFlags struct {
	config string
	env    string
	output string
	index  string
	debug  bool
	tz     string
	tf     string
}

func addCommonFlags(cmd *cobra.Command) *commonFlags {
	cf := commonFlags{}

	configPath := ""
	home, err := os.UserHomeDir()
	if err == nil {
		configPath = path.Join(home, ".config", "elastiq", "config.toml")
	}

	flags := cmd.PersistentFlags()
	flags.StringVarP(&cf.config, "config", "c", configPath, "set path to config")
	flags.StringVarP(&cf.env, "env", "e", "", "specify env to use for quering ElasticSearch")
	flags.StringVarP(&cf.output, "output", "o", "", "specify output")
	flags.StringVarP(&cf.index, "index", "i", "", "specify index for querying")
	flags.BoolVarP(&cf.debug, "debug", "d", false, "enable debug")
	flags.StringVarP(&cf.tz, "timezone", "t", "", "specify timezone to use to compose time filters")
	flags.StringVarP(&cf.tf, "timeformat", "T", "", "specify golang timeformat to use to compose time filters")

	return &cf
}
