package response

type GetServiceTreeResponse struct {
	CmdbServices []string `json:"cmdb_services"`
}

type GetUserRoleResponse struct {
	Email string `json:"email"`
	Role  string `json:"role"` // normal:普通用户   admin:管理员用户
}

type GetLogicDBByServiceResponse struct {
	LogicDbs []string `json:"logic_dbs"`
}
