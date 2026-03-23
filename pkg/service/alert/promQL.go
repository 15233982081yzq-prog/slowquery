package alert

import (
	"smart-slowquery/conf"
	"smart-slowquery/pkg/log"

	"bytes"
	"fmt"
	"strings"
	"text/template"
)

type TemplateParam struct {
	DBS             []string
	Expression      string
	ExpressionValue int
	DBInfo          string
}

func (t *TemplateParam) genDBSlot() {
	if len(t.DBS) == 0 {
		t.DBInfo = ""
		return
	}
	t.DBInfo = fmt.Sprintf("database_name=~\"%s\"", strings.Join(t.DBS, "|"))
}

func generatePromQL(templateConf *conf.AlertTemplate, param *TemplateParam) (string, error) {
	var (
		tmpl *template.Template
		err  error
	)
	// 新建模版
	if tmpl, err = template.New("templateParam").Parse(templateConf.PromQLTemplate); err != nil {
		log.Warningf("GeneratePromQL CreateTemplate Error, err:%v", err)
		return "", err
	}

	param.genDBSlot()

	// 生成promQL
	var buffer bytes.Buffer
	if err = tmpl.Execute(&buffer, param); err != nil {
		log.Warningf("GeneratePromQL Execute Err, err:%v", err)
		return "", err
	}

	return buffer.String(), nil
}
