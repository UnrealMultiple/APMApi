package pack

import (
	"context"
	"os"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/common/hlog"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/UnrealMultiple/APMApi/pkg/errno"
)

// RespError 新API统一错误响应(code/msg平铺)
func RespError(ctx context.Context, c *app.RequestContext, err error) {
	e := errno.ConvertErr(err)

	if e.ErrCode == errno.ServiceErrCode {
		hlog.CtxErrorf(ctx,
			"[%s] %s: %+v",
			string(c.Method()),
			string(c.Path()),
			err,
		)
	}

	c.JSON(consts.StatusOK, utils.H{
		"code": e.ErrCode,
		"msg":  e.ErrMsg,
	})
}

// RespLegacyError 老API错误响应, 保持 FastAPI 的 {"detail": ...} 格式
func RespLegacyError(c *app.RequestContext, statusCode int, detail string) {
	c.JSON(statusCode, utils.H{
		"detail": detail,
	})
}

// RespZipFile 以附件形式返回zip文件
// 整体读入内存后发送, 避免 c.File 的句柄缓存在 Windows 下造成上传时文件占用冲突
func RespZipFile(c *app.RequestContext, filePath, fileName string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Data(consts.StatusOK, "application/zip", data)
	return nil
}
