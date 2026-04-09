package commands

import (
"fmt"
"github.com/nitzzzu/hush/internal/vault"
"github.com/spf13/cobra"
)

func newEnvCmd() *cobra.Command {
cmd := &cobra.Command{
Use:   "env",
Short: "Manage vault environments",
}
cmd.AddCommand(&cobra.Command{
Use:   "list",
Short: "List available environments",
RunE: func(cmd *cobra.Command, args []string) error {
envs, err := vault.ListEnvs(rootFlags.Global)
if err != nil {
return err
}
if len(envs) == 0 {
fmt.Println("No environments found.")
return nil
}
for _, e := range envs {
fmt.Printf("  %s\n", e)
}
return nil
},
})
return cmd
}
