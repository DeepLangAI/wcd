package constslib

const (
	TraceIdKey             = "trace_id"
	OperationIdKey         = "operation_id"
	ModeEnvName            = "MODE_ENV"
	ModeEnvProd            = "prod"
	ModeEnvPre             = "pre"
	ModeEnvDev             = "dev"
	LogPath_Stdout         = "stdout"
	DataTooLongUpper       = 10000
	DataTooLongUpperErrMsg = "data too long"
)

const (
	HttpHeaderTraceId = "Trace-Id"
	HttpHeaderVersion = "X-API-Version"
)
