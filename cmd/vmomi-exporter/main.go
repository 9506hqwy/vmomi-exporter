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

const rootFolerName = ""

//revive:disable:deep-exit

var rootCmd = &cobra.Command{
	Use:     "vmomi-exporter",
	Short:   "VMOMI Exporter",
	Long:    "VMOMI Exporter",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(_ *cobra.Command, _ []string) {
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
	Run: func(_ *cobra.Command, _ []string) {
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

		_, err = fmt.Print(conf)
		if err != nil {
			log.Fatalf("Print error: %v", err)
		}
	},
}

var counterCmd = &cobra.Command{
	Use:     "counter",
	Short:   "VMOMI Exporter Counter",
	Long:    "VMOMI Exporter Counter",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(_ *cobra.Command, _ []string) {
		ctx := context.Background()
		ctx = fromArgument(ctx)

		counters, err := vmomi.GetCounterInfo(ctx)
		if err != nil {
			log.Fatalf("GetCounterInfo error: %v", err)
		}

		sort.Slice(*counters, func(a, b int) bool {
			return (*counters)[a].ID < (*counters)[b].ID
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

		_, err = fmt.Print(conf)
		if err != nil {
			log.Fatalf("Print error: %v", err)
		}
	},
}

var entityCmd = &cobra.Command{
	Use:     "entity",
	Short:   "VMOMI Exporter Entity",
	Long:    "VMOMI Exporter Entity",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, _ []string) {
		ctx := context.Background()
		ctx = fromArgument(ctx)

		root, err := getRootEntity(cmd)
		if err != nil {
			log.Fatalf("Get arguments: %v", err)
		}

		entities, err := exporter.ToEntityFromRoot(ctx, []config.Root{*root})
		if err != nil {
			log.Fatalf("ToEntityFromRoot error: %v", err)
		}

		if entities == nil {
			entities, err = vmomi.GetEntityFromRoot(ctx, vmomi.ManagedEntityTypeValues())
			if err != nil {
				log.Fatalf("GetEntityFromRoot error: %v", err)
			}
		} else {
			entities, err = vmomi.GetEntity(ctx, *entities, vmomi.ManagedEntityTypeValues(), true)
			if err != nil {
				log.Fatalf("GetEntity error: %v", err)
			}
		}

		sort.Slice(*entities, func(a, b int) bool {
			return (*entities)[a].Type < (*entities)[b].Type
		})

		roots := []config.Root{}
		for _, e := range *entities {
			r := config.Root{
				Type: e.Type,
				Name: e.Name,
			}
			roots = append(roots, r)
		}

		conf, err := config.EncodeRoots(&roots)
		if err != nil {
			log.Fatalf("EncodeRoots error: %v", err)
		}

		_, err = fmt.Print(conf)
		if err != nil {
			log.Fatalf("Print error: %v", err)
		}
	},
}

var instanceCmd = &cobra.Command{
	Use:     "instance",
	Short:   "VMOMI Exporter Instance",
	Long:    "VMOMI Exporter Instance",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(_ *cobra.Command, _ []string) {
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
				EntityID:   c.EntityID,
				EntityName: c.EntityName,
				Instance:   c.Instance,
				CounterID:  c.CounterID,
			}
			inses = append(inses, cnt)
		}

		conf, err := config.EncodeInstances(&inses)
		if err != nil {
			log.Fatalf("EncodeInstances error: %v", err)
		}

		_, err = fmt.Print(conf)
		if err != nil {
			log.Fatalf("Print error: %v", err)
		}
	},
}

var intervalCmd = &cobra.Command{
	Use:     "interval",
	Short:   "VMOMI Exporter Interval",
	Long:    "VMOMI Exporter Interval",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, _ []string) {
		entityTypeStr, err := cmd.Flags().GetString("entity-type")
		if err != nil || entityTypeStr == "" {
			log.Fatalf("Get entity-type error: %v", err)
		}

		entityID, err := cmd.Flags().GetString("entity-id")
		if err != nil || entityID == "" {
			log.Fatalf("Get entity-id error: %v", err)
		}

		entityType := vmomi.ManagedEntityType(entityTypeStr)

		ctx := context.Background()
		ctx = fromArgument(ctx)

		entities, err := vmomi.GetEntityFromRoot(ctx, []vmomi.ManagedEntityType{entityType})
		if err != nil {
			log.Fatalf("GetInstanceInfo error: %v", err)
		}

		var entity *vmomi.Entity
		for _, e := range *entities {
			if e.ID == entityID {
				entity = &e
			}
		}

		if entity == nil {
			log.Fatalf("Entity not found: %s %s", entityType, entityID)
		}

		intervals, err := vmomi.GetIntervalInfo(ctx, *entity)
		if err != nil {
			log.Fatalf("GetIntervalInfo error: %v", err)
		}

		for _, interval := range intervals {
			_, err = fmt.Printf("%d (Current: %v)\n", interval.ID, interval.Current)
			if err != nil {
				log.Fatalf("Print error: %v", err)
			}
		}
	},
}

