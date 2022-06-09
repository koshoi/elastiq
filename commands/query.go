package commands

import (
	"fmt"
	"io"
	"os"
	"strings"

	"elastiq/client"
	"elastiq/config"
	q "elastiq/query"
	"elastiq/source/elasticsearch"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func getQueryCommand(name, usage string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   name,
		Short: usage,
	}

	cf := addCommonFlags(cmd)

	strs := []string{}
	ascurl := false
	recursive := ""
	raw := false
	limit := 0
	timeRange := ""
	orderBy := ""

	pflags := cmd.PersistentFlags()
	pflags.StringArrayVarP(&strs, "filter", "f", []string{}, "filter values like key=value")
	pflags.BoolVarP(&ascurl, "curl", "", false, "output elasticsearch request as curl")
	pflags.BoolVarP(&raw, "raw", "r", false, "toggle raw ouput from elasticsearch (disables output post processing)")
	pflags.IntVarP(&limit, "limit", "l", 50, "specify limit for output records (specifying more than 10000 will apply paging)")
	pflags.StringVarP(&recursive, "recursive", "R", "", "toggle recursive decoding")
	pflags.StringVarP(&timeRange, "time", "t", "", "specify time filter as a/b (equivalent to -f '@timestamp intime a b'")
	pflags.StringVarP(&orderBy, "orderby", "O", "", "specify records order (defaults to descending by @timestamp)")

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

		timeSettings := q.TimeFilterSettings{
			TimeZone:   tz,
			TimeFormat: e.GetTimeFormat(cf.tf),
		}

		var client client.Client
		switch e.Source {
		case config.SourceElasticSearch:
			client = elasticsearch.NewClient(cfg)
		case config.SourceDataDog:
			client = elasticsearch.NewClient(cfg)
		}

		query := &q.Query{
			Filters: []*q.Filter{},
		}

		if orderBy == "" {
			orderBy = e.Order
		}

		if orderBy != "" {
			o, err := q.GetOrder(orderBy)
			if err != nil {
				return fmt.Errorf("failed to parse order: %w", err)
			}
			query.Order = o
		}

		if timeRange != "" {
			t := strings.Split(timeRange, "/")
			if len(t) > 2 {
				return fmt.Errorf("to many delimiters in timerange='%s'", timeRange)
			}

			if len(t) == 1 {
				t = append(t, "now")
			}

			strs = append(strs, fmt.Sprintf("@timestamp intime '%s' '%s'", t[0], t[1]))
		}

		for _, v := range strs {
			filter, err := q.ParseFilter(v, timeSettings, cfg.Aliases)
			if err != nil {
				return fmt.Errorf("failed to parse filter='%s': %w", v, err)
			}

			query.Filters = append(query.Filters, filter)
		}

		query.Index = cf.index
		query.Output = cf.output
		query.Limit = e.GetLimit(limit)

		options := q.Options{
			Debug:     cf.debug,
			FromStdin: cf.stdin,
			Raw:       raw,
			AsCurl:    ascurl,
			Recursive: nil,
		}

		cmd.Flags().Visit(func(f *pflag.Flag) {
			if f.Name == "recursive" {
				rlist := strings.Split(recursive, ",")
				options.Recursive = &rlist
			}
		})

		result, err := client.Query(cmd.Context(), e, query, options)
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
