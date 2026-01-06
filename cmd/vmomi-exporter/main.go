package main

import (
	"context"
	"fmt"
	"log"
	"sort"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/9506hqwy/vmomi-exporter/pkg/config"
	"github.com/9506hqwy/vmomi-exporter/pkg/exporter"
	"github.com/9506hqwy/vmomi-exporter/pkg/flag"
	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

var version = "<version>"
var commit = "<commit>"

var rootCmd = &cobra.Command{
	Use:     "vmomi-exporter",
	Short:   "VMOMI Exporter",
	Long:    "VMOMI Exporter",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx = fromArgument(ctx)

		if err := exporter.Run(ctx); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	},
}

var configCmd = &cobra.Command{
	Use:     "config",
	Short:   "VMOMI Exporter Config",
	Long:    "VMOMI Exporter Config",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx = fromArgument(ctx)

		cfg, err := config.GetConfig(ctx)
		if err != nil {
			log.Fatalf("GetConfig error: %v", err)
		}

		conf, err := config.EncodeConfig(cfg)
		if err != nil {
			log.Fatalf("EncodeConfig error: %v", err)
		}

		fmt.Print(conf)
	},
}

var counterCmd = &cobra.Command{
	Use:     "counter",
	Short:   "VMOMI Exporter Counter",
	Long:    "VMOMI Exporter Counter",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx = fromArgument(ctx)

		counters, err := vmomi.GetCounterInfo(ctx)
		if err != nil {
			log.Fatalf("GetCounterInfo error: %v", err)
		}

		sort.Slice(*counters, func(a, b int) bool {
			return (*counters)[a].Id < (*counters)[b].Id
		})

		cnts := []config.Counter{}
		for _, c := range *counters {
			cnt := config.Counter{
				Group:  c.Group,
				Name:   c.Name,
				Rollup: c.Rollup,
			}
			cnts = append(cnts, cnt)
		}

		conf, err := config.EncodeCounters(&cnts)
		if err != nil {
			log.Fatalf("EncodeCounters error: %v", err)
		}

		fmt.Print(conf)
	},
}

var instanceCmd = &cobra.Command{
	Use:     "instance",
	Short:   "VMOMI Exporter Instance",
	Long:    "VMOMI Exporter Instance",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		ctx := context.Background()
		ctx = fromArgument(ctx)

		instances, err := vmomi.GetInstanceInfo(ctx, vmomi.ManagedEntityTypeValues())
		if err != nil {
			log.Fatalf("GetInstanceInfo error: %v", err)
		}

		sort.Slice(*instances, func(a, b int) bool {
			return (*instances)[a].EntityType < (*instances)[b].EntityType
		})

		inses := []config.Instance{}
		for _, c := range *instances {
			cnt := config.Instance{
				EntityType: c.EntityType,
				EntityId:   c.EntityId,
				EntityName: c.EntityName,
				Instance:   c.Instance,
				CounterId:  c.CounterId,
			}
			inses = append(inses, cnt)
		}

		conf, err := config.EncodeInstances(&inses)
		if err != nil {
			log.Fatalf("EncodeInstances error: %v", err)
		}

		fmt.Print(conf)
	},
}

var intervalCmd = &cobra.Command{
	Use:     "interval",
	Short:   "VMOMI Exporter Interval",
	Long:    "VMOMI Exporter Interval",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		entityTypeStr, err := cmd.Flags().GetString("entity-type")
		if err != nil || entityTypeStr == "" {
			log.Fatalf("Get entity-type error: %v", err)
		}

		entityId, err := cmd.Flags().GetString("entity-id")
		if err != nil || entityId == "" {
			log.Fatalf("Get entity-id error: %v", err)
		}

		entityType := vmomi.ManagedEntityType(entityTypeStr)

		ctx := context.Background()
		ctx = fromArgument(ctx)

		entities, err := vmomi.GetEntity(ctx, []vmomi.ManagedEntityType{entityType})
		if err != nil {
			log.Fatalf("GetInstanceInfo error: %v", err)
		}

		var entity *vmomi.Entity
		for _, e := range *entities {
			if e.Id == entityId {
				entity = &e
			}
		}

		if entity == nil {
			log.Fatalf("Entity not found: %s %s", entityType, entityId)
		}

		intervals, err := vmomi.GetIntervalInfo(ctx, *entity)
		if err != nil {
			log.Fatalf("GetIntervalInfo error: %v", err)
		}

		for _, interval := range intervals {
			fmt.Printf("%d\n", interval)
		}
	},
}

