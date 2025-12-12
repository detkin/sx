package commands

import (
	"context"

	"github.com/sleuth-io/skills/internal/clients"
	"github.com/sleuth-io/skills/internal/logger"
)

// installAllClientHooks detects installed clients and installs hooks for each.
func installAllClientHooks(ctx context.Context, out *outputHelper) {
	log := logger.Get()
	registry := clients.NewRegistry()
	installedClients := registry.DetectInstalled()

	for _, client := range installedClients {
		if err := client.InstallHooks(ctx); err != nil {
			out.printfErr("Warning: failed to install hooks for %s: %v\n", client.DisplayName(), err)
			log.Error("failed to install client hooks", "client", client.ID(), "error", err)
		}
	}
}
