package errors

const (
	UnknownErrorSynopsis = "unknown error"

	DBErrorSynopsis = "database error"
	DBErrorCode     = 10000

	AppErrorSynopsis = "app error"
	AppErrorCode     = 20000

	ParameterErrorSynopsis = "parameter error"
	ParameterErrorCode     = 30000

	ConfigErrorSynopsis = "config error"
	ConfigErrorCode     = 40000

	PermissionErrorSynopsis = "permission error"
	PermissionErrorCode     = 50000
)

var (
	ErrPermission = NewError(PermissionErrorCode, PermissionErrorSynopsis, "")
)

var (
	// ErrApp .
	ErrApp = NewError(AppErrorCode, AppErrorSynopsis, AppErrorSynopsis)
	// ErrStatusTrans .
	ErrStatusTrans        = AnnotateAppErrorf(&ErrApp, "Error status transfer, please refresh page.")
	ErrNotifyRobotIDEmpty = AnnotateAppErrorf(&ErrApp, "robot id is empty")
	ErrRetryJob           = AnnotateAppErrorf(&ErrApp, "retry job failed, please check job status")
	ErrTaskValidate       = AnnotateAppErrorf(&ErrApp, "some task checked failed，please refresh and check task validate result")
	ErrSQLTypeValidate    = AnnotateAppErrorf(&ErrApp, "SQLType not match sql text, please check sql text")
	ErrForbidExecute      = AnnotateAppErrorf(&ErrApp, "Forbid execute job，Please contact admin or wait a minute")
)

var (
	// ErrDB .
	ErrDB = NewError(DBErrorCode, DBErrorSynopsis, DBErrorSynopsis)
	// ErrRecordNotFound .
	ErrRecordNotFound = AnnotateDBErrorf(&ErrDB, "No record found")
	// ErrZeroAffectedRow .
	ErrZeroAffectedRow = AnnotateDBErrorf(&ErrDB, "Affected rows are zero")
	ErrDBRecordVersion = AnnotateDBErrorf(&ErrDB, "record version error, please refresh")
)

var (
	// ErrParam .
	ErrParam = NewError(ParameterErrorCode, ParameterErrorSynopsis, ParameterErrorSynopsis)
)

var (
	ErrConfigDBUserNotSet = NewError(ConfigErrorCode, ConfigErrorSynopsis, "db_user not set")
	ErrConfigDBPassNotSet = NewError(ConfigErrorCode, ConfigErrorSynopsis, "db_pass not set")
	ErrConfigDBHostNotSet = NewError(ConfigErrorCode, ConfigErrorSynopsis, "db_host not set")
	ErrConfigDBPortNotSet = NewError(ConfigErrorCode, ConfigErrorSynopsis, "db_port not set")
	ErrConfigDBNameNotSet = NewError(ConfigErrorCode, ConfigErrorSynopsis, "db_name not set")
)
