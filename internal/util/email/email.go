package email

import (
	goSasl "github.com/emersion/go-sasl"
	goSmtp "github.com/emersion/go-smtp"
	log "github.com/sirupsen/logrus"

	"bytes"
	"fmt"
	"strings"
	"text/template"
)

const delimiter = "**=ah46H4yhVTZyCG5FtzbyKSUTB6D4AjLwY1ZNJz4MeAZWsfZth7"

func SendMail(smtpServer, sender, subject, content string, recipients, ccs []string, emailUser, emailPwd string) error {
	auth := goSasl.NewPlainClient("", emailUser, emailPwd)
	// basic email headers
	msg := fmt.Sprintf("From: %s\r\n", sender)
	msg += fmt.Sprintf("To: %s\r\n", strings.Join(recipients, ","))
	msg += fmt.Sprintf("Cc: %s\r\n", strings.Join(ccs, ","))
	msg += fmt.Sprintf("Subject: %s\r\n", subject)

	// Mark content to accept multiple contents
	msg += "MIME-Version: 1.0\r\n"
	msg += fmt.Sprintf("Content-Type: multipart/mixed; boundary=\"%s\"\r\n", delimiter)

	// place HTML message
	msg += fmt.Sprintf("\r\n--%s\r\n", delimiter)
	msg += "Content-Type: text/html; charset=\"utf-8\"\r\n"
	msg += "Content-Transfer-Encoding: 7bit\r\n"
	msg += fmt.Sprintf("\r\n%s", content)

	c, err := goSmtp.Dial(smtpServer)
	if err != nil {
		return err
	}
	defer c.Close()
	if err = c.Auth(auth); err != nil {
		return err
	}
	recipients = append(recipients, ccs...)

	return c.SendMail(sender, recipients, strings.NewReader(msg))
}

func BuildEmailContent(templatePath string, data any) (string, error) {
	// parse template to string
	var doc bytes.Buffer
	reportTemplate, err := template.ParseFiles(templatePath)
	if err != nil {
		return "", err
	}

	err = reportTemplate.Execute(&doc, data)
	if err != nil {
		log.Errorf("report rendered error:%s", err.Error())
		return "", err
	}

	return doc.String(), err
}
