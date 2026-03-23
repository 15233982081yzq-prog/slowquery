package conf

type ReportEmailConfig struct {
	RankSwitch                  bool   `toml:"rank_switch"`
	NewFingerSwitch             bool   `toml:"new_finger_switch"`
	User                        string `toml:"user"`
	Pwd                         string `toml:"pwd"`
	Env                         string `toml:"env"`
	OwnerEmails                 string `toml:"owner_emails"` //split by ','
	SMTPAddress                 string `toml:"smtp_address" json:"smtp_address"`
	MailSender                  string `toml:"mail_sender" json:"mail_sender"`
	DBReportTemplatePath        string `toml:"db_report_template_path"`
	FingerReportTemplatePath    string `toml:"finger_report_template_path"`
	NewFingerReportTemplatePath string `toml:"new_finger_report_template_path"`
	DBRankDetailUrl             string `toml:"db_rank_detail_url"`
	FingerRankDetailUrl         string `toml:"finger_rank_detail_url"`
	NewFingerReportDetailUrl    string `toml:"new_finger_report_detail_url"`
}
