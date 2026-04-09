package main

import (
"fmt"
"os"
"github.com/nitzzzu/hush/internal/commands"
)

var version = "dev"

func main() {
root := commands.NewRootCmd(version)
if err := root.Execute(); err != nil {
fmt.Fprintln(os.Stderr, err)
os.Exit(1)
}
}
