package app

import (
	"context"

	"github.com/caos/orbos/v5/internal/operator/networking/kinds/networking/legacycf/cloudflare"
)

func (a *App) EnsureFilters(ctx context.Context, domain string, filters []*cloudflare.Filter) ([]*cloudflare.Filter, func() error, error) {

	del := func() error { return nil }

	result := make([]*cloudflare.Filter, 0)
	currentFilters, err := a.cloudflare.GetFilters(ctx, domain)
	if err != nil {
		return nil, nil, err
	}

	createFilters, updateFilters := getFilterToCreateAndUpdate(currentFilters, filters)
	if createFilters != nil && len(createFilters) > 0 {
		created, err := a.cloudflare.CreateFilters(ctx, domain, createFilters)
		if err != nil {
			return nil, nil, err
		}

		result = append(result, created...)
	}

	if updateFilters != nil && len(updateFilters) > 0 {
		updated, err := a.cloudflare.UpdateFilters(ctx, domain, updateFilters)
		if err != nil {
			return nil, nil, err
		}

		result = append(result, updated...)
	}

	deleteFilters, restFilters := getFilterToDelete(currentFilters, filters)
	if deleteFilters != nil && len(deleteFilters) > 0 {
		del = func() error {
			return a.cloudflare.DeleteFilters(ctx, domain, deleteFilters)
		}
	} else {
		del = func() error { return nil }
	}
	result = append(result, restFilters...)

	return result, del, nil
}

func getFilterToDelete(currentFilters []*cloudflare.Filter, filters []*cloudflare.Filter) ([]string, []*cloudflare.Filter) {
	deleteFilters := make([]string, 0)
	restFilters := make([]*cloudflare.Filter, 0)

	if filters != nil {
		for _, currentFilter := range currentFilters {
			desc := currentFilter.Description
			found := false
			for _, filter := range filters {
				if desc == filter.Description {
					restFilters = append(restFilters, filter)
					found = true
				}
			}

			if found == false {
				deleteFilters = append(deleteFilters, currentFilter.ID)
			}
		}
	}

	return deleteFilters, restFilters
}

func getFilterToCreateAndUpdate(currentFilters []*cloudflare.Filter, filters []*cloudflare.Filter) ([]*cloudflare.Filter, []*cloudflare.Filter) {
	createFilters := make([]*cloudflare.Filter, 0)
	updateFilters := make([]*cloudflare.Filter, 0)

	if filters != nil {
		for _, filter := range filters {
			found := false
			for _, currentFilter := range currentFilters {
				if currentFilter.Description == filter.Description {

					filter.ID = currentFilter.ID
					if currentFilter.Expression != filter.Expression {
						updateFilters = append(updateFilters, filter)
					}

					found = true
					break
				}
			}
			if found == false {
				createFilters = append(createFilters, filter)
			}
		}
	}

	return createFilters, updateFilters
}
