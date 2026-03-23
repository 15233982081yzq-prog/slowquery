package cmdb

import (
	"fmt"
	"smart-slowquery/conf"

	"git.garena.com/shopee/platform/space-sdk/core/space"
	"git.garena.com/shopee/platform/space-sdk/uic"
)

/** https://git.garena.com/shopee/platform/space-sdk/-/blob/master/uic/client.go **/

func NewUicClient(spaceConfig *conf.Space) (uic.ClientInterface, error) {
	var (
		spaceClient space.Client
		err         error
	)
	if spaceConfig == nil {
		return nil, fmt.Errorf("config is empty")
	}

	spaceClient, err = space.NewSpaceClient(
		space.WithBaseUrl(spaceConfig.SpaceHost),
		space.WithTokenManager(
			space.NewAccountTokenManager(spaceConfig.SpaceHost, spaceConfig.User, spaceConfig.Pass),
		))
	if err != nil {
		return nil, err
	}
	return uic.NewUICClient(&spaceClient), nil
}
