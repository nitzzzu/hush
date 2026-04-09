package commands

import (
"bufio"
"fmt"
"io"
"os"
"strings"
"github.com/spf13/cobra"
"golang.org/x/term"
)

func newSetCmd() *cobra.Command {
var tags []string
var fromStdin bool
cmd := &cobra.Command{
Use:   "set <NAME> [VALUE]",
Short: "Set a secret value",
Args:  cobra.RangeArgs(1, 2),
RunE: func(cmd *cobra.Command, args []string) error {
name := args[0]
var value string
if len(args) == 2 {
value = args[1]
} else if fromStdin {
data, err := io.ReadAll(os.Stdin)
if err != nil {
return err
}
value = strings.TrimRight(string(data), "\r\n")
} else {
fmt.Fprintf(os.Stderr, "Enter value for %s: ", name)
if term.IsTerminal(int(os.Stdin.Fd())) {
data, err := term.ReadPassword(int(os.Stdin.Fd()))
if err != nil {
return err
}
fmt.Fprintln(os.Stderr)
value = string(data)
} else {
sc := bufio.NewScanner(os.Stdin)
if sc.Scan() {
value = sc.Text()
}
}
}
v, err := openVault()
if err != nil {
return err
}
defer v.Close()
allTags := append(rootFlags.Tags, tags...)
if err := v.SetSecret(name, value, allTags); err != nil {
return err
}
if !rootFlags.Quiet {
fmt.Printf("Secret %q stored successfully\n", name)
}
return nil
},
}
cmd.Flags().StringArrayVarP(&tags, "tag", "t", nil, "add tag (repeatable)")
cmd.Flags().BoolVar(&fromStdin, "stdin", false, "read value from stdin")
return cmd
}
