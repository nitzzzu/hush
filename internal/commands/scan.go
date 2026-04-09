package commands

import (
"encoding/json"
"fmt"
"os"
"github.com/nitzzzu/hush/internal/scanner"
"github.com/spf13/cobra"
)

func newScanCmd() *cobra.Command {
var staged bool
var dir string
cmd := &cobra.Command{
Use:   "scan",
Short: "Scan for secret leaks in files",
RunE: func(cmd *cobra.Command, args []string) error {
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
secrets, err := v.AllSecrets()
if err != nil {
return fmt.Errorf("failed to load secrets: %w", err)
}
secretMap := make(map[string]string)
for _, s := range secrets {
secretMap[s.Name] = s.Value
}
sc := scanner.New(secretMap)
var results []scanner.Result
if dir != "" {
results, err = sc.ScanDir(dir)
} else if staged {
results, err = sc.ScanGitStaged()
} else {
results, err = sc.ScanGitTracked()
}
if err != nil {
return fmt.Errorf("scan failed: %w", err)
}
if len(results) == 0 {
if !rootFlags.Quiet {
fmt.Println("No secret leaks found.")
}
return nil
}
if rootFlags.JSON {
return json.NewEncoder(os.Stdout).Encode(results)
}
fmt.Fprintf(os.Stderr, "Found %d potential secret leak(s):\n", len(results))
for _, r := range results {
fmt.Fprintf(os.Stderr, "  %s:%d — secret %q\n", r.File, r.Line, r.Secret)
}
os.Exit(1)
return nil
},
}
cmd.Flags().BoolVar(&staged, "staged", false, "scan only staged files")
cmd.Flags().StringVarP(&dir, "dir", "d", "", "scan a specific directory")
return cmd
}
