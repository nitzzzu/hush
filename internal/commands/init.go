package commands

import (
"fmt"
"github.com/nitzzzu/hush/internal/vault"
"github.com/spf13/cobra"
)

func newInitCmd() *cobra.Command {
return &cobra.Command{
Use:   "init",
Short: "Initialize a new vault",
RunE: func(cmd *cobra.Command, args []string) error {
path, err := resolveVaultPath()
if err != nil {
return err
}
v, err := vault.Init(path)
if err != nil {
return fmt.Errorf("failed to initialize vault: %w", err)
}
v.Close()
if !rootFlags.Quiet {
fmt.Printf("Vault initialized at %s\n", path)
}
return nil
},
}
}
