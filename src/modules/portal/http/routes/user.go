package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"

	"github.com/open-falcon/falcon-ng/src/model"
	"github.com/open-falcon/falcon-ng/src/modules/portal/config"
)

func userListGet(c *gin.Context) {
	limit := queryInt(c, "limit", 20)
	query := c.Query("query")

	total, err := model.UserTotal(query)
	errors.Dangerous(err)

	list, err := model.UserGets(query, limit, offset(c, limit, total))
	errors.Dangerous(err)

	c.JSON(200, gin.H{
		"list":  list,
		"total": total,
	})
}

type userAddForm struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Dispname string `json:"dispname"`
	Phone    string `json:"phone"`
	Email    string `json:"email"`
	Im       string `json:"im"`
	IsRoot   int    `json:"is_root"`
}

func userAddPost(c *gin.Context) {
	loginRoot(c)

	var f userAddForm
	errors.Dangerous(c.ShouldBind(&f))

	u := model.User{
		Username: f.Username,
		Password: config.CryptoPass(f.Password),
		Dispname: f.Dispname,
		Phone:    f.Phone,
		Email:    f.Email,
		Im:       f.Im,
		IsRoot:   f.IsRoot,
	}

	renderError(c, u.Save())
}
