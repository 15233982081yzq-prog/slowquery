package service

import (
	"smart-slowquery/pkg/log"

	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

func GetCoreClusters() (*CoreCluster, error) {
	var (
		err          error
		body         []byte
		resp         *http.Response
		coreClusters = CoreCluster{}
	)
	if resp, err = http.Get("https://shadow-watch.fcst.shopee.io/platform/api/v1/get_db_meta_info?CampaignId=149&TrafficType=normal"); err != nil {
		log.Errorf("platform getCoreDBCluster get fcst http error:%s", err.Error())
		return nil, err
	}

	if body, err = ioutil.ReadAll(resp.Body); err != nil {
		log.Errorf("platform getCoreDBCluster get fcst http error:%s", err.Error())
		return nil, err
	}
	clusters := &FCTSCoreDBClusters{}
	_ = json.Unmarshal(body, clusters)
	resp.Body.Close()

	fcClusters := clusters.Data.ClustersInfo.K8SContainer

	for _, clusterName := range fcClusters {
		req, err := http.NewRequest("GET", fmt.Sprintf("http://space.shopee.io/api/rds/live/v1/cluster/get_cluster_list?cluster_name=%s&environment=live", clusterName), nil)
		if err != nil {
			log.Infof("platform getCoreDBCluster get dbaas http error:%s", err.Error())
			return nil, err
		}
		req.Header.Set("Token", "CCOR8fE4s5SwgR2h65vkV7bELoWMQx8b")

		if resp, err = (&http.Client{}).Do(req); err != nil {
			log.Infof("platform getCoreDBCluster get dbaas http error:%s", err.Error())
			return nil, err
		}

		if body, err = ioutil.ReadAll(resp.Body); err != nil {
			log.Infof("platform getCoreDBCluster get dbaas http error:%s", err.Error())
			return nil, err
		}
		dbaasCluters := &DbaasGetClusterList{}
		_ = json.Unmarshal(body, dbaasCluters)
		resp.Body.Close()

		if len(dbaasCluters.Data.Data) > 0 {
			if dbaasCluters.Data.Data[0].SupportSlowQuery {
				coreClusters.RDSs.SupportSlowQuery = append(coreClusters.RDSs.SupportSlowQuery, clusterName)
			} else {
				coreClusters.RDSs.UnSupportSlowQuery = append(coreClusters.RDSs.UnSupportSlowQuery, clusterName)
			}
		}
	}
	coreClusters.Physical = clusters.Data.ClustersInfo.Physical
	return &coreClusters, nil
}

type CoreCluster struct {
	RDSs struct {
		SupportSlowQuery   []string `json:"supports"`
		UnSupportSlowQuery []string `json:"unSupport"`
	}
	Physical []string `json:"physical"`
}

type FCTSCoreDBClusters struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    struct {
		DBsInfo struct {
			MySQL        []string `json:"MySQL"`
			K8SContainer []string `json:"k8s-container"`
			Physical     []string `json:"physical"`
		} `json:"DBsInfo"`
		ClustersInfo struct {
			MySQL        []string `json:"MySQL"`
			K8SContainer []string `json:"k8s-container"`
			Physical     []string `json:"physical"`
		} `json:"ClustersInfo"`
	} `json:"data"`
}

