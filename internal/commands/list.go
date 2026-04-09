package commands

import (
"encoding/json"
"fmt"
"strings"
"github.com/spf13/cobra"
)

func newListCmd() *cobra.Command {
return &cobra.Command{
Use:     "list",
Aliases: []string{"ls"},
Short:   "List all secrets",
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
secrets, err := v.ListSecrets(rootFlags.Tags)
if err != nil {
return err
}
if rootFlags.JSON {
type entry struct {
Name string   `json:"name"`
Tags []string `json:"tags"`
}
var out []entry
for _, s := range secrets {
out = append(out, entry{Name: s.Name, Tags: s.Tags})
}
return json.NewEncoder(cmd.OutOrStdout()).Encode(out)
}
if len(secrets) == 0 {
if !rootFlags.Quiet {
fmt.Println("No secrets found.")
}
return nil
}
for _, s := range secrets {
if len(s.Tags) > 0 {
fmt.Printf("  %s  [%s]\n", s.Name, strings.Join(s.Tags, ", "))
} else {
fmt.Printf("  %s\n", s.Name)
}
}
return nil
},
}
}
