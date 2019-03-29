package user

import (
	. "apiserver/handler"
	"apiserver/model"
	"github.com/gin-gonic/gin"
)

// Get gets an user by the user identifier.
func Get(c *gin.Context) {
	username := c.Param("username")
	// Get the user by the `username` from the database.
	user, err := model.GetUser(username)
	if err != nil {
		SendResponse(c, err, nil)
	}

	SendResponse(c, nil, user)

}
