package api

import (
	"night-fury/pkgs/db"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/util/gconv"
)

func PageMeta(pageSize, pageNo, total, number int) map[string]interface{} {
	return map[string]interface{}{
		"pageSize": pageSize, // 本页size
		"pageNo":   pageNo,   // 页数
		"number":   number,   // 本页数量
		"total":    total,    // 总共数量
	}
}

func Success(c *gin.Context, data interface{}, meta interface{}) {
	res := make(map[string]interface{})
	res["data"] = data
	res["code"] = 1
	res["meta"] = meta
	c.AbortWithStatusJSON(200, res)
}

func Fail(c *gin.Context, code int, meta interface{}) {
	res := make(map[string]interface{})
	res["data"] = nil
	res["code"] = -1
	res["meta"] = meta
	c.AbortWithStatusJSON(code, res)
}

func NewMeta(code string, msg ...interface{}) map[string]interface{} {
	meta := make(map[string]interface{}, 5)
	meta["time"] = time.Now()
	if code != "" {
		meta["code"] = code
	}
	if len(msg) == 1 {
		switch msg[0].(type) {
		case string:
			meta["msg"] = msg[0]
		case []byte:
			b := msg[0].([]byte)
			meta["msg"] = b
		case time.Time, *time.Time:
			t := msg[0].(time.Time)
			meta["time"] = t.String()
		default:
			m := gconv.Map(msg)
			for k, v := range m {
				meta[k] = v
			}
		}
		meta["msg"] = msg[0]
	} else if len(msg) == 2 { // 兼容简单 key value 格式
		k, ok := msg[0].(string)
		if !ok {
			return meta
		}
		meta[k] = msg[1]
	}

	return meta
}

func GetSessUser(c *gin.Context) *db.SessUser {
	i, ok := c.Get("_sess_user")
	if !ok {
		return nil
	}
	u, _ := i.(*db.SessUser)
	return u
}
