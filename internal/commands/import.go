package commands

import (
"bufio"
"fmt"
"os"
"strings"
"github.com/spf13/cobra"
)

func newImportCmd() *cobra.Command {
var file string
var overwrite bool
cmd := &cobra.Command{
Use:   "import",
Short: "Import secrets from a .env file",
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
var reader *os.File
if file != "" {
reader, err = os.Open(file)
if err != nil {
return fmt.Errorf("failed to open file: %w", err)
}
defer reader.Close()
} else {
reader = os.Stdin
}
imported, skipped := 0, 0
sc := bufio.NewScanner(reader)
for sc.Scan() {
line := strings.TrimSpace(sc.Text())
if line == "" || strings.HasPrefix(line, "#") {
continue
}
idx := strings.IndexByte(line, '=')
if idx < 0 {
continue
}
key := strings.TrimSpace(line[:idx])
val := strings.TrimSpace(line[idx+1:])
if len(val) >= 2 {
if (val[0] == '"' && val[len(val)-1] == '"') || (val[0] == '\'' && val[len(val)-1] == '\'') {
val = val[1 : len(val)-1]
}
}
if !overwrite {
if _, err := v.GetSecret(key); err == nil {
skipped++
continue
}
}
if err := v.SetSecret(key, val, rootFlags.Tags); err != nil {
return fmt.Errorf("failed to set %q: %w", key, err)
}
imported++
}
if !rootFlags.Quiet {
fmt.Printf("Imported %d secrets, skipped %d\n", imported, skipped)
}
return sc.Err()
},
}
cmd.Flags().StringVarP(&file, "file", "f", "", "path to .env file (default: stdin)")
cmd.Flags().BoolVar(&overwrite, "overwrite", false, "overwrite existing secrets")
return cmd
}
