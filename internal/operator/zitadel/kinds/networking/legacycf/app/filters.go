package app

import (
	"github.com/caos/orbos/internal/operator/zitadel/kinds/networking/legacycf/cloudflare"
)

func (a *App) EnsureFilters(domain string, filters []*cloudflare.Filter) ([]*cloudflare.Filter, func() error, error) {

	del := func() error { return nil }

	result := make([]*cloudflare.Filter, 0)
	currentFilters, err := a.cloudflare.GetFilters(domain)
	if err != nil {
		return nil, nil, err
	}

	createFilters, updateFilters := getFilterToCreateAndUpdate(currentFilters, filters)
	if len(createFilters) > 0 {
		created, err := a.cloudflare.CreateFilters(domain, createFilters)
		if err != nil {
			return nil, nil, err
		}

		result = append(result, created...)
	}

	if len(updateFilters) > 0 {
		updated, err := a.cloudflare.UpdateFilters(domain, updateFilters)
		if err != nil {
			return nil, nil, err
		}

		result = append(result, updated...)
	}

	deleteFilters := getFilterToDelete(currentFilters, filters)
	if len(deleteFilters) > 0 {
		del = func() error {
			return a.cloudflare.DeleteFilters(domain, deleteFilters)
		}
	}

	return result, del, nil
}

func getFilterToDelete(currentFilters []*cloudflare.Filter, filters []*cloudflare.Filter) []string {
	deleteFilters := make([]string, 0)

	for _, currentFilter := range currentFilters {
		desc := currentFilter.Description
		found := false
		for _, filter := range filters {
			if desc == filter.Description {
				found = true
			}
		}

		if found == false {
			deleteFilters = append(deleteFilters, currentFilter.ID)
		}
	}

	return deleteFilters
}

func getFilterToCreateAndUpdate(currentFilters []*cloudflare.Filter, filters []*cloudflare.Filter) ([]*cloudflare.Filter, []*cloudflare.Filter) {
	createFilters := make([]*cloudflare.Filter, 0)
	updateFilters := make([]*cloudflare.Filter, 0)

	for _, filter := range filters {
		found := false
		for _, currentFilter := range currentFilters {
			if currentFilter.Description == filter.Description ||
				currentFilter.Expression == filter.Expression {

				filter.ID = currentFilter.ID
				updateFilters = append(updateFilters, filter)
				found = true
				break
			}
		}
		if found == false {
			createFilters = append(createFilters, filter)
		}
	}

	return createFilters, updateFilters
}
