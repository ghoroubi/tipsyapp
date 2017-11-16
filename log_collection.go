package main

import (
	"time"
)

type Log struct {
	Id            int       `json:"id"`
	User_Id       int       `json:"user_id"`
	Lng           float32   `json:"lng"`
	Lat           float32   `json:"lat"`
	Ip            string    `json:"ip"`
	Creation_Date time.Time `json:"creation_date"`
}
