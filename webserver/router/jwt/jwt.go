// 自制 基于jwt-go和gin的Token认证机制功能包 by Huxd 2020.08.06
// JSON Web Token（JWT）是一个非常轻巧的规范。这个规范允许我们使用JWT在用户和服务器之间传递安全可靠的信息。
// JWT由三部分组成，头部、载荷与签名
// 头部，通常包括两部分：token类型（JWT），和使用到的算法，如HMAC、SHA256或RSA；
// 载荷，就是要传递出去的声明，其中包含了实体（通常是用户）和附加元数据，一般包含保留声明、公共声明和私有声明
// 签名，将上面两部分编码后，使用.连接在一起，形成了xxxxx.yyyyyy，采用头部指定的算法，和私钥对上面的字符串进行签名。

package jwt

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"webserver/router/errcode"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// 从请求头中获取token字符串
func GetHeaderToken(c *gin.Context) string  {
	// 一般的实现是将token内容放在Authorization头的Bearer中
	auth := c.Request.Header.Get("Authorization")

	// 使用空格进行分段处理
	content := strings.Split(auth, " ")
	// 遍历生成的字符串切片，如果找到“Bearer”后面跟随的字符串，则返回该字符串，也就是token
	sum := len(content)
	for i := 0; i<sum; i=i+2 {
		if content[i] == "Bearer" {
			return content[i+1]
		}
	}

	// 如果未找到需要的内容，则直接返回空字符串
	return ""
}

// JWTAuth方法为自定义中间件，用于检查token的合法性
func JWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头中获取token的内容
		token := GetHeaderToken(c)
		// 如果token的内容为空，则直接返回并忽略后续的操作
		if token == "" {
			// 返回错误信息内容
			serveErrorJSON(c,
				errcode.ErrorCodeUnauthorized.WithMessage("请求未携带有效格式的token，无权限访问"))
			// 后续操作被忽略
			c.Abort()
			return
		}

		// 调试输出，正式版可删除
		//log.Print("JWTAuth get token: ", token)

		// 创建jwt对象
		j := NewJWT()
		// 解析token包含的信息
		claims, err := j.ParseToken(token)
		// 如果出现错误，则进行错误判断，并确定返回给前段的错误信息
		if err != nil {
			// 如果token已经失效
			if err == TokenExpired && claims != nil {
				// 处理token失效的情况
				resolveTokenExpired(c, j, *claims)
				// 后续操作被忽略
				c.Abort()
				return
			}

			// 其他错误情况，直接返回err的内容
			serveErrorJSON(c,
				errcode.ErrorCodeDenied.WithMessage(err.Error()))
			c.Abort()
			return
		}
		// 继续交由下一个路由处理,并将解析出的信息传递下去
		c.Set("claims", claims)
	}
}

// JWT 结构体
type JWT struct {
	SigningKey []byte
}

// 在jwt包中需要使用的变量
var (
	// 响应头中的自定义字段key值
	//headerTokenRefresh string = "token-refresh"		// token刷新标志名称
	//headerTokenExpired string = "token-expired"		// token过期标志名称
	headerErrorChildCode string = "error-child-code"	// 错误信息的子码

	//TokenExpired     error = errors.New("Token is expired")
	TokenExpired error = errors.New("Token失效")
	//TokenNotValidYet error = errors.New("Token not active yet")
	//TokenNotValidYet error = errors.New("Token未被激活")
	//TokenMalformed   error = errors.New("That's not even a token")
	//TokenMalformed error = errors.New("请求中的Token结构错误")
	//TokenInvalid     error = errors.New("Couldn't handle this token:")
	//TokenInvalid error = errors.New("不能处理请求中的Token")

	SignKey string = "cetc15clusterserver" // 签名使用的关键字

	ExpireTime int64 = 60000 // token过期时间，单位s
)

// 载荷，可以加一些自己需要的信息
type CustomClaims struct {
	ID                 string `json:"userId"`
	Name               string `json:"name"`
	Auth               string `json:"authority"`
	jwt.StandardClaims        // 在jwt-go中定义的标准claim
}

// 新建一个jwt实例
func NewJWT() *JWT {
	return &JWT{
		[]byte(GetSignKey()),
	}
}

// 获取签名关键字
func GetSignKey() string {
	return SignKey
}

