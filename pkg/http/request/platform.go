package request

type DataBaseHost struct {
	DBName string `json:"db_name" binding:"required"`
	DBEnv  string `json:"db_env" binding:"required"`
	Roler  string `json:"roler" binding:"required"`
}
