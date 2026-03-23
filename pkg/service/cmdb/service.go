package cmdb

import (
	cmdbCli "smart-slowquery/pkg/service/cmdb/client"

	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"

	"fmt"

	cmdbSdk "git.garena.com/shopee/platform/space-sdk/cmdb"
)

const L1ProductLine = "shopee"

type Service struct {
	cfg *conf.Space
	Cli *cmdbCli.SpaceCMDBClient
}

type GetServiceByNameResponse struct {
	UUID             string   `json:"uuid"`
	ServiceID        uint64   `json:"service_id"`
	ServiceName      string   `json:"service_name"`
	OrganisationName string   `json:"organisation_name,omitempty"`
	ProductName      string   `json:"product_name,omitempty"`
	SubProductName   string   `json:"sub_product_name,omitempty"`
	Owners           []string `json:"service_owners"`
}

func (resp *GetServiceByNameResponse) GetProductLine() (pl string) {
	pl = L1ProductLine
	if len(resp.OrganisationName) > 0 {
		pl = fmt.Sprintf("%s.%s", pl, resp.OrganisationName)
	}
	if len(resp.ProductName) > 0 {
		pl = fmt.Sprintf("%s.%s", pl, resp.ProductName)
	}
	if len(resp.SubProductName) > 0 {
		pl = fmt.Sprintf("%s.%s", pl, resp.SubProductName)
	}
	return
}

func NewService(cfg *conf.Space) (*Service, error) {
	var (
		err error
		cli *cmdbCli.SpaceCMDBClient
	)

	if cli, err = cmdbCli.NewSpaceCMDBClient(cfg); err != nil {
		return nil, err
	}

	return &Service{
		cfg: cfg,
		Cli: cli,
	}, nil
}

func (srv *Service) GetServiceByName(name string) (*GetServiceByNameResponse, error) {
	log.Infof("cmdbService.GetServiceByName cfg.user:%s ,cfg.pass:%s", srv.cfg.User, srv.cfg.Pass)
	cli, err := cmdbSdk.NewCMDBClient(
		cmdbSdk.WithUsernameOpt(srv.cfg.User),
		cmdbSdk.WithPasswordOpt(srv.cfg.Pass),
	)
	log.Infof("cmdbService.GetServiceByName cfg.user:%s ,cfg.pass:%s,name:%s ,error:%v", srv.cfg.User, srv.cfg.Pass, name, err)
	if err != nil {
		log.Infof("cmdbService.GetServiceByName error:%s", err.Error())
		return nil, err
	}
	log.Infof("cmdbService.GetServiceByName cmdbSdk.NewCMDBClient finish")

	var resp *cmdbSdk.ServiceResponse

	log.Infof("cmdbService.GetServiceByName start GetServiceByName name:%s", name)
	param := cmdbSdk.GetServiceByNameParam{ServiceName: name}
	if resp, err = cli.GetServiceByName(param); err != nil {
		log.Infof("cmdbService.GetServiceByName name:%s ,error:%s", name, err.Error())
		return nil, err
	}
	log.Infof("cmdbService.GetServiceByName GetServiceByName resp:%v", resp)

	if !resp.Success || resp.BusinessCode != 2000 {
		log.Infof("cmdbService.GetServiceByName name:%s ,success:%t ,code:%d", name, resp.Success, resp.BusinessCode)
		return nil, fmt.Errorf("GetServiceByName response error, success:%t ,code:%d", resp.Success, resp.BusinessCode)
	}

	log.Infof("cmdbService.GetServiceByName response:%v", resp)

	return &GetServiceByNameResponse{
		UUID:             resp.Service.UUID,
		ServiceID:        resp.Service.ServiceID,
		ServiceName:      resp.Service.ServiceName,
		OrganisationName: resp.Service.OrganisationName,
		ProductName:      resp.Service.ProductName,
		SubProductName:   resp.Service.SubProductName,
		Owners:           resp.Service.ServiceOwners,
	}, err
}

/**
 *  TODO: mock 接口模拟cmdb服务GetserviceName接口的访问及返回值， 原因：space-sdk v0.4.49 GetServiceByName方法存在bug，接口数据返回异常，已告知cmdb同学bug原因，等待hotfix版本发布。
 */
func (srv *Service) MockGetServiceByName(name string) (*GetServiceByNameResponse, error) {
	return &GetServiceByNameResponse{
		UUID:             "75bc94ca-51a9-48ad-a34e-282977592e0f",
		ServiceID:        24188,
		ServiceName:      "shopee.customer_service_and_chatbot.customer_service.csdms.adminapi",
		OrganisationName: "customer_service_and_chatbot.customer_service",
		ProductName:      "csdms",
		SubProductName:   "",
		Owners:           []string{"fada.ye@shopee.com"},
	}, nil
}
