package exporter

import (
	"context"
	"log/slog"

	"github.com/9506hqwy/vmomi-exporter/pkg/config"
	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

func ToEntityFromRoot(ctx context.Context, roots []config.Root) (*[]vmomi.Entity, error) {
	moTypes := []vmomi.ManagedEntityType{}
	for _, r := range roots {
		if r.Type == vmomi.ManagedEntityTypeFolder && r.Name == "" {
			return nil, nil
		}

		moTypes = append(moTypes, r.Type)
	}

	entities, err := vmomi.GetEntityFromRoot(ctx, moTypes)
	if err != nil {
		return nil, err
	}

	selected := []vmomi.Entity{}
	for _, e := range *entities {
		for _, r := range roots {
			if e.Type == r.Type && e.Name == r.Name {
				selected = append(selected, e)
				break
			}
		}
	}

	if len(selected) == 0 {
		slog.WarnContext(ctx, "Not found root", "roots", roots)
	}

	return &selected, nil
}
