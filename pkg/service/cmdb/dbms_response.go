package cmdb

// ListDatabasesResponse
// https://confluence.shopee.io/pages/viewpage.action?spaceKey=SPDEV&title=%5BDBMS%5D+API+Documentation+for+Database+Metadata#heading-5Listdatabases
type ListDatabasesResponse struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"error_code"`
	Data      struct {
		PageSize  int `json:"page_size"`
		PageNum   int `json:"page_num"`
		TotalSize int `json:"total_size"`
		Databases []struct {
			DatabaseUuid string `json:"database_uuid"`
			DatabaseName string `json:"database_name"`
			Environment  string `json:"environment"`
			DatabaseType string `json:"database_type"`
			IsRetired    bool   `json:"is_retired"`
			CmdbServices struct {
				ReadServices []struct {
					ServiceName string `json:"service_name"`
				} `json:"read_services"`
				ReadWriteServices []struct {
					ServiceName string `json:"service_name"`
				} `json:"read_write_services"`
				ReadBinlogServices []interface{} `json:"read_binlog_services"`
			} `json:"cmdb_services"`
		} `json:"databases"`
	} `json:"data"`
}

// GetDataBaseResponse
// https://confluence.shopee.io/pages/viewpage.action?spaceKey=SPDEV&title=%5BDBMS%5D+API+Documentation+for+Database+Metadata#heading-6Getdatabasedetails
type GetDataBaseResponse struct {
	Error     string `json:"error"`
	ErrorCode int    `json:"error_code"`
	Data      struct {
		Database struct {
			DatabaseUuid string `json:"database_uuid"`
			DatabaseName string `json:"database_name"`
			DatabaseType string `json:"database_type"`
			Environment  string `json:"environment"`
			IsRetired    bool   `json:"is_retired"`
			CmdbServices struct {
				ReadServices []struct {
					ServiceName string `json:"service_name"`
				} `json:"read_services"`
				ReadWriteServices []struct {
					ServiceName string `json:"service_name"`
				} `json:"read_write_services"`
				ReadBinlogServices []interface{} `json:"read_binlog_services"`
			} `json:"cmdb_services"`
			Clusters []struct {
				ClusterUuid string `json:"cluster_uuid"`
				ClusterName string `json:"cluster_name"`
				ClusterType string `json:"cluster_type"`
				Domains     []struct {
					Role   string `json:"role"`
					Domain string `json:"domain"`
					Port   int    `json:"port"`
				} `json:"domains"`
			} `json:"clusters"`
		} `json:"database"`
	} `json:"data"`
}
