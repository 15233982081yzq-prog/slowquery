package notify

import "time"

const (
	SeaTalkRobotWebHook                  = "https://openapi.seatalk.io/webhook/group/%s"
	SeaTalkTimeout         time.Duration = time.Second * 5
	SeaTalkMsgTypeText                   = "text"
	SeaTalkMsgTypeMarkDown               = "markdown"
)
