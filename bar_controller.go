package main

import (
	"github.com/6thplaneta/hermes"

	"github.com/gin-gonic/gin"
	"net/http"
)

type BarController struct {
	*hermes.Controller
}

func NewBarController(coll hermes.Collectionist, base string) *BarController {
	cnt := hermes.NewController(coll, base)
	controller := &BarController{cnt}
	return controller
}

func (cont *BarController)GetStaffs(c *gin.Context){
	token:=c.Request.Header.Get("Authorization")
	//staffs:= []User{}
	var err error
	var venueId string
	if c.Query("venue_id")!="" {
		venueId=c.Query("venue_id")
		if err!=nil {
			c.JSON(http.StatusNotFound,"Error on getting input from posted URL!")
		}
	}
	staffs,err:= BarColl.GetStaffs(token, venueId)
	if err!=nil {
		c.JSON(http.StatusNotFound,"Error on getting input from posted URL!")
	}
	c.JSON(http.StatusOK,staffs)
}
func (cont *BarController) NewBar(c *gin.Context){
	token:=c.Request.Header.Get("Authorization")
	bar := Bar{}
	c.BindJSON(&bar)
	last_id,err:= BarColl.NewBar(token, bar.Venue_Id, bar.Title, bar.Latitude, bar.Longitude)
	if err!=nil {
		c.JSON(500,nil)
		return
	}
	c.JSON(200,gin.H{"last id":last_id})
}
func (cont *BarController)GetBarId(c *gin.Context){
	token:=c.Request.Header.Get("Authorization")
	var venue_id string
	if x:=c.Query("venue_id");x!=""{
		venue_id=x
	}
	bar_id,err:=BarColl.GetBarId(token,venue_id)
	if err!=nil {
		c.JSON(404,err.Error())
	}
	c.JSON(200,bar_id)
}