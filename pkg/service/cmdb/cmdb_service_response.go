package cmdb

var (
	rds = map[string]bool{
		"eru-container": true,
		"k8s-container": true,
	}
)

// GetServicesResponse
// https://rap.shopee.io/organization/repository/editor?id=329&itf=15470
type GetServiceTreeResponse struct {
	Services []struct {
		ServiceId      int    `json:"service_id"`
		ServiceName    string `json:"service_name"`
		Identifier     string `json:"identifier"`
		TenantName     string `json:"tenant_name"`
		ProductName    string `json:"product_name"`
		SubProductName string `json:"sub_product_name"`
		GitLink        string `json:"git_link"`
		Data           struct {
			HasLiveContainers bool   `json:"has_live_containers"`
			HasServers        bool   `json:"has_servers"`
			Impact            string `json:"impact"`
			Personas          struct {
				App struct {
					HasBotOwnerOnly  bool `json:"has_bot_owner_only"`
					HasNoOwner       bool `json:"has_no_owner"`
					MismatchAppModel bool `json:"mismatch_app_model"`
					MissingAppModel  bool `json:"missing_app_model"`
				} `json:"app"`
			} `json:"personas"`
		} `json:"data"`
		UpdatedBy     string   `json:"updated_by"`
		ServiceOwners []string `json:"service_owners"`
		IsProtected   bool     `json:"is_protected"`
		IsZombie      bool     `json:"is_zombie"`
		Enabled       bool     `json:"enabled"`
		CreatedAt     int      `json:"created_at"`
		UpdatedAt     int      `json:"updated_at"`
	} `json:"services"`
	TotalSize    int  `json:"total_size"`
	BusinessCode int  `json:"biz_code"`
	Success      bool `json:"success"`
}

// GetDatabaseDetailResponse
// https://confluence.shopee.io/pages/viewpage.action?spaceKey=SPDEV&title=%5BDBMS%5D+API+Documentation+for+Database+Metadata#heading-6Getdatabasedetails

type GetDatabaseDetailResponse struct {
	Data      *DatabaseDetail `json:"data"`
	Error     string          `json:"error"`
	ErrorCode int             `json:"error_code"`
}

type DatabaseDetail struct {
	Database *Database `json:"database"`
}

type Database struct {
	DatabaseUuid string   `json:"database_uuid"`
	DatabaseName string   `json:"database_name"`
	DatabaseType string   `json:"database_type"`
	ReTired      bool     `json:"is_retired"`
	Zone         string   `json:"zone"`
	Environment  string   `json:"environment"`
	Applicant    string   `json:"person_in_charge"`
	OwnerShip    *OwnShip `json:"service_in_charge"`
	CMDBServices struct {
		ReadServices []struct {
			ServiceName string `json:"service_name"`
		} `json:"read_services"`
		ReadWriteServices []struct {
			ServiceName string `json:"service_name"`
		} `json:"read_write_services"`
		ReadBinLogServices []struct {
			ServiceName string `json:"service_name"`
		} `json:"read_bin_log_service"`
	} `json:"cmdb_services"`
	Clusters []struct {
		ClusterUUid    string            `json:"cluster_uuid"`
		RdsClusterUUid string            `json:"rds_cluster_uuid"`
		ClusterName    string            `json:"cluster_name"`
		ClusterType    string            `json:"cluster_type"`
		ClusterDrType  string            `json:"cluster_dr_type"`
		ClusterZone    string            `json:"cluster_zone"`
		Instances      []*InstanceGroups `json:"instance_groups"`
		Domains        []*Domain         `json:"domains"`
	} `json:"clusters"`
}

type OwnShip struct {
	ServiceName    string        `json:"service_name"`
	Organisation   *Organisation `json:"organisation"` //L1 or L1.L2
	ProductName    *Organisation `json:"product"`
	SubProductName *Organisation `json:"sub_product"`
}

type Organisation struct {
	Name   string   `json:"name"`
	Owners []string `json:"owners"`
}

type Domains []*Domain

func (d Domains) Urls() (urls []string) {
	for _, v := range d {
		urls = append(urls, v.Domain)
	}
	return
}

type Domain struct {
	Domain     string `json:"domain"`
	Port       int    `json:"port"`
	DomainType string `json:"domain_type"`
	Role       string `json:"role"`
}

type InstanceGroups struct {
	Role      string      `json:"role"`
	Instances []*Instance `json:"instances"`
}

type Instance struct {
	InstanceID   string `json:"instance_uuid"`
	InstanceType string `json:"instance_type"`
	Role         string `json:"role"`
	IPLan        string `json:"ip_lan"`
	Port         int    `json:"port"`
}

func (detail *GetDatabaseDetailResponse) GetRdsInstanceGroups() (groups []*InstanceGroups) {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.Clusters == nil {
		return nil
	}

	for _, cluster := range detail.Data.Database.Clusters {
		if rds[cluster.ClusterType] && len(cluster.Instances) > 0 {
			groups = append(groups, cluster.Instances...)
		}
	}
	return
}

func (detail *GetDatabaseDetailResponse) GetDomainsByClusterUUID(clusterUUID string) (domains []*Domain) {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.Clusters == nil {
		return nil
	}

	for _, cluster := range detail.Data.Database.Clusters {
		if rds[cluster.ClusterType] && len(cluster.Domains) > 0 && cluster.RdsClusterUUid == clusterUUID {
			domains = append(domains, cluster.Domains...)
		}
	}
	return
}

func (detail *GetDatabaseDetailResponse) GetOwnerCmdb() string {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.OwnerShip == nil {
		return ""
	}
	return detail.Data.Database.OwnerShip.ServiceName
}

func (detail *GetDatabaseDetailResponse) GetL1L2() string {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.OwnerShip == nil {
		return ""
	}
	if detail.Data.Database == nil || detail.Data.Database.OwnerShip == nil || detail.Data.Database.OwnerShip.Organisation == nil {
		return ""
	}
	return detail.Data.Database.OwnerShip.Organisation.Name
}

func (detail *GetDatabaseDetailResponse) GetTeam() string {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.OwnerShip == nil {
		return ""
	}
	if detail.Data.Database == nil || detail.Data.Database.OwnerShip == nil || detail.Data.Database.OwnerShip.ProductName == nil {
		return ""
	}
	return detail.Data.Database.OwnerShip.ProductName.Name
}

func (detail *GetDatabaseDetailResponse) GetRoleMap() map[string]string {
	roleMap := make(map[string]string)
	if detail.Data == nil || detail.Data.Database == nil || len(detail.Data.Database.Clusters) == 0 {
		return roleMap
	}
	for _, cluster := range detail.Data.Database.Clusters {
		for _, instanceInfo := range cluster.Instances {
			for _, instance := range instanceInfo.Instances {
				roleMap[instance.IPLan] = instance.Role
			}
		}
	}
	return roleMap
}

func (detail *GetDatabaseDetailResponse) GetL2OwnerShip() (string, string, string, []string) {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.OwnerShip == nil || detail.Data.Database.OwnerShip.Organisation == nil {
		return "", "", "", nil
	}

	orgL2 := detail.Data.Database.OwnerShip.Organisation

	return detail.Data.Database.OwnerShip.ServiceName, orgL2.Name, detail.Data.Database.Applicant, orgL2.Owners
}

func (detail *GetDatabaseDetailResponse) GetClusterUUIDs() (uuids []string) {
	if detail.Data == nil || detail.Data.Database == nil || detail.Data.Database.Clusters == nil {
		return
	}

	clusters := detail.Data.Database.Clusters
	for idx := range clusters {
		uuids = append(uuids, clusters[idx].RdsClusterUUid)
	}
	return
}
