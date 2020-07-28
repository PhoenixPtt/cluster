package errcode

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// ServeJSON attempts to serve the errcode in a JSON envelope. It marshals err
// and sets the content-type header to 'application/json'. It will handle
// ErrorCoder and Errors, and if necessary will create an envelope.
func ServeJSON(c *gin.Context, err error) error {
	// 设定http头中的内容类型属性，在最后使用的c.JSON，所以这里可以暂不设定
	c.Writer.Header().Set("Content-Type", "application/json; charset=utf-8")

	// 定义状态头变量
	var sc int
	// 根据错误的类型，进行数据格式处理以及状态头变量赋值
	switch errs := err.(type) {
	case Errors:
		if len(errs) < 1 {
			break
		}

		if err, ok := errs[0].(ErrorCoder); ok {
			sc = err.ErrorCode().Descriptor().HTTPStatusCode
		}
	case ErrorCoder:
		sc = errs.ErrorCode().Descriptor().HTTPStatusCode
		err = Errors{err} // create an envelope.
	default:
		// We just have an unhandled error type, so just place in an envelope
		// and move along.
		err = Errors{err}
	}

	// 如果状态头为0，则设定为服务器内部错误
	if sc == 0 {
		sc = http.StatusInternalServerError
	}

	// 写入状态头，其实也可以不写入
	c.Writer.WriteHeader(sc)

	// 按照Json格式发送response
	c.JSON(sc, err)

	return nil
}