type DbaasGetClusterList struct {
	Data struct {
		TotalPages  int `json:"total_pages"`
		TotalNumber int `json:"total_number"`
		Current     int `json:"current"`
		PageSize    int `json:"page_size"`
		Data        []struct {
			Id                        int       `json:"id"`
			Uuid                      string    `json:"uuid"`
			ClusterName               string    `json:"cluster_name"`
			ClusterType               int       `json:"cluster_type"`
			Version                   int       `json:"version"`
			GroupId                   int       `json:"group_id"`
			GroupUuid                 string    `json:"group_uuid"`
			CreatorEmail              string    `json:"creator_email"`
			OwnerEmail                string    `json:"owner_email"`
			ContactUsers              string    `json:"contact_users"`
			Description               string    `json:"description"`
			ServiceRanking            int       `json:"service_ranking"`
			CreatedTs                 time.Time `json:"created_ts"`
			UpdatedTs                 time.Time `json:"updated_ts"`
			Environment               string    `json:"environment"`
			OverwatchType             int       `json:"overwatch_type"`
			ClusterEngine             int       `json:"cluster_engine"`
			EnableDynamicIoThrottling int       `json:"enable_dynamic_io_throttling"`
			ClusterDrType             string    `json:"cluster_dr_type"`
			Region                    string    `json:"region"`
			RelateClusterUuid         string    `json:"relate_cluster_uuid"`
			RelateClusterName         string    `json:"relate_cluster_name"`
			StorageType               string    `json:"storage_type"`
			IoType                    string    `json:"io_type"`
			IoTypeObject              struct {
				Id                   int    `json:"id"`
				Uuid                 string `json:"uuid"`
				Environment          string `json:"environment"`
				IoType               string `json:"io_type"`
				StorageLowerBound    string `json:"storage_lower_bound"`
				StorageUpperBound    string `json:"storage_upper_bound"`
				IopsFormula          string `json:"iops_formula"`
				BpsFormula           string `json:"bps_formula"`
				StorageSpecification string `json:"storage_specification"`
				MediaType            string `json:"media_type"`
			} `json:"io_type_object"`
			MasterFailover bool   `json:"master_failover"`
			SlaveFailover  bool   `json:"slave_failover"`
			Failover       bool   `json:"failover"`
			Status         int    `json:"status"`
			HostType       int    `json:"host_type"`
			UnitName       string `json:"unit_name"`
			UnitObject     struct {
				Id                    int       `json:"id"`
				Uuid                  string    `json:"uuid"`
				UnitName              string    `json:"unit_name"`
				ServerType            string    `json:"server_type"`
				UnitDescription       string    `json:"unit_description"`
				Environment           string    `json:"environment"`
				MaxUnitScale          int       `json:"max_unit_scale"`
				CpuNumber             int       `json:"cpu_number"`
				MemoryNumber          int       `json:"memory_number"`
				DiskSize              int       `json:"disk_size"`
				ReadOperationsPerSec  int       `json:"read_operations_per_sec"`
				WriteOperationsPerSec int       `json:"write_operations_per_sec"`
				ReadBytesPerSec       int       `json:"read_bytes_per_sec"`
				WriteBytesPerSec      int       `json:"write_bytes_per_sec"`
				StorageType           string    `json:"storage_type"`
				MinimumDiskSize       int       `json:"minimum_disk_size"`
				CreatedTs             time.Time `json:"created_ts"`
				UpdatedTs             time.Time `json:"updated_ts"`
				Cpu                   int       `json:"cpu"`
				Memory                int       `json:"memory"`
			} `json:"unit_object"`
			DatabaseVersion string `json:"database_version"`
			MasterInstance  struct {
				Id                    int      `json:"id"`
				Uuid                  string   `json:"uuid"`
				InstanceType          int      `json:"instance_type"`
				InstanceStatus        int      `json:"instance_status"`
				InstanceRole          int      `json:"instance_role"`
				ClusterId             int      `json:"cluster_id"`
				ClusterUuid           string   `json:"cluster_uuid"`
				ClusterName           string   `json:"cluster_name"`
				InstanceGroupId       int      `json:"instance_group_id"`
				InstanceGroupUuid     string   `json:"instance_group_uuid"`
				Environment           string   `json:"environment"`
				Zone                  string   `json:"zone"`
				DomainId              int      `json:"domain_id"`
				DomainUuid            string   `json:"domain_uuid"`
				Domain                string   `json:"domain"`
				CnameDomains          []string `json:"cname_domains"`
				HostId                int      `json:"host_id"`
				HostUuid              string   `json:"host_uuid"`
				Hostname              string   `json:"hostname"`
				HostType              int      `json:"host_type"`
				IpLan                 string   `json:"ip_lan"`
				Port                  int      `json:"port"`
				Cpu                   int      `json:"cpu"`
				Memory                int      `json:"memory"`
				DiskSize              int      `json:"disk_size"`
				DiskSizeRequest       int      `json:"disk_size_request"`
				ReadOperationsPerSec  int      `json:"read_operations_per_sec"`
				WriteOperationsPerSec int      `json:"write_operations_per_sec"`
				ReadBytesPerSec       int      `json:"read_bytes_per_sec"`
				WriteBytesPerSec      int      `json:"write_bytes_per_sec"`
				UnitName              string   `json:"unit_name"`
				MediaType             string   `json:"media_type"`
				StorageType           string   `json:"storage_type"`
				Az                    string   `json:"az"`
				IoTypeObject          struct {
					Id                   int    `json:"id"`
					Uuid                 string `json:"uuid"`
					Environment          string `json:"environment"`
					IoType               string `json:"io_type"`
					StorageLowerBound    string `json:"storage_lower_bound"`
					StorageUpperBound    string `json:"storage_upper_bound"`
					IopsFormula          string `json:"iops_formula"`
					BpsFormula           string `json:"bps_formula"`
					StorageSpecification string `json:"storage_specification"`
					MediaType            string `json:"media_type"`
				} `json:"io_type_object"`
				UnitObject struct {
					Id                    int       `json:"id"`
					Uuid                  string    `json:"uuid"`
					UnitName              string    `json:"unit_name"`
					ServerType            string    `json:"server_type"`
					UnitDescription       string    `json:"unit_description"`
					Environment           string    `json:"environment"`
					MaxUnitScale          int       `json:"max_unit_scale"`
					CpuNumber             int       `json:"cpu_number"`
					MemoryNumber          int       `json:"memory_number"`
					DiskSize              int       `json:"disk_size"`
					ReadOperationsPerSec  int       `json:"read_operations_per_sec"`
					WriteOperationsPerSec int       `json:"write_operations_per_sec"`
					ReadBytesPerSec       int       `json:"read_bytes_per_sec"`
					WriteBytesPerSec      int       `json:"write_bytes_per_sec"`
					StorageType           string    `json:"storage_type"`
					MinimumDiskSize       int       `json:"minimum_disk_size"`
					CreatedTs             time.Time `json:"created_ts"`
					UpdatedTs             time.Time `json:"updated_ts"`
					Cpu                   int       `json:"cpu"`
					Memory                int       `json:"memory"`
				} `json:"unit_object"`
				ContainerId       string    `json:"container_id"`
				Idc               string    `json:"idc"`
				PhysicalServer    string    `json:"physical_server"`
				CreatedTs         time.Time `json:"created_ts"`
				UpdatedTs         time.Time `json:"updated_ts"`
				ImageVersion      string    `json:"image_version"`
				DatabaseVersion   string    `json:"database_version"`
				K8SDeploymentUuid string    `json:"k8s_deployment_uuid"`
				Problems          string    `json:"problems"`
				Host              struct {
					Id                    int       `json:"id"`
					HostName              string    `json:"host_name"`
					HostType              int       `json:"host_type"`
					Environment           string    `json:"environment"`
					IpLan                 string    `json:"ip_lan"`
					Sn                    string    `json:"sn"`
					Pod                   string    `json:"pod"`
					Cpu                   int       `json:"cpu"`
					Memory                int       `json:"memory"`
					DiskSizeRequest       int       `json:"disk_size_request"`
					DiskSize              int       `json:"disk_size"`
					RaidType              string    `json:"raid_type"`
					NicSpeed              int       `json:"nic_speed"`
					ContainerId           string    `json:"container_id"`
					PhysicalServerId      int       `json:"physical_server_id"`
					PhysicalServerUuid    string    `json:"physical_server_uuid"`
					CreatedTs             time.Time `json:"created_ts"`
					UpdatedTs             time.Time `json:"updated_ts"`
					ReadOperationsPerSec  int       `json:"read_operations_per_sec"`
					WriteOperationsPerSec int       `json:"write_operations_per_sec"`
					ReadBytesPerSec       int       `json:"read_bytes_per_sec"`
					WriteBytesPerSec      int       `json:"write_bytes_per_sec"`
					UnitName              string    `json:"unit_name"`
					IoType                string    `json:"io_type"`
					MediaType             string    `json:"media_type"`
					StorageType           string    `json:"storage_type"`
					Az                    string    `json:"az"`
					ClusterUuid           string    `json:"cluster_uuid"`
					Uuid                  string    `json:"uuid"`
				} `json:"host"`
			} `json:"master_instance"`
			MasterZone string `json:"master_zone"`
			ShadowZone string `json:"shadow_zone"`
			RoGroups   []struct {
				Id            int    `json:"id"`
				Uuid          string `json:"uuid"`
				ClusterId     int    `json:"cluster_id"`
				ClusterUuid   string `json:"cluster_uuid"`
				Name          string `json:"name"`
				Zone          string `json:"zone"`
				InstanceCount int    `json:"instance_count"`
			} `json:"ro_groups"`
			MonitorUrl          string        `json:"monitor_url"`
			MysqlMonitorUrl     string        `json:"mysql_monitor_url"`
			PodMonitorUrl       string        `json:"pod_monitor_url"`
			NodeMonitorUrl      string        `json:"node_monitor_url"`
			IsGdsEnabled        bool          `json:"is_gds_enabled"`
			GdsStatus           string        `json:"gds_status"`
			IsSuezliteEnabled   bool          `json:"is_suezlite_enabled"`
			SuezliteStatus      string        `json:"suezlite_status"`
			DomainOperationType int           `json:"domain_operation_type"`
			IsBoundWithCmdb     bool          `json:"is_bound_with_cmdb"`
			CmdbServiceId       string        `json:"cmdb_service_id"`
			DatabaseNum         int           `json:"database_num"`
			MainSiteClusterUuid string        `json:"main_site_cluster_uuid"`
			Remark              string        `json:"remark"`
			Labels              []interface{} `json:"labels"`
			SupportSlowQuery    bool          `json:"support_slow_query"`
		} `json:"data"`
	} `json:"data"`
	Error        string `json:"error"`
	ErrorMessage string `json:"error_message"`
}
