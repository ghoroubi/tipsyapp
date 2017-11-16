package main

import (
	"fmt"
	"github.com/6thplaneta/hermes"
	"reflect"
)

type Bar struct {
	Id        int     `json:"id" hermes:"dbspace:bars,index,searchable"`
	Venue_Id  string  `json:"venue_id" hermes:"index,searchable,editable"`
	Title     string  `json:"title"`
	Latitude  float64 `json:"latitude" hermes:"searchable,editable"`
	Longitude float64 `json:"longitude" hermes:"searchable,editable"`
}

type BarCollection struct {
	*hermes.Collection
}

func NewBarCollection() (*BarCollection, error) {
	coll, err := hermes.NewDBCollection(&Bar{}, application.DataSrc)
	typ := reflect.TypeOf(Bar{})
	bColl := &BarCollection{coll}
	hermes.CollectionsMap[typ] = bColl
	return &BarCollection{coll}, err
}
func (col *BarCollection) NewBar(token string, venue_id, title string, lat, lng float64) (int, error) {
	var err error
	var bar_id int
	id, err := getUserIdByToken(token)
	if err != nil {
		return 0, err
	}
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Create, id, "Create", cnf.CheckAccess) {
		err = givePermission(&token, id)
		if err != nil {
			return 0, err
		}
	}
	query := fmt.Sprintf("INSERT INTO bars(venue_id,title,latitude,longitude) VALUES('%s','%s',%f,%f) RETURNING id", venue_id, title, lat, lng)
	application.DataSrc.DB.QueryRow(query).Scan(&bar_id)
	return bar_id, nil
}
func (col *BarCollection) GetStaffs(token string, venue_id string) ([]User, error) {
	var err error
	var user_id int
	var staffs []User
	// getting user id for check authorization of user to do
	user_id, err = getUserIdByToken(token)
	if err != nil {
		return nil, err
	}
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Read, user_id, "GET", cnf.CheckAccess) {

		err = ownerToken(&token, user_id)

		if err != nil {
			return nil, err
		}
	}
	isExist := col.ExistBar(venue_id)
	if isExist {
		bar_id,err:=col.GetBarId(token,venue_id)
		if err != nil {
			return nil, err
		}
		query := fmt.Sprintf("SELECT * FROM users WHERE users.is_staff=true and users.work_place_id = %d", bar_id)
		 err=application.DataSrc.DB.Select(&staffs,query)
		if err != nil {
			return nil, err
		}
		for _,v:=range staffs{
			UserColl.UpdateRates(&v)
		}
	}
	return staffs, nil
}
func (col *BarCollection) ExistBar(venue_id string) bool {
	isExist := true
	barCount := 0
	query := fmt.Sprintf("SELECT count(*) FROM bars WHERE venue_id='%s'", venue_id)
	err := application.DataSrc.DB.Get(&barCount, query)
	if err != nil {
		return false
	}
	if barCount == 0 {
		isExist = false
		return isExist
	}
	return isExist
}
func (col *BarCollection) GetBarId(token,venue_id string) (int,error) {
	var id int
	user_id, err := getUserIdByToken(token)
	if err != nil {
		return 0, err
	}
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Read, user_id, "GET", cnf.CheckAccess) {

		err = ownerToken(&token, user_id)

		if err != nil {
			return 0, err
		}
	}
	query := fmt.Sprintf("SELECT id FROM bars WHERE venue_id='%s'", venue_id)
	err=application.DataSrc.DB.Get(&id, query)
	if err!=nil {
		return 0,err
	}
	return  id,nil
}











