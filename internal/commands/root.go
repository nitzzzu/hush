package commands

import (
"os"
"github.com/spf13/cobra"
)

type GlobalFlags struct {
Global bool
Env    string
Tags   []string
JSON   bool
Quiet  bool
}

var rootFlags GlobalFlags

func NewRootCmd() *cobra.Command {
root := &cobra.Command{
Use:          "hush",
Short:        "An AI-native secrets manager",
Long:         `hush stores secrets encrypted at rest and injects them into subprocesses.`,
SilenceUsage: true,
}
root.PersistentFlags().BoolVarP(&rootFlags.Global, "global", "g", false, "use global vault (~/.hush/)")
root.PersistentFlags().StringVar(&rootFlags.Env, "env", "", "use specific environment")
root.PersistentFlags().StringArrayVar(&rootFlags.Tags, "tag", nil, "filter by tag (repeatable)")
root.PersistentFlags().BoolVar(&rootFlags.JSON, "json", false, "JSON output")
root.PersistentFlags().BoolVarP(&rootFlags.Quiet, "quiet", "q", false, "suppress output, use exit codes")
root.AddCommand(
newInitCmd(), newSetCmd(), newGetCmd(), newListCmd(), newRmCmd(),
newTagCmd(), newUntagCmd(), newHistoryCmd(), newRollbackCmd(),
newImportCmd(), newExportCmd(), newRunCmd(), newScanCmd(), newEnvCmd(),
)
return root
}

func resolveVaultPath() (string, error) {
global := rootFlags.Global
env := rootFlags.Env
if os.Getenv("HUSH_GLOBAL") == "1" {
global = true
}
if e := os.Getenv("HUSH_ENV"); e != "" && env == "" {
env = e
}
return vaultVaultPath(global, env)
}
