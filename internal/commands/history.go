package commands

import (
"encoding/json"
"fmt"
"github.com/spf13/cobra"
)

func newHistoryCmd() *cobra.Command {
return &cobra.Command{
Use:   "history <NAME>",
Short: "Show version history for a secret",
Args:  cobra.ExactArgs(1),
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
history, err := v.GetHistory(args[0])
if err != nil {
return err
}
if rootFlags.JSON {
return json.NewEncoder(cmd.OutOrStdout()).Encode(history)
}
if len(history) == 0 {
fmt.Println("No history found.")
return nil
}
for _, h := range history {
fmt.Printf("  v%d  %s\n", h.Version, h.CreatedAt.Format("2006-01-02 15:04:05"))
}
return nil
},
}
}
