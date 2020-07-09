package cloudflare

import (
	"github.com/cloudflare/cloudflare-go"
)

var pageOpts = cloudflare.PaginationOptions{
	PerPage: 25,
	Page:    1,
}

type Cloudflare struct {
	api *cloudflare.API
}

func New(user string, key string) (*Cloudflare, error) {
	api, err := cloudflare.New(key, user)
	if err != nil {
		return nil, err
	}

	return &Cloudflare{
		api: api,
	}, nil
}
