package request

type CkRequest interface {
	Valid() error
	TableName() string
	LocalTableName() string
	ClusterName() string
}

type BaseRequest struct {
}

func (req *BaseRequest) ClusterName() string {
	return "cluster_szinfra_szinfra_clouddba_online"
}
