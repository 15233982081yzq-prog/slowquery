package response

type SetRemoteDBPasswdResponse struct {
	UserName   string `json:"user_name"`
	Password   string `json:"password"`
	Env        string `json:"env"`
	Key        string `json:"key"`
	UpdateTime string `json:"update_time"`
}
