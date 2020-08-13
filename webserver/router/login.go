// "login.go" file is create by Huxd 2020.08.06
// it used to due login opera

package router

import (
	"fmt"
	"log"
	"net/http"
	"time"

	header "clusterHeader"

	"webserver/router/errcode"
	myjwt "webserver/router/jwt"

	jwtgo "github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
)

// 定义用户登录结构体
type userLogin struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

// 登录结果信息
type loginResult struct {
	LoginStatus bool            // 登录验证状态，true表示成功，false表示失败
	Message     string          // 登录相关信息
	User        header.UserInformation // 用户信息结构体对象
	Token       string          // token内容
}

// 用户登录的处理方法
func login(c *gin.Context) {
	// 解析body中的用户内容
	var user userLogin
	if err := c.ShouldBindWith(&user, binding.JSON); err != nil {
		// 此时解析用户内容不是符合格式的内容，返回错误信息
		serveErrorJSON(c,
			errcode.ErrorCodeUnauthorized.WithMessage("待登录的用户信息格式错误"))
	} else { // 此时用户登录信息格式正确，执行验证等后续操作
		// 对用户名和密码进行验证，一般是登录验证服务器
		if bSuccess, userInfor := verifyUser(user); bSuccess {
			generateToken(c, userInfor)
		} else {
			serveErrorJSON(c,
				errcode.ErrorCodeUnauthorized.WithMessage("用户名或密码输入错误"))
		}
	}
}

// 刷新用户token的请求处理方法，此时需要使用创建时生成的CustomClaims对象，暂时不实现刷新token功能
func refreshToken(c *gin.Context) {
	// 从请求头中获取token的内容
	token := myjwt.GetHeaderToken(c)
	// 如果token的内容为空，则直接返回并忽略后续的操作
	if token == "" {
		serveErrorJSON(c,
			errcode.ErrorCodeDenied.WithMessage("请求未携带token，无法刷新token"))
		return
	}

	// 创建jwt对象
	j := myjwt.NewJWT()
	// 执行刷新token操作
	newToken, err := j.RefreshToken(token)
	if err != nil {
		serveErrorJSON(c,
			errcode.ErrorCodeUnknown.WithMessage(fmt.Sprintf("刷新token过程中出现问题：%v", err)))
		return
	}

	// 返回前端的信息
	c.JSON(http.StatusOK, gin.H{
		"Token": newToken,
	})
}

//////////////////////////////////////////////////////////////////////////////////

// 创建两个用户，一个管理员，一个是访客用户，访客用户只需要名称输入正确即可
var administrator userLogin = userLogin{
	User:     "admin",
	Password: "admin",
}
var guest userLogin = userLogin{
	User:     "guest",
	Password: "",
}

// 对用户名和密码进行验证
func verifyUser(user userLogin) (bool, header.UserInformation) {
	// 目前进行简单验证，将来将使用专用的用户认证模块
	if user == administrator {
		return true, header.UserInformation{
			ID:   "001",
			Name: user.User,
			Auth: "high",
		}
	} else if user.User == "guest" {
		return true, header.UserInformation{
			ID:   "9999",
			Name: user.User,
			Auth: "low",
		}
	}

	// 未验证通过返回失败以及空的用户信息
	return false, header.UserInformation{}
}

// 根据用户信息，生成token
func generateToken(c *gin.Context, user header.UserInformation) {
	// 生成一个基本的JWT对象，默认使用
	j := myjwt.NewJWT()

	// 生成自定义声明
	claims := myjwt.CustomClaims{
		user.ID,
		user.Name,
		user.Auth,
		jwtgo.StandardClaims{
			NotBefore: int64(time.Now().Unix() - 10),               // 签名生效时间
			ExpiresAt: int64(time.Now().Unix() + myjwt.ExpireTime), // 过期时间
			Issuer:    "cetc15clusterserver",                       // 签名的发行者
		},
	}

	// 创建一个token
	token, err := j.CreateToken(claims)
	// 如果创建过程中存在问题，则返回错误信息
	if err != nil {
		serveErrorJSON(c,
			errcode.ErrorCodeUnknown.WithMessage("创建token失败"))
		return
	}

	// 调试使用，输出token信息，正式版本应删除
	log.Println(token)

	// 生成返回信息数据
	returnData := loginResult{
		LoginStatus: true,
		Message:     "登录成功",
		User:        user,
		Token:       token,
	}

	// 返回前端的信息
	c.JSON(http.StatusOK, returnData)

	return
}
