package session

import (
	"fmt"
	"night-fury/dashboard/api"

	"github.com/gin-gonic/gin"
)

type SigninParams struct {
	UserID   string `json:"userID"`
	Password string `json:"password"`
}
type SigninRes struct {
	Token string `json:"token"`
}

// @Title 登录接口
// @Description 用户登录
// @Param data body SigninParams true "用户id, 用户密码"
// @Success 200 {object} SigninRes res
// @Router	/license/api/v1/user/signin [post]
func Signin(c *gin.Context) {
	fmt.Println("you sign in")
	params := &SigninParams{}
	err := c.BindJSON(params)
	if err != nil {
		api.Fail(c, 400, api.NewMeta(api.CODE_ERR_PARAMMETER, "parmeter error"))
		return
	}

	res := &SigninRes{}

	api.Success(c, res, nil)
}
