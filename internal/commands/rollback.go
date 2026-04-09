package commands

import (
"fmt"
"strconv"
"github.com/spf13/cobra"
)

func newRollbackCmd() *cobra.Command {
return &cobra.Command{
Use:   "rollback <NAME> <VERSION>",
Short: "Rollback a secret to a previous version",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
version, err := strconv.Atoi(args[1])
if err != nil {
return fmt.Errorf("invalid version %q: must be an integer", args[1])
}
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
if err := v.Rollback(args[0], version); err != nil {
return err
}
if !rootFlags.Quiet {
fmt.Printf("Secret %q rolled back to version %d\n", args[0], version)
}
return nil
},
}
}
