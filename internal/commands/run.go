package commands

import (
"fmt"
"os"
"os/exec"
"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
return &cobra.Command{
Use:   "run -- <COMMAND> [ARGS...]",
Short: "Run a command with secrets injected as environment variables",
Args:  cobra.MinimumNArgs(1),
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
environ := os.Environ()
for _, s := range secrets {
environ = append(environ, fmt.Sprintf("%s=%s", s.Name, s.Value))
}
c := exec.Command(args[0], args[1:]...)
c.Env = environ
c.Stdin = os.Stdin
c.Stdout = os.Stdout
c.Stderr = os.Stderr
if err := c.Run(); err != nil {
if exitErr, ok := err.(*exec.ExitError); ok {
os.Exit(exitErr.ExitCode())
}
return err
}
return nil
},
}
}
