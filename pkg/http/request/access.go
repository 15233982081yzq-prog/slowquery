package request

type SetRemoteDBPasswdRequest struct {
	UserName string `json:"user_name" binding:"required"`
	Password string `json:"password" binding:"required"`
	Env      string `json:"env"`
}
