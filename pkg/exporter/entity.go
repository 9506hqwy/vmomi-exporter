package exporter

import (
	"context"
	"log/slog"

	"github.com/9506hqwy/vmomi-exporter/pkg/config"
	"github.com/9506hqwy/vmomi-exporter/pkg/vmomi"
)

const empty = 0

func ToEntityFromRoot(ctx context.Context, roots []config.Root) (*[]vmomi.Entity, error) {
	moTypes := filterEntityType(roots)
	if moTypes == nil {
		return nil, nil
	}

	entities, err := vmomi.GetEntityFromRoot(ctx, *moTypes)
	if err != nil {
		return nil, err
	}

	selected := filterEntity(entities, roots)

	if len(selected) == empty {
		slog.WarnContext(ctx, "Not found root", "roots", roots)
	}

	return &selected, nil
}

func filterEntity(entities *[]vmomi.Entity, roots []config.Root) []vmomi.Entity {
	selected := []vmomi.Entity{}
	for _, e := range *entities {
		for _, r := range roots {
			if e.Type == r.Type && e.Name == r.Name {
				selected = append(selected, e)
				break
			}
		}
	}

	return selected
}

func filterEntityType(roots []config.Root) *[]vmomi.ManagedEntityType {
	moTypes := []vmomi.ManagedEntityType{}
	for _, r := range roots {
		if r.Type == vmomi.ManagedEntityTypeFolder && r.Name == "" {
			return nil
		}

		moTypes = append(moTypes, r.Type)
	}

	return &moTypes
}
