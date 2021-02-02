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

func New(user string, key string, userServiceKey string) (*Cloudflare, error) {
	api, err := cloudflare.New(key, user)
	if err != nil {
		return nil, err
	}

	api.APIUserServiceKey = userServiceKey
	api.AccountID = "86ed4d09664b2b395548c37339c7f179"

	return &Cloudflare{
		api: api,
	}, nil
}
