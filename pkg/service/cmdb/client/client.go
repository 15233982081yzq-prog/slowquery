package client

import (
	"context"
	"smart-slowquery/conf"

	"git.garena.com/shopee/platform/space-sdk/core/space"
)

type SpaceCMDBClient struct {
	url  string
	user string
	pass string
	ctx  context.Context
}

func NewSpaceCMDBClient(conf *conf.Space) (cli *SpaceCMDBClient, err error) {
	cli = &SpaceCMDBClient{
		url:  conf.SpaceHost,
		user: conf.User,
		pass: conf.Pass,
		ctx:  context.TODO(),
	}
	err = cli.check()
	return cli, err
}

func (cli *SpaceCMDBClient) check() (err error) {
	_, err = space.NewSpaceClient(
		space.WithBaseUrl(cli.url),
		space.WithTokenManager(
			space.NewAccountTokenManager(cli.url, cli.user, cli.pass),
		),
	)
	return err
}
