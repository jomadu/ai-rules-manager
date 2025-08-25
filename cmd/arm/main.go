package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/jomadu/ai-rules-manager/pkg/cache"
	"github.com/jomadu/ai-rules-manager/pkg/config"
	"github.com/jomadu/ai-rules-manager/pkg/filesystem"
	"github.com/jomadu/ai-rules-manager/pkg/provider"
)

func main() {
	// Initialize core infrastructure components
	configManager := config.NewFileConfigManager(".armrc.json", "arm.json", "arm.lock")
	fsManager := filesystem.NewAtomicFileSystemManager(".")
	cacheImpl := cache.NewFileCache(os.ExpandEnv("$HOME/.arm/cache"))

	// Root command
	rootCmd := &cobra.Command{
		Use:   "arm",
		Short: "AI Rules Manager",
		Long:  "A package manager for AI coding assistant rulesets",
	}

	// Install command - implements Installation Flow from components.md
	installCmd := &cobra.Command{
		Use:   "install [ruleset]",
		Short: "Install a ruleset",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// TODO: implement Installation Flow:
			// 1. Configuration Manager parses registry and ruleset info
			infraConfig, err := configManager.LoadInfraConfig()
			if err != nil {
				return err
			}

			// 2. Create provider based on registry type and create components
			for _, regConfig := range infraConfig.Registries {
				var registryProvider provider.RegistryProvider

				switch regConfig.Type {
				case "git":
					registryProvider = provider.NewGitRegistryProvider()
				// TODO: add other registry types
				default:
					return fmt.Errorf("unsupported registry type: %s", regConfig.Type)
				}

				// Create ALL registry-specific components using provider
				registry, _ := registryProvider.CreateRegistry(regConfig)
				versionResolver, _ := registryProvider.CreateVersionResolver()
				contentResolver, _ := registryProvider.CreateContentResolver()
				keyGenerator, _ := registryProvider.CreateCacheKeyGenerator()

				_ = registry
				_ = versionResolver
				_ = contentResolver
				_ = keyGenerator
				break
			}

			// 3-8. Continue with rest of Installation Flow...
			_ = fsManager
			_ = cacheImpl
			return nil
		},
	}

	installCmd.Flags().StringSlice("patterns", nil, "File patterns to install")
	rootCmd.AddCommand(installCmd)

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
