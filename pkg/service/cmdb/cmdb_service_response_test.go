package cmdb

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetL2OwnerShip(t *testing.T) {
	tests := []struct {
		parameter         *GetDatabaseDetailResponse
		expectedSrvName   string
		expectedL2Name    string
		expectedApplicant string
		expectedL2Owners  []string
	}{
		{
			parameter: &GetDatabaseDetailResponse{
				Data: &DatabaseDetail{Database: &Database{
					DatabaseUuid: "",
					DatabaseName: "",
					DatabaseType: "",
					ReTired:      false,
					Zone:         "",
					Environment:  "",
					Applicant:    "applicant",
					OwnerShip: &OwnShip{
						ServiceName: "srv_name",
						Organisation: &Organisation{
							Name:   "l2_name",
							Owners: []string{"l2_a", "l2_b", "l2_c"},
						},
						ProductName:    nil,
						SubProductName: nil,
					},
					CMDBServices: struct {
						ReadServices []struct {
							ServiceName string `json:"service_name"`
						} `json:"read_services"`
						ReadWriteServices []struct {
							ServiceName string `json:"service_name"`
						} `json:"read_write_services"`
						ReadBinLogServices []struct {
							ServiceName string `json:"service_name"`
						} `json:"read_bin_log_service"`
					}{},
					Clusters: nil,
				}},
				Error:     "",
				ErrorCode: 0,
			},
			expectedSrvName:   "srv_name",
			expectedL2Name:    "l2_name",
			expectedApplicant: "applicant",
			expectedL2Owners:  []string{"l2_a", "l2_b", "l2_c"},
		},
		{
			parameter: &GetDatabaseDetailResponse{
				Data: &DatabaseDetail{Database: &Database{
					DatabaseUuid: "",
					DatabaseName: "",
					DatabaseType: "",
					ReTired:      false,
					Zone:         "",
					Environment:  "",
					Applicant:    "",
					OwnerShip: &OwnShip{
						ServiceName:    "",
						Organisation:   nil,
						ProductName:    nil,
						SubProductName: nil,
					},
					CMDBServices: struct {
						ReadServices []struct {
							ServiceName string `json:"service_name"`
						} `json:"read_services"`
						ReadWriteServices []struct {
							ServiceName string `json:"service_name"`
						} `json:"read_write_services"`
						ReadBinLogServices []struct {
							ServiceName string `json:"service_name"`
						} `json:"read_bin_log_service"`
					}{},
					Clusters: nil,
				}},
				Error:     "",
				ErrorCode: 0,
			},
			expectedSrvName:   "",
			expectedL2Name:    "",
			expectedApplicant: "",
			expectedL2Owners:  nil,
		},
	}

	for _, resp := range tests {
		srvName, l2Name, applicant, l2Owners := resp.parameter.GetL2OwnerShip()

		assert.Equal(t, resp.expectedSrvName, srvName)
		assert.Equal(t, resp.expectedL2Name, l2Name)
		assert.Equal(t, resp.expectedApplicant, applicant)
		assert.Equal(t, resp.expectedL2Owners, l2Owners)
	}
}

func TestUrls(t *testing.T) {
	domains := Domains{
		&Domain{
			Domain:     "127.0.0.1",
			Port:       0,
			DomainType: "master",
			Role:       "master",
		},
	}
	assert.Equal(t, 1, len(domains.Urls()))
}

func TestGetDatabaseDetailResponse_GetRdsInstanceGroupsNil(t *testing.T) {
	detail := &GetDatabaseDetailResponse{
		Data:      nil,
		Error:     "",
		ErrorCode: 0,
	}
	detail.GetRdsInstanceGroups()
	detail.GetDomainsByClusterUUID("")
	detail.GetOwnerCmdb()
	detail.GetClusterUUIDs()
}

func TestGetDatabaseDetailResponse_GetRdsInstanceGroups(t *testing.T) {
	detail := &GetDatabaseDetailResponse{
		Data:      &DatabaseDetail{},
		Error:     "",
		ErrorCode: 0,
	}
	detail.GetRdsInstanceGroups()
	detail.GetDomainsByClusterUUID("")
	detail.GetOwnerCmdb()
	detail.GetClusterUUIDs()
}
