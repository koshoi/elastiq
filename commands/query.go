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
	ascurl := false

	pflags := cmd.PersistentFlags()
	pflags.StringArrayVarP(&strs, "filter", "f", []string{}, "filter values like key=value")
	pflags.BoolVarP(&ascurl, "curl", "", false, "output elasticsearch request as curl")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		cfg, err := config.ReadConfig(cf.config)
		if err != nil {
			return err
		}

		e, err := cfg.GetEnv(cf.env)
		if err != nil {
			return err
		}

		tz, err := e.GetTimezone(cf.tz)
		if err != nil {
			return err
		}

		timeSettings := elasticsearch.TimeFilterSettings{
			TimeZone:   tz,
			TimeFormat: e.GetTimeFormat(cf.tf),
		}

		client := elasticsearch.NewClient(cfg)
		query := &elasticsearch.Query{
			Filters: []*elasticsearch.Filter{},
		}

		for _, v := range strs {
			filter, err := elasticsearch.ParseFilter(v, timeSettings)
			if err != nil {
				return fmt.Errorf("failed to parse filter='%s': %w", v, err)
			}

			query.Filters = append(query.Filters, filter)
		}

		query.Index = cf.index
		query.Output = cf.output

		result, err := client.Query(cmd.Context(), e, query, elasticsearch.Options{
			Debug:  cf.debug,
			AsCurl: ascurl,
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
