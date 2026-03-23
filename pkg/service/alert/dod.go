package alert

import (
	"smart-slowquery/conf"

	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"time"
)

type DodClient struct {
	client *http.Client
	conf   *conf.DodConfig
}

type ListDevResponse struct {
	ErrCode int    `json:"err_code"`
	ErrMsg  string `json:"err_msg"`
	Data    struct {
		Devs DoDdevs `json:"devs"`
	} `json:"data"`
}

type DoDdevs []*DoDdev

func (d DoDdevs) GetDoDs() DoDdevs {
	ret := make([]*DoDdev, 0)
	for _, dod := range d {
		if dod.IsDOD {
			ret = append(ret, dod)
		}
	}
	return ret
}

type DoDdev struct {
	DevID       int    `json:"dev_id"`
	Username    string `json:"username"`
	RoleID      int    `json:"role_id"`
	Contact     string `json:"contact"`
	Email       string `json:"email"`
	DialingCode string `json:"dialing_code"`
	MMHandle    string `json:"mm_handle"`
	TeamID      int    `json:"team_id"`
	IsDeleted   bool   `json:"is_deleted"`
	IsDOD       bool   `json:"is_dod"`
	DODType     int    `json:"dod_type"`
	ExtInfo     string `json:"ext_info"`
}

func NewDodClient(dodConf *conf.DodConfig) (*DodClient, error) {
	return &DodClient{
		conf: dodConf,
		client: &http.Client{
			Timeout: dodConf.Timeout * time.Second, // 设置HTTP客户端超时时间
		},
	}, nil
}

func (d *DodClient) ListDodsByTeamID(teamID int) ([]string, error) {
	var (
		res  *ListDevResponse
		dods []string
		err  error
	)
	if res, err = d.ListDevsByTeamID(teamID); err != nil {
		return nil, err
	}
	dodDevs := res.Data.Devs.GetDoDs()
	for _, dod := range dodDevs {
		dods = append(dods, dod.Email)
	}
	return dods, nil
}

func (d *DodClient) ListDevsByTeamID(teamID int) (*ListDevResponse, error) {
	var (
		apiURL        string
		req           *http.Request
		resp          *http.Response
		body, payload []byte
		err           error
	)
	if apiURL, err = d.composeEndpointURL("/api/gateway/v1/get_dev_list_of_team"); err != nil {
		return nil, fmt.Errorf("compose url of list dev by team error:%s", err.Error())
	}
	if payload, err = json.Marshal(map[string]interface{}{
		"team_id": teamID,
	}); err != nil {
		return nil, err
	}
	if req, err = http.NewRequest(http.MethodPost, apiURL, bytes.NewBuffer(payload)); err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	if resp, err = d.client.Do(req); err != nil {
		return nil, fmt.Errorf("team_id: %d, list dev by team error:%s", teamID, err.Error())
	}
	defer func() { _ = resp.Body.Close() }()

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		return nil, err
	}
	r := &ListDevResponse{}
	if err = json.Unmarshal(body, r); err != nil {
		return nil, err
	}
	return r, nil
}

func (d *DodClient) composeEndpointURL(pathFragments ...string) (string, error) {
	var (
		u   *url.URL
		err error
	)
	baseUrl := d.conf.DODBaseURL
	if u, err = url.Parse(baseUrl); err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, path.Join(pathFragments...))
	return u.String(), nil
}
