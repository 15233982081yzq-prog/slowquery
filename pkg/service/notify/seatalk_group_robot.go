package notify

import (
	"smart-slowquery/conf"
	"smart-slowquery/internal/util/errors"
	"smart-slowquery/pkg/log"
	"smart-slowquery/pkg/service/cmdb"

	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"git.garena.com/shopee/platform/space-sdk/uic"
	"gopkg.in/resty.v1"
)

var DBANotifyRobot *SeaTalkGroupRobot
var DevNotifyRobot *SeaTalkGroupRobot

type SeaTalkGroupRobot struct {
	RobotIds   []string
	HttpClient *resty.Client
	uicClient  uic.ClientInterface

	atDod    bool
	dodUicId uint64
}

type SeaTalkMessage struct {
	Tag      string           `json:"tag"`
	Text     *SeaTalkText     `json:"text,omitempty"`
	MarkDown *SeaTalkMarkDown `json:"markdown,omitempty"`
}

type SeaTalkText struct {
	Content            string   `json:"content"`
	AtAll              bool     `json:"at_all,omitempty"`
	MentionedEmailList []string `json:"mentioned_email_list,omitempty"`
}

type SeaTalkMarkDown struct {
	Content string `json:"content"`
}

type SeaTalkResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

func NewSeaTalkGroupRobot(atDod bool, dodUic uint64, conf *conf.Space, robotIds ...string) (*SeaTalkGroupRobot, error) {
	if len(robotIds) == 0 {
		return nil, &errors.ErrNotifyRobotIDEmpty
	}

	var (
		restyCli *resty.Client
		uicCli   uic.ClientInterface
		err      error
	)

	restyCli = resty.New()
	restyCli.SetHeader("Content-Type", "application/json").
		SetTimeout(SeaTalkTimeout).
		SetRetryCount(3)

	if uicCli, err = cmdb.NewUicClient(conf); err != nil {
		log.Errorf("init uic client failed, %v", err)
		return nil, err
	}

	return &SeaTalkGroupRobot{
		RobotIds:   robotIds,
		HttpClient: restyCli,
		uicClient:  uicCli,
		atDod:      atDod,
		dodUicId:   dodUic,
	}, nil
}

func (robot *SeaTalkGroupRobot) GetDodMembers() ([]string, error) {
	if !robot.atDod {
		return []string{}, nil
	}

	res, err := robot.uicClient.GetGroupMembers(context.Background(), uic.GetGroupMemberReq{
		GroupIDs: []uint64{robot.dodUicId},
	})
	if err != nil {
		return []string{}, err
	}

	if !res.Success {
		return []string{}, fmt.Errorf("get UIC list error, teamid = %v", robot.dodUicId)
	}

	var ul []string
	for _, group := range res.GroupMembers {
		for _, user := range group.Members {
			ul = append(ul, user.Email)
		}
	}
	if len(ul) == 0 {
		return []string{}, fmt.Errorf("get UIC list error, teamid = %v", robot.dodUicId)
	}
	return []string{ul[0]}, nil
}

func (robot *SeaTalkGroupRobot) pushMsgHttp(msg *SeaTalkMessage) {
	body, err := json.Marshal(msg)
	if err != nil {
		log.Infof("SeaTalkMessage to json error")
		return
	}
	for _, rid := range robot.RobotIds {
		response := &SeaTalkResponse{}
		url := fmt.Sprintf(SeaTalkRobotWebHook, rid)
		res, err := robot.HttpClient.R().
			SetBody(body).
			SetResult(response).
			Post(url)
		if err != nil {
			log.Error(err)
		}
		if res.StatusCode() != http.StatusOK {
			log.Errorf("seaTalk push msg http code:%v", res.StatusCode())
		}
		if response.Code != 0 {
			log.Errorf("seaTalk response error. code:%v, msg:%v", response.Code, response.Msg)
		}
	}
}

func (robot *SeaTalkGroupRobot) PushTextMessage(content string, atAll bool, mentionEmails ...string) {
	seaTalkText := &SeaTalkText{
		Content:            content,
		AtAll:              atAll,
		MentionedEmailList: mentionEmails,
	}
	seaTalkMessage := &SeaTalkMessage{
		Tag:  SeaTalkMsgTypeText,
		Text: seaTalkText,
	}
	robot.pushMsgHttp(seaTalkMessage)
}

func (robot *SeaTalkGroupRobot) PushMarkDownMessage(content string, atAll bool, mentionEmails ...string) {
	for _, email := range mentionEmails {
		content = fmt.Sprintf("<mention-tag target=\"seatalk://user?email=%s\"/>\n\n", email) + content
	}

	if atAll {
		content = "<mention-tag target=\"seaTalk://user?id=0\"/>\n\n" + content
	}

	seaTalkText := &SeaTalkMarkDown{
		Content: content,
	}
	seaTalkMessage := &SeaTalkMessage{
		Tag:      SeaTalkMsgTypeMarkDown,
		MarkDown: seaTalkText,
	}
	robot.pushMsgHttp(seaTalkMessage)
}

func (robot *SeaTalkGroupRobot) PushMarkDownMessageAtDod(content string) {
	dodEmails, err := robot.GetDodMembers()
	if err != nil {
		log.Errorf("get dod emails err: %v", err)
	}
	robot.PushMarkDownMessage(content, false, dodEmails...)
}

func GenDBANotifyMarkDown(DBANotifyParam *DBANotifyParam) string {
	status := "✅"
	if !DBANotifyParam.IsOk {
		status = "❌"
	}

	return fmt.Sprintf(DevNotifyMarkDownTemplate,
		status,
		DBANotifyParam.LogicDB,
		DBANotifyParam.Env,
		DBANotifyParam.Query.Sql,
		DBANotifyParam.Query.ConnectionID,
		DBANotifyParam.Query.Timeout,
		DBANotifyParam.Query.KillHung,
	)
}