var perfCmd = &cobra.Command{
	Use:     "perf",
	Short:   "VMOMI Exporter Performance",
	Long:    "VMOMI Exporter Performance",
	Version: fmt.Sprintf("%s\nCommit: %s", version, commit),
	Run: func(cmd *cobra.Command, _ []string) {
		entityTypeStr, err := cmd.Flags().GetString("entity-type")
		if err != nil || entityTypeStr == "" {
			log.Fatalf("Get entity-type error: %v", err)
		}

		entityID, err := cmd.Flags().GetString("entity-id")
		if err != nil || entityID == "" {
			log.Fatalf("Get entity-id error: %v", err)
		}

		entityType := vmomi.ManagedEntityType(entityTypeStr)

		counterID, err := cmd.Flags().GetInt32("counter")
		if err != nil || counterID == 0 {
			log.Fatalf("Get counter error: %v", err)
		}

		interval, err := cmd.Flags().GetInt32("interval")
		if err != nil || interval == 0 {
			log.Fatalf("Get interval error: %v", err)
		}

		entity := &vmomi.Entity{
			ID:   entityID,
			Type: entityType,
		}

		ctx := context.Background()
		ctx = fromArgument(ctx)

		metrics, err := vmomi.QueryEntity(ctx, *entity, interval, counterID)
		if err != nil {
			log.Fatalf("QueryEntity error: %v", err)
		}

		for _, metric := range metrics {
			_, err = fmt.Printf(
				"%v\tinstance=%v\tinterval=%v\tcounter=%v\tvalue=%v\n",
				metric.Timestamp,
				metric.Instance,
				metric.Interval,
				metric.Counter.ID,
				metric.Value)
			if err != nil {
				log.Fatalf("Print error: %v", err)
			}
		}
	},
}

//revive:enable:deep-exit

func getRootEntity(cmd *cobra.Command) (*config.Root, error) {
	entityTypeStr, err := cmd.Flags().GetString("entity-type")
	if err != nil {
		return nil, err
	}

	entityName, err := cmd.Flags().GetString("entity-name")
	if err != nil {
		return nil, err
	}

	if entityTypeStr == "" || entityName == "" {
		root := config.Root{
			Type: vmomi.ManagedEntityTypeFolder,
			Name: rootFolerName,
		}
		return &root, nil
	}

	entityType := vmomi.ManagedEntityType(entityTypeStr)

	root := config.Root{
		Type: entityType,
		Name: entityName,
	}
	return &root, nil
}

//revive:disable:line-length-limit

func fromArgument(ctx context.Context) context.Context {
	ctx = context.WithValue(ctx, flag.TargetURLKey{}, viper.GetString("target_url"))
	ctx = context.WithValue(ctx, flag.TargetUserKey{}, viper.GetString("target_user"))
	ctx = context.WithValue(ctx, flag.TargetPasswordKey{}, viper.GetString("target_password"))
	ctx = context.WithValue(ctx, flag.TargetNoVerifySSLKey{}, viper.GetBool("target_no_verify_ssl"))
	ctx = context.WithValue(ctx, flag.TargetTimeoutKey{}, viper.GetInt("target_timeout"))
	ctx = context.WithValue(ctx, flag.ExporterConfigKey{}, viper.GetString("config"))
	ctx = context.WithValue(ctx, flag.ExporterURLKey{}, viper.GetString("url"))
	ctx = context.WithValue(ctx, flag.LogLevelKey{}, viper.GetString("log_level"))
	return ctx
}

//revive:enable:line-length-limit

//revive:disable:add-constant

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().String("url", "https://127.0.0.1/sdk", "vSphere server URL.")
	rootCmd.PersistentFlags().String("user", "", "vSphere server username.")
	rootCmd.PersistentFlags().String("password", "", "vSphere server password.")
	rootCmd.PersistentFlags().Bool("no-verify-ssl", false, "Skip SSL verification.")
	rootCmd.PersistentFlags().Int("timeout", 10, "API call timeout seconds.")
	rootCmd.PersistentFlags().String("config", "", "Config file path.")
	rootCmd.Flags().String("exporter", "127.0.0.1:9247", "Exporter URL.")
	rootCmd.Flags().String("log-level", "INFO", "Log level.")

	entityCmd.Flags().String("entity-type", "", "Entity type.")
	entityCmd.Flags().String("entity-name", "", "Entity Name.")

	intervalCmd.Flags().String("entity-type", "", "Entity type.")
	intervalCmd.Flags().String("entity-id", "", "Entity ID.")

	perfCmd.Flags().String("entity-type", "", "Entity type.")
	perfCmd.Flags().String("entity-id", "", "Entity ID.")
	perfCmd.Flags().Int32("counter", 0, "Counter ID.")
	perfCmd.Flags().Int32("interval", 0, "Interval.")

	rootCmd.AddCommand(configCmd)
	rootCmd.AddCommand(counterCmd)
	rootCmd.AddCommand(entityCmd)
	rootCmd.AddCommand(instanceCmd)
	rootCmd.AddCommand(intervalCmd)
	rootCmd.AddCommand(perfCmd)

	viper.BindPFlag("target_url", rootCmd.PersistentFlags().Lookup("url"))
	viper.BindPFlag("target_user", rootCmd.PersistentFlags().Lookup("user"))
	viper.BindPFlag("target_password", rootCmd.PersistentFlags().Lookup("password"))
	viper.BindPFlag("target_no_verify_ssl", rootCmd.PersistentFlags().Lookup("no-verify-ssl"))
	viper.BindPFlag("target_timeout", rootCmd.PersistentFlags().Lookup("timeout"))
	viper.BindPFlag("config", rootCmd.PersistentFlags().Lookup("config"))
	viper.BindPFlag("url", rootCmd.Flags().Lookup("exporter"))
	viper.BindPFlag("log_level", rootCmd.Flags().Lookup("log-level"))
}

//revive:enable:add-constant

func initConfig() {
	viper.SetEnvPrefix("vmomi_exporter")
	viper.AutomaticEnv()
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}
