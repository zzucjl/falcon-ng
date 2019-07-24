package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/toolkits/pkg/errors"

	"github.com/open-falcon/falcon-ng/src/dataobj"
	"github.com/open-falcon/falcon-ng/src/modules/sender/redi"
)

func im(c *gin.Context) {
	auth(c)

	var f dataobj.Notify
	errors.Dangerous(c.BindJSON(&f))

	if len(f.Tos) == 0 || len(f.Content) == 0 {
		renderMessage(c, "tos or content cannot be empty")
		return
	}

	redi.Write(&f, "im")

	renderMessage(c, "")
}

func mail(c *gin.Context) {
	auth(c)

	var f dataobj.Notify
	errors.Dangerous(c.BindJSON(&f))

	if len(f.Tos) == 0 || len(f.Content) == 0 || len(f.Subject) == 0 {
		renderMessage(c, "tos, subject or content cannot be empty")
		return
	}

	redi.Write(&f, "mail")

	renderMessage(c, "")
}

func sms(c *gin.Context) {
	auth(c)

	var f dataobj.Notify
	errors.Dangerous(c.BindJSON(&f))

	if len(f.Tos) == 0 || len(f.Content) == 0 {
		renderMessage(c, "tos or content cannot be empty")
		return
	}

	redi.Write(&f, "sms")

	renderMessage(c, "")
}

func voice(c *gin.Context) {
	auth(c)

	var f dataobj.Notify
	errors.Dangerous(c.BindJSON(&f))

	if len(f.Tos) == 0 || len(f.Content) == 0 {
		renderMessage(c, "tos or content cannot be empty")
		return
	}

	redi.Write(&f, "voice")

	renderMessage(c, "")
}
