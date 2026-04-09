package commands

import (
"fmt"
"github.com/spf13/cobra"
)

func newTagCmd() *cobra.Command {
return &cobra.Command{
Use:   "tag <NAME> <TAG>",
Short: "Add a tag to a secret",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
if err := v.AddTag(args[0], args[1]); err != nil {
return err
}
if !rootFlags.Quiet {
fmt.Printf("Tag %q added to %q\n", args[1], args[0])
}
return nil
},
}
}

func newUntagCmd() *cobra.Command {
return &cobra.Command{
Use:   "untag <NAME> <TAG>",
Short: "Remove a tag from a secret",
Args:  cobra.ExactArgs(2),
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
if err := v.RemoveTag(args[0], args[1]); err != nil {
return err
}
if !rootFlags.Quiet {
fmt.Printf("Tag %q removed from %q\n", args[1], args[0])
}
return nil
},
}
}
