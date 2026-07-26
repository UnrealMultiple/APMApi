package errno

import (
	"errors"
	"fmt"
)

type ErrNo struct {
	ErrCode int32
	ErrMsg  string
}

func (e ErrNo) Error() string {
	return fmt.Sprintf("[%d] %s", e.ErrCode, e.ErrMsg)
}

func NewErrNo(code int32, msg string) ErrNo {
	return ErrNo{ErrCode: code, ErrMsg: msg}
}

func (e ErrNo) WithMessage(msg string) ErrNo {
	e.ErrMsg = msg
	return e
}

func (e ErrNo) WithError(err error) ErrNo {
	e.ErrMsg = e.ErrMsg + ": " + err.Error()
	return e
}

var (
	Success    = NewErrNo(SuccessCode, SuccessMsg)
	ServiceErr = NewErrNo(ServiceErrCode, "服务内部错误")
	ParamErr   = NewErrNo(ParamErrCode, "参数错误")
	AuthErr    = NewErrNo(AuthErrCode, "没认证捏~")

	PluginNotExistErr        = NewErrNo(PluginNotExistErrCode, "没有找到这个插件捏~")
	VersionNotExistErr       = NewErrNo(VersionNotExistErrCode, "没有找到这个版本捏~")
	PluginPackNotUploadedErr = NewErrNo(PluginPackNotUploadedErrCode, "插件包当前未上传,无法下载~")
	PluginSyncingErr         = NewErrNo(PluginSyncingErrCode, "API正在同步插件中,请稍后重试~")
	ScoreInvalidErr          = NewErrNo(ScoreInvalidErrCode, "评分只能是0-10分~")
)

// ConvertErr 将任意 error 转换为 ErrNo, 未知错误归为服务内部错误
func ConvertErr(err error) ErrNo {
	var e ErrNo
	if errors.As(err, &e) {
		return e
	}
	return ServiceErr.WithError(err)
}