var perfCmd = &cobra.Command{
	Use:     "perf",
	Short:   "VMOMI Exporter Performance",
	Long:    "VMOMI Exporter Performance",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, args []string) {
		entityTypeStr, err := cmd.Flags().GetString("entity-type")
		if err != nil || entityTypeStr == "" {
			log.Fatalf("Get entity-type error: %v", err)
		}

		entityId, err := cmd.Flags().GetString("entity-id")
		if err != nil || entityId == "" {
			log.Fatalf("Get entity-id error: %v", err)
		}

		entityType := vmomi.ManagedEntityType(entityTypeStr)

		counterId, err := cmd.Flags().GetInt32("counter")
		if err != nil || counterId == 0 {
			log.Fatalf("Get counter error: %v", err)
		}

		interval, err := cmd.Flags().GetInt32("interval")
		if err != nil || interval == 0 {
			log.Fatalf("Get interval error: %v", err)
		}

		entity := &vmomi.Entity{
			Id:   entityId,
			Type: entityType,
		}

		ctx := context.Background()
		ctx = fromArgument(ctx)

		metrics, err := vmomi.QueryEntity(ctx, *entity, interval, counterId)
		if err != nil {
			log.Fatalf("QueryEntity error: %v", err)
		}

		for _, metric := range metrics {
			fmt.Printf(
				"%v\tinstance=%v\tinterval=%v\tcounter=%v\tvalue=%v\n",
				metric.Timestamp,
				metric.Instance,
				metric.Interval,
				metric.Counter.Id,
				metric.Value)
		}
	},
}

func fromArgument(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, flag.TargetUrlKey{}, viper.GetString("target_url"))
	ctx = context.WithValue(ctx, flag.TargetUserKey{}, viper.GetString("target_user"))
	ctx = context.WithValue(ctx, flag.TargetPasswordKey{}, viper.GetString("target_password"))
	ctx = context.WithValue(ctx, flag.TargetNoVerifySSLKey{}, viper.GetBool("target_no_verify_ssl"))
	ctx = context.WithValue(ctx, flag.ExporterConfigKey{}, viper.GetString("config"))
	ctx = context.WithValue(ctx, flag.ExporterUrlKey{}, viper.GetString("url"))
	return ctx
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("url", "https://127.0.0.1/sdk", "vSphere server URL.")
	rootCmd.PersistentFlags().String("user", "", "vSphere server username.")
	rootCmd.PersistentFlags().String("password", "", "vSphere server password.")
	rootCmd.PersistentFlags().Bool("no-verify-ssl", false, "Skip SSL verification.")
	rootCmd.PersistentFlags().String("config", "", "Config file path.")
	rootCmd.Flags().String("exporter", "127.0.0.1:9247", "Exporter URL.")

	intervalCmd.Flags().String("entity-type", "", "Entity type.")
	intervalCmd.Flags().String("entity-id", "", "Entity ID.")

	perfCmd.Flags().String("entity-type", "", "Entity type.")
	perfCmd.Flags().String("entity-id", "", "Entity ID.")
	perfCmd.Flags().Int32("counter", 0, "Counter ID.")
	perfCmd.Flags().Int32("interval", 0, "Interval.")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(counterCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(intervalCmd)
	rootCmd.AddCommand(perfCmd)

	viper.BindPFlag("target_url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("target_user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("target_password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("target_no_verify_ssl", rootCmd.PersistentFlags().Lookup("no-verify-ssl"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("url", rootCmd.Flags().Lookup("exporter"))
}

func initConfig() {
	viper.SetEnvPrefix("vmomi_exporter")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
