package main

import (
	"github.com/6thplaneta/hermes"
	"reflect"
	"fmt"

)

type Rate struct {
	Id          int `json:"id" hermes:"dbspace:rates,index,searchable"`
	Rater_Id    int `json:"rater_Id" hermes:"index,searchable,editable,many2one"`
	Rated_Id    int `json:"rated_Id" hermes:"index,searchable,editable,many2one"`
	Rate_Amount float64 `json:"rate_amount" hermes:"searchable,editable"`
}
type RateCollection struct {
	*hermes.Collection
}

func NewRateCollection() (*RateCollection, error) {
	col, err := hermes.NewDBCollection(&Rate{}, application.DataSrc)
	if err != nil {
		return nil, err
	}
	_type := reflect.TypeOf(Rate{})
	rColl := &RateCollection{col}
	hermes.CollectionsMap[_type] = rColl
	return &RateCollection{col}, nil
}
func (col *RateCollection) CreateRate(token string, rater_id int, rated_id int, rate_amount float64) (int, error) {
	var id int
	cnf := col.Conf()
	var err error
	if !hermes.Authorize(token, cnf.Authorizations.Create, rater_id, "CREATE", cnf.CheckAccess) {
		err = givePermission(&token, rated_id)
		if err != nil {
			return -1, err
		}
	}
	query := fmt.Sprintf("INSERT INTO rates(rater_id,rated_id,rate_amount) VALUES(%d,%d,%f) returning id", rater_id, rated_id, rate_amount)
	application.DataSrc.DB.QueryRow(query).Scan(&id)

	return id, nil
}
func (col *RateCollection)GetRate(token string,id int)(Rate,error){
	var err error
	var rate Rate
	rater_id,err:=getUserIdByToken(token)
	if err!=nil {
		return Rate{},err
	}
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Read, rater_id, "GET", cnf.CheckAccess) {
		err = givePermission(&token, rater_id)
		if err != nil {
			return Rate{}, err
		}
	}
	rate=Rate{}
	query:=fmt.Sprintf("SELECT * FROM rates WHERE id= %d",id)
	err=application.DataSrc.DB.Get(&rate,query)
	if err!=nil {
		return Rate{},err
	}
	return rate,nil
}
/*
func (col *RateCollection)GetRatesOfUser(token string,user_id int)(int,error){
	var err error
	rater_id,err:=getUserIdByToken(token)
	if err!=nil {
		return -1,err
	}
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Read, rater_id, "GET", cnf.CheckAccess) {
		err = givePermission(&token, rater_id)
		if err != nil {
			return -1, err
		}
	}
	var rate_average int
	simple_rate:= struct {
		sum,count int
	}{}
	query:=fmt.Sprintf("SELECT SUM(amount) , COUNT(amount) FROM rates WHERE rated_id=%d",user_id)
	err=application.DataSrc.DB.Select(&simple_rate,query)
	if err!=nil {
		return -1,err
	}

	rate_average=int(simple_rate.sum/simple_rate.count)
	return rate_average,nil
}*/
