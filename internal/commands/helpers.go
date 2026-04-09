package commands

import "github.com/nitzzzu/hush/internal/vault"

func vaultVaultPath(global bool, env string) (string, error) {
return vault.VaultPath(global, env)
}

func openVault() (*vault.Vault, error) {
path, err := resolveVaultPath()
if err != nil {
return nil, err
}
return vault.Open(path)
}
