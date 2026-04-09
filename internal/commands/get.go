package commands

import (
"encoding/json"
"fmt"
"github.com/spf13/cobra"
)

func newGetCmd() *cobra.Command {
var reveal bool
cmd := &cobra.Command{
Use:   "get <NAME>",
Short: "Get a secret value",
Args:  cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
s, err := v.GetSecret(args[0])
if err != nil {
return err
}
if rootFlags.JSON {
return json.NewEncoder(cmd.OutOrStdout()).Encode(map[string]interface{}{
"name": s.Name, "value": s.Value, "tags": s.Tags,
"created_at": s.CreatedAt, "updated_at": s.UpdatedAt,
})
}
if reveal {
fmt.Println(s.Value)
} else if rootFlags.Quiet {
fmt.Println(s.Value)
} else {
fmt.Printf("Name:  %s\nValue: %s\nTags:  %v\n", s.Name, "********", s.Tags)
}
return nil
},
}
cmd.Flags().BoolVarP(&reveal, "reveal", "r", false, "reveal the secret value in plaintext")
return cmd
}
