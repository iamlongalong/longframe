package intercepter

import (
	"night-fury/dashboard/api"
	"night-fury/pkgs/auth"
	"night-fury/pkgs/db"

	"github.com/gin-gonic/gin"
)

func MiddleWareAuth(c *gin.Context) {
	authToken := c.GetHeader("x-auth")
	if authToken == "" {
		c.Status(403)
		return
	}

	jwtClaims, err := auth.JwtTokenValidate(authToken)
	if err != nil {
		api.Fail(c, 403, api.NewMeta(api.CODE_ERR_NOTPERMIT, err))
		return
	}

	c.Set("_sess_user", &db.SessUser{
		Name:  jwtClaims.Name,
		Email: jwtClaims.Email,
		ID:    jwtClaims.ID,
	})
}
