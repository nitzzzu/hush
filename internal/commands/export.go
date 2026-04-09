package commands

import (
"fmt"
"os"
"strings"
"github.com/spf13/cobra"
)

func newExportCmd() *cobra.Command {
var file string
var format string
cmd := &cobra.Command{
Use:   "export",
Short: "Export secrets to a .env file or shell format",
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
var out *os.File
if file != "" {
out, err = os.Create(file)
if err != nil {
return fmt.Errorf("failed to create file: %w", err)
}
defer out.Close()
} else {
out = os.Stdout
}
for _, s := range secrets {
switch strings.ToLower(format) {
case "shell", "sh":
fmt.Fprintf(out, "export %s=%q\n", s.Name, s.Value)
default:
fmt.Fprintf(out, "%s=%s\n", s.Name, s.Value)
}
}
return nil
},
}
cmd.Flags().StringVarP(&file, "file", "f", "", "output file (default: stdout)")
cmd.Flags().StringVar(&format, "format", "dotenv", "output format: dotenv or shell")
return cmd
}
