package main

import (
	"github.com/6thplaneta/hermes"
	"github.com/gin-gonic/gin"
	"strconv"
)

type RateController struct {
	*hermes.Controller
}

func NewRateController(coll hermes.Collectionist, base string) *RateController {

	cnt := hermes.NewController(coll, base)
	cont := &RateController{cnt}
	return cont
}
func (cont *RateController) CreateRate(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	var rater_id, rated_id int
	var rate_amount float64
	var err error
	Rating_Info := struct {
		Rater_Id, Rated_Id int
		Rate_Amount *float64
	}{}
	c.BindJSON(&Rating_Info)

	if Rating_Info.Rater_Id != 0 {
		rater_id = Rating_Info.Rater_Id
	}
	if Rating_Info.Rated_Id != 0 {
		rated_id = Rating_Info.Rated_Id
	}
	if Rating_Info.Rate_Amount != nil {
		rate_amount = *Rating_Info.Rate_Amount
	}
	id, err := RateColl.CreateRate(token, rater_id, rated_id, rate_amount)
	if err != nil {
		c.JSON(500, nil)
		return
	}
	c.JSON(200, gin.H{"Successfully Created Rate ID:": id})
}
func (cont *RateController)GetRate(c *gin.Context){
	token:=c.Request.Header.Get("Authorization")
	var rate_id int
	if c.Param("id")!="" {
		rate_id,_=strconv.Atoi(c.Param("id"))
	}
	rate,err:=RateColl.Get(token,rate_id,"")
	if err!=nil {
		c.JSON(400,"Not Found!")
		return
	}
	c.JSON(200,rate)
}
