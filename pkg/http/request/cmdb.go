package request

type GetLogicDBByServiceRequest struct {
	CMDBService string `json:"cmdb_service"`
	DBEnv       int    `json:"db_env" binding:"oneof=0 1 2 3 4 5"`
	DBType      int    `json:"db_type" binding:"oneof=0 1"`
}
