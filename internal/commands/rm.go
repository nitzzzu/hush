package commands

import (
"fmt"
"github.com/spf13/cobra"
)

func newRmCmd() *cobra.Command {
return &cobra.Command{
Use:     "rm <NAME>",
Aliases: []string{"delete", "remove"},
Short:   "Delete a secret",
Args:    cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
if err := v.DeleteSecret(args[0]); err != nil {
return err
}
if !rootFlags.Quiet {
fmt.Printf("Secret %q deleted\n", args[0])
}
return nil
},
}
}
