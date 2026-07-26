package errno

const (
	SuccessCode    = 0
	ServiceErrCode = 10001
	ParamErrCode   = 10002
	AuthErrCode    = 10003

	PluginNotExistErrCode        = 20001
	VersionNotExistErrCode       = 20002
	PluginPackNotUploadedErrCode = 20003
	PluginSyncingErrCode         = 20004
	ScoreInvalidErrCode          = 20005
)

const (
	SuccessMsg = "成功"
)