// 设置签名关键字
func SetSignKey(key string) string {
	SignKey = key
	return SignKey
}

// CreateToken 根据自定义的声明结构体创建一个token
func (j *JWT) CreateToken(claims CustomClaims) (string, error) {
	// 目前使用HS256算法生成token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.SigningKey)
}

// 解析获取的Token信息
func (j *JWT) ParseToken(tokenString string) (*CustomClaims, error) {
	// 声明 自定义的声明对象
	claims := &CustomClaims{}
	// 使用jwt-go的方法对token进行解码，获取自定义声明中的具体信息
	token, err := jwt.ParseWithClaims(tokenString, claims,
		func(token *jwt.Token) (interface{}, error) {
			return j.SigningKey, nil
		})
	//log.Println(token, claims)

	// 如果解析失败，则根据错误类型，返回不同的错误信息
	if err != nil {
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorExpired != 0 {
				// Token is expired
				return claims, TokenExpired
				//} else if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				//	return nil, TokenMalformed
				//} else  else if ve.Errors&jwt.ValidationErrorNotValidYet != 0 {
				//	return nil, TokenNotValidYet
				//} else {
				//	return nil, TokenInvalid
			} else {
				return nil, err
			}
		}
	}

	// 如果token对象不为空，则进行后续判断
	if token != nil {
		// 如果token中的声明变量可以转换为自定义声明对象，且验证通过，则返回转换后的自定义声明对象
		if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
			return claims, nil
		}
	}

	// 此时返回空的自定义声明对象和错误信息
	return nil, errors.New("token不合法或无法正常转换为CustomClaims对象")
}

// 更新token，此时需要使用创建时生成的CustomClaims对象，所以暂时无法实现刷新功能
func (j *JWT) RefreshToken(tokenString string) (string, error) {
	//jwt.TimeFunc = func() time.Time {
	//	return time.Unix(0, 0)
	//}
	// 使用jwt-go的方法对token进行解析
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return j.SigningKey, nil
	})
	if err != nil {
		return "", err
	}
	// 如果token中的声明变量可以转换为自定义声明对象，且验证通过，则设置有效期，并创建token
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		jwt.TimeFunc = time.Now
		claims.StandardClaims.ExpiresAt = time.Now().Add(1 * time.Hour).Unix()
		return j.CreateToken(*claims)
	}
	return "", errors.New("token不合法或无法正常转换为CustomClaims对象")
}

// 处理token的情况
func resolveTokenExpired(c *gin.Context, j *JWT, claims CustomClaims) {
	// 获得当前时间
	nowTime := time.Now().Unix()
	// 计算token失效后的时间差
	expiredTime := nowTime - claims.ExpiresAt

	// 如果过期在一个自定义的失效时间内，则刷新一个新的token并返回，但需要前端重新请求
	if expiredTime <= ExpireTime && expiredTime > 0 {
		// 设置响应头中的错误信息子码（ErrorChildCode），目前序号1表示刷新操作
		c.Header(headerErrorChildCode, "1")
		// 执行token刷新操作
		// 设置token的启用时间和过期时间
		claims.StandardClaims.NotBefore = int64(nowTime - 10)
		claims.StandardClaims.ExpiresAt = int64(nowTime + ExpireTime)
		// 通过自定义声明对象，创建新的token
		token, err := j.CreateToken(claims)
		// 如果创建过程正常，则返回token字符串，仅仅返回纯粹的token字符串
		if err == nil {
			// 将token字符串写入响应头中
			c.Writer.Header().Set("token", token)
			serveErrorJSON(c,
				errcode.ErrorCodeDenied.WithMessage("请求携带的token已经过期，已返回刷新后的token信息，请前端更新"))
		} else { // 出现错误则返回错误信息
			serveErrorJSON(c,
				errcode.ErrorCodeUnknown.WithMessage(fmt.Sprintf("刷新token过程中出现错误：", err)))
		}

	} else { // 如果超过一个自定义的失效时间，则返回token无效，前端需要重新登录，这些过程都可以再讨论的
		serveErrorJSON(c,
			errcode.ErrorCodeUnauthorized.WithMessage("请求携带的token已经过期，请重新登录"))
	}
}

// 返回错误信息
func serveErrorJSON(c *gin.Context, err errcode.Error) {
	// 添加跨域头
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 返回客户端错误信息
	errcode.ServeJSON(c, err)
}
