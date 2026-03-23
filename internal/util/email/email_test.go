package email

import (
	modelReport "smart-slowquery/internal/model/report"

	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	content     string
	smtpService = "smtp.test.shopee.io:2525"
	sender      = "dbaas.slowquery@notifications.shopee.com"
	subject     = "daily_rank_email_unit_test"
	mailUser    = "dbaas_smart_slow_query"
	mailPwd     = "Pok3OBey83lW1C9GNhx3epSq5BqzqqtB"
	recipients  = []string{"jian.bian@shopee.com"}
	ccs         = []string{"jian.bian@shopee.com"}
	path        = "./test/template/finger_daily_report.html"
	list        []*modelReport.FingerQueryTime
	err         error
)

func TestSendMail(t *testing.T) {
	content, err = BuildEmailContent(path, list)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)

	err = SendMail(smtpService, sender, subject, content, recipients, ccs, mailUser, mailPwd)
	fmt.Println(err)
	assert.NoError(t, err)
}

func TestBuildEmailContent(t *testing.T) {
	content, err := BuildEmailContent(path, list)
	assert.NoError(t, err)
	assert.NotEmpty(t, content)
}

func init() {
	for i := 1; i <= 3; i++ {
		queryTime := &modelReport.FingerQueryTime{
			SerialNo:    i,
			FingerID:    fmt.Sprintf("finger_id_%d", i),
			FingerSql:   fmt.Sprintf("finger_sql_%d", i),
			ClusterUUID: fmt.Sprintf("cluster_uuid_%d", i),
			DBName:      fmt.Sprintf("db_name_%d", i),
			OwnCMDB:     fmt.Sprintf("own_cmdb_%d", i),
			ProductLine: fmt.Sprintf("pl_%d", i),
			Owners:      []string{"owner"},
			Total:       float64(i),
			Count:       i,
			WeekOnWeek:  "-",
			DetailLink:  fmt.Sprintf("url_%d", i),
		}
		list = append(list, queryTime)
	}
}
