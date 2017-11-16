package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"github.com/6thplaneta/hermes"
)

type UserController struct {
	*hermes.Controller
}

func NewUserController(coll hermes.Collectionist, base string) *UserController {
	cnt := hermes.NewController(coll, base)
	cont := &UserController{cnt}
	return cont
}

type Password struct {
	OldPassword string
	NewPassword string
}

func (cont *UserController) NewActivationToken(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")

	_, err := UserColl.NewActivationToken(token, c.Param("email"))

	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}
	c.JSON(http.StatusOK, Messages["ResourceCreated"])

}
func (cont *UserController) RequestPasswordToken(c *gin.Context) {

	agentToken, err := hermes.AgentColl.RequestPasswordToken(c.Param("identity"))
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	emailSender := application.GetSettings("email_provider")

	result, err := hermes.AgentColl.Get(hermes.SystemToken, agentToken.Agent_Id, "")
	agent := result.(*hermes.Agent)

	strSubject := Messages["ChangePasswordSubject"]
	strMessage := fmt.Sprintf(Messages["ChangePasswordBody"], agentToken.Token)
	go hermes.SendEmail(strSubject, strMessage, agent.Identity, emailSender, true)

	c.JSON(http.StatusOK, Messages["ResourceCreated"])

}

func (cont *UserController) List(c *gin.Context) {

	if hermes.RedirectPage(c) {
		return
	}
	params := cont.ReadParams(c)
	// var pageInfo PageInfo
	token := c.Request.Header.Get("Authorization")
	pg, err := hermes.ReadPaging(c)
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
	}
	// var user User
	var CustomWhere = ""
	if c.Query("$lng") != "" && c.Query("$lat") != "" && c.Query("$distance") != "" {
		CustomWhere = " (point( " + c.Query("$lng") + "," + c.Query("$lat") + ") <@> point(longitude, latitude)) < " + c.Query("$distance")
	}
	paramsList := params.List
	paramsList["$$custom"] = hermes.Filter{Type: "exact", Value: CustomWhere}

	result, err := cont.Coll.List(token, params, pg, c.Query("$populate"), "")
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, result)
}

func (cont *UserController) GetProfile(c *gin.Context) {

	var agentColl *hermes.AgentCollection
	var agent hermes.Agent
	token := c.Request.Header.Get("Authorization")
	agent, err := agentColl.GetByLoginToken(token)
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}
	//get user by agent_id and populate agent
	result, err := UserColl.Collection.ListQuery("agent_id="+strconv.Itoa(agent.Id), "Agent")
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}
	userL := *result.(*[]User)
	user := userL[0]

	c.JSON(http.StatusOK, user)

}

func (cont *UserController) EditProfile(c *gin.Context) {
	user := &User{}
	err := c.BindJSON(&user)
	if err != nil {
		hermes.HandleHttpError(c, hermes.ErrJsonFormat, application.Logger)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	token := c.Request.Header.Get("Authorization")
	user, err = UserColl.EditProfile(token, id, user)

	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, hermes.Messages["ResourceUpdated"])
}

func (cont *UserController) UpdateLocation(c *gin.Context) {
	user := &User{}
	err := c.BindJSON(&user)
	if err != nil {
		hermes.HandleHttpError(c, hermes.ErrJsonFormat, application.Logger)
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	token := c.Request.Header.Get("Authorization")
	err = UserColl.UpdateLocation(token, id, user.Current_Longitude, user.Current_Latitude)

	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, hermes.Messages["ResourceUpdated"])
}

func (cont *UserController) FBLogin(c *gin.Context) {

	fb_token := c.Param("fb_token")
	fmt.Println("fbtoken:",fb_token)
	user := &User{}
	err := c.BindJSON(&user)
	if err != nil {
		hermes.HandleHttpError(c, hermes.ErrJsonFormat, application.Logger)
		return
	}

	user.Agent.Device.Ip = c.ClientIP()
	result, userId, err := UserColl.FBLogin(user, fb_token)

	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userId, "agent_token": result})

}

func (cont *UserController) TipsyLogin(c *gin.Context) {
	user := &User{}
	err := c.BindJSON(&user)
	if err != nil {
		hermes.HandleHttpError(c, hermes.ErrJsonFormat, application.Logger)
		return
	}
	user.Agent.Device.Ip = c.ClientIP()
	result, userId, err := UserColl.TipsyLogin(user, "")
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, gin.H{"user_id": userId, "agent_token": result})
}

func (cont *UserController) ExistsByEmail(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	result, err := UserColl.ExistsByEmail(token, c.Param("email"))
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
		return
	}

	c.JSON(http.StatusOK, result)

}

