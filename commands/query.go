package commands

import (
	"fmt"
	"io"
	"os"

	"github.com/koshoi/elastiq/config"
	"github.com/koshoi/elastiq/elasticsearch"
	"github.com/spf13/cobra"
)

func getQueryCommand(name, usage string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: usage,
	}

	cf := addCommonFlags(cmd)

	strs := []string{}

	pflags := cmd.PersistentFlags()
	pflags.StringArrayVarP(&strs, "filter", "f", []string{}, "filter values like key=value")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig(cf.config)
		if err != nil {
			return err
		}

		client := elasticsearch.NewClient(cfg)
		query := &elasticsearch.Query{
			Filters: []*elasticsearch.Filter{},
		}

		for _, v := range strs {
			filter, err := elasticsearch.ParseFilter(v)
			if err != nil {
				return fmt.Errorf("failed to parse filter='%s': %w", v, err)
			}

			query.Filters = append(query.Filters, filter)
		}

		query.Index = cf.index
		query.Output = cf.output

		result, err := client.Query(cmd.Context(), cf.env, query, elasticsearch.Options{
			Debug: cf.debug,
		})
		if err != nil {
			return fmt.Errorf("failed to run query: %w", err)
		}

		io.Copy(os.Stdout, result)
		return nil
	}

	return cmd
}

func AddQueryCommand(rootCmd *cobra.Command) {
	rootCmd.AddCommand(
		getQueryCommand("query", "query elasticsearch"),
		getQueryCommand("q", "alias for query command"),
	)
}
