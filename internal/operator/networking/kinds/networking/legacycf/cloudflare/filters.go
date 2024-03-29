package cloudflare

import (
	"context"

	"github.com/cloudflare/cloudflare-go"
)

type Filter struct {
	ID          string `json:"id,omitempty"`
	Expression  string `json:"expression"`
	Paused      bool   `json:"paused"`
	Description string `json:"description"`
}

func (c *Cloudflare) GetFilters(ctx context.Context, domain string) ([]*Filter, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	filters, err := c.api.Filters(ctx, id, pageOpts)
	return filtersToInternalFilters(filters), err
}

func (c *Cloudflare) CreateFilters(ctx context.Context, domain string, filters []*Filter) ([]*Filter, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	createdFilters, err := c.api.CreateFilters(ctx, id, internalFiltersToFilters(filters))
	if err != nil {
		return nil, err
	}

	return filtersToInternalFilters(createdFilters), err
}

func (c *Cloudflare) UpdateFilters(ctx context.Context, domain string, filters []*Filter) ([]*Filter, error) {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return nil, err
	}

	updatedFilters, err := c.api.UpdateFilters(ctx, id, internalFiltersToFilters(filters))
	if err != nil {
		return nil, err
	}

	return filtersToInternalFilters(updatedFilters), err
}

func (c *Cloudflare) DeleteFilters(ctx context.Context, domain string, filterIDs []string) error {
	id, err := c.api.ZoneIDByName(domain)
	if err != nil {
		return err
	}

	for _, filterID := range filterIDs {
		if err := c.api.DeleteFilter(ctx, id, filterID); err != nil {
			return err
		}
	}
	return nil
}

func filtersToInternalFilters(filters []cloudflare.Filter) []*Filter {
	retFilters := make([]*Filter, 0)
	for _, filter := range filters {
		retFilters = append(retFilters, filterToInternalFilter(filter))
	}
	return retFilters
}

func filterToInternalFilter(filter cloudflare.Filter) *Filter {
	return &Filter{
		ID:          filter.ID,
		Expression:  filter.Expression,
		Paused:      filter.Paused,
		Description: filter.Description,
	}
}

func internalFiltersToFilters(filters []*Filter) []cloudflare.Filter {
	retFilters := make([]cloudflare.Filter, 0)
	for _, filter := range filters {
		retFilters = append(retFilters, internalFilterToFiler(filter))
	}
	return retFilters
}

func internalFilterToFiler(filter *Filter) cloudflare.Filter {
	return cloudflare.Filter{
		ID:          filter.ID,
		Expression:  filter.Expression,
		Paused:      filter.Paused,
		Description: filter.Description,
	}
}
