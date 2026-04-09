package main

import (
"fmt"
"os"
"github.com/nitzzzu/hush/internal/commands"
)

func main() {
root := commands.NewRootCmd()
if err := root.Execute(); err != nil {
fmt.Fprintln(os.Stderr, err)
os.Exit(1)
}
}
