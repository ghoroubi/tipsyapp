package main

import (
	"database/sql"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"github.com/6thplaneta/hermes"
	"time"
	"errors"
	"github.com/antonholmquist/jason"
	"net/http"
)

type User struct {
	Id                int          `json:"id" hermes:"index,dbspace:users"`
	Agent_Id          int          `json:"agent_id"`
	Agent             hermes.Agent `db:"-" json:"agent,omitempty" validate:"structonly" hermes:"one2one:User"`
	Registration_Date time.Time    `json:"registration_date" hermes:"type:time"`
	Birth_Date        time.Time    `json:"birth_date" hermes:"type:time,editable,dbtype:date"`
	Modify_Date       time.Time    `json:"modify_date" hermes:"editable,searchable,default:$now,dbtype:date"`
	Average_Rate      float64      `json:"average_rate" hermes:"index,editable,searchable"`
	// Is_Staff <--> true:Who want to take tip false:By default all users is not staff
	Is_Staff          bool    `json:"is_staff" hermes:"index,searchable,editable default=false"`
	Work_Place_id     int     `json:"work_place_id" hermes:"index,searchable,editable"`
	Credit_Balance    float32     `json:"credit_balance" hermes:"index,editable,searchable" `
	First_Name        string  `json:"first_name" hermes:"editable,searchable,index"`
	Last_Name         string  `json:"last_name" hermes:"editable,searchable,index"`
	Display_Name      string  `json:"display_name" hermes:"editable,searchable,index"`
	Bio               string  `json:"about_me" hermes:"editable"`
	Image_Url         string  `json:"image_url" hermes:"editable,searchable"`
	Email             string  `json:"email" validate:"required,email" hermes:"searchable,index"`
	Current_Longitude float64 `json:"current_longitude" hermes:"editable,dbtype:double" `
	Current_Latitude  float64 `json:"current_latitude" hermes:"editable,dbtype:double" `
	Is_Spammed        bool    `json:"is_spammed" hermes:"index,editable,searchable"`
	Is_Deleted        bool    `json:"is_deleted" hermes:"index,editable,searchable"`
	//0 not important //1 male	//2 female
	Gender        int    `json:"gender" hermes:"editable,index"`
	Address       string `json:"address" hermes:"editable,searchable"`
	Mobile_Number string `json:"mobile_number"  hermes:"editable,searchable"`
	Phone_Number  string `json:"phone_number"  hermes:"editable,searchable"`
}

type UserCollection struct {
	*hermes.Collection
}

func NewUserCollection() (*UserCollection, error) {
	coll, err := hermes.NewDBCollection(&User{}, application.DataSrc)

	typ := reflect.TypeOf(User{})
	OColl := &UserCollection{coll}
	hermes.CollectionsMap[typ] = OColl
	return &UserCollection{coll}, err
}

// getCollectionn
func (col *UserCollection) List(token string, params *hermes.Params, pg *hermes.Paging, populate, project string) (interface{}, error) {

	obj, err := UserColl.Collection.List(token, params, pg, populate, project)
	if err != nil {
		return nil, err
	}
	users := *obj.(*[]User)

	for i := 0; i < len(users); i++ {
		if users[i].Is_Staff == true {
			col.UpdateRates(&users[i])
		}
		if users[i].Is_Deleted {
			users[i].Display_Name = "Deleted Account"
			users[i].Image_Url = ""
		}
	}

	return obj, nil
}

func (col *UserCollection) GetFbUser(accessToken string) (*jason.Object, error) {
	response, err := http.Get("https://graph.facebook.com/me?access_token=" + accessToken + "&fields=id,email,name,gender,first_name,last_name,work,about,bio,education")

	if err != nil {
		return nil, err
	}

	if response.StatusCode != 200 {
		return nil, errors.New(Messages["NotFound"])
	}
	str := ReadHttpBody(response)
	user, _ := jason.NewObjectFromBytes([]byte(str))
	return user, nil
}

//override
func (col *UserCollection) Create(token string, trans *sql.Tx, input_user interface{}) (interface{}, error) {

	user := input_user.(*User)
	user.Registration_Date = time.Now()
	user.Agent.Identity = user.Email

	// isExist, err := hermes.AgentColl.ExistsByIdentity(user.Email)

	// if err != nil {
	// 	//general error
	// 	return &User{}, err
	// }
	// if isExist {
	// 	return &User{}, hermes.ErrDuplicate
	// }

	intCnt, err := col.NotDeleteds(user.Email)
	if err != nil {
		return &User{}, err
	}

	if intCnt > 0 {
		return &User{}, hermes.ErrDuplicate
	}
	obj, err := UserColl.Collection.Create(hermes.SystemToken, trans, user)

	if err != nil {
		return &User{}, err
	}

	usr := obj.(*User)

	role_agent := hermes.Role_Agent{}
	role_agent.Agent_Id = usr.Agent_Id
	role_agent.Role_Id = 1
	_, err = hermes.RoleAgentColl.Create(hermes.SystemToken, trans, &role_agent)

	if err != nil {
		return &User{}, err
	}

	var atoken hermes.AgentToken

	err = col.DataSrc.DB.Get(&atoken, " select * from agent_tokens where agent_id="+strconv.Itoa(usr.Agent_Id)+" and type='activation' and is_expired=false order by id desc limit 1 ")
	emailSender := application.GetSettings("email_provider")

	strSubject := Messages["ActiveUserSubject"]
	strMessage := fmt.Sprintf(Messages["ActiveUserBody"], usr.Email, atoken.Token)

	go hermes.SendEmail(strSubject, strMessage, usr.Email, emailSender, true)

	// if err != nil {
	// 	application.Logger.Error(err.Error())
	// }

	return obj, nil
}

func (col *UserCollection) UpdateLocation(token string, id int, longitude, latitude float64) error {

	var err error
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Update, id, "UPDATE", cnf.CheckAccess) {

		err = ownerToken(&token, id)

		if err != nil {
			return err
		}
	}

	rusr, err := UserColl.Collection.Get(hermes.SystemToken, id, "")

	if err != nil {
		return err
	}
	usr := rusr.(*User)

	usr.Current_Latitude = latitude
	usr.Current_Longitude = longitude

	err = UserColl.Collection.Update(token, nil, id, usr)

	if err != nil {
		return err
	}

	return nil
}

//override
func (col *UserCollection) Update(token string, tx *sql.Tx, id int, obj interface{}) error {
	var err error
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Update, id, "UPDATE", cnf.CheckAccess) {

		err = ownerToken(&token, id)
		if err != nil {
			return err
		}
	}
	user := obj.(*User)
	user.Modify_Date = time.Now()

	if user.Is_Staff == true {
		user.Average_Rate, err = col.UpdateRates(user)
	}
	err = UserColl.Collection.Update(token, tx, id, obj)
	if err != nil {
		return err
	}

	return nil
}

func (col *UserCollection) FBLogin(user *User, accessToken string) (*hermes.AgentToken, int, error) {
	fb_user, err := col.GetFbUser(accessToken)
	if err != nil {
		return &hermes.AgentToken{}, 0, err
	}
	email, _ := fb_user.GetString("email")
	name, _ := fb_user.GetString("name")
	fid, _ := fb_user.GetString("id")
	first_name, _ := fb_user.GetString("first_name")
	last_name, _ := fb_user.GetString("last_name")
	g, _ := fb_user.GetString("gender")
	birthday, _ := fb_user.GetString("birthday")
	var gender int
	if g == "female" {
		gender = 2
	} else if g == "male" {
		gender = 1
	}

	exists, err := col.ExistsByEmail(hermes.SystemToken, email)
	if err != nil {
		return &hermes.AgentToken{}, 0, err
	}
	if !exists {
		agent := hermes.Agent{}
		agent.Identity = email
		agent.Is_Active = true
		agent.FId = fid
		tipsyUser := User{}
		tipsyUser.Registration_Date = time.Now()
		tipsyUser.Email = email
		tipsyUser.Agent = agent
		tipsyUser.Display_Name = name
		tipsyUser.First_Name = first_name
		tipsyUser.Last_Name = last_name
		tipsyUser.Gender = gender
		if len(birthday) == 10 {
			arrBith := strings.Split(birthday, "/")
			// tipsyUser.Birth_Date = time.Date(arrBith[2], arrBith[0], arrBith[1], 0, 0, 0, 0, time.UTC)
			year, _ := strconv.Atoi(arrBith[2])
			month, _ := strconv.Atoi(arrBith[0])
			day, _ := strconv.Atoi(arrBith[1])

			tipsyUser.Birth_Date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		}
		url, err := DownloadAndSave("https://graph.facebook.com/" + fid + "/picture?type=large&width=9999&height=9999")
		if err == nil {
			tipsyUser.Image_Url = url.Url
		}

		obj, err := hermes.CreateTrans(hermes.SystemToken, col, &tipsyUser)
		if err != nil {
			return &hermes.AgentToken{}, 0, err

		}

		usr := obj.(*User)

		role_agent := hermes.Role_Agent{}
		role_agent.Agent_Id = usr.Agent_Id
		role_agent.Role_Id = 1
		_, err = hermes.RoleAgentColl.Create(hermes.SystemToken, nil, &role_agent)

		if err != nil {

			return &hermes.AgentToken{}, 0, err
		}

	}

	result, err := hermes.AgentColl.ListQuery("identity="+email+"&is_deleted=false", "")
	if err != nil {
		return &hermes.AgentToken{}, 0, err
	}
	arr := *result.(*[]hermes.Agent)
	lagent := arr[0]

	//update agent if has no fid
	if lagent.FId == "" {
		lagent.FId = fid
		_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agents set fid=%s where id=%d ", fid, lagent.Id))

		if err != nil {
			return &hermes.AgentToken{}, 0, err
		}
	}

	user.Agent.Id = lagent.Id
	user.Agent.Identity = email
	user.Agent.FId = fid
	user.Email = email

	return col.TipsyLogin(user, "Facebook")

}

func (col *UserCollection) NewActivationToken(token string, email string) (hermes.AgentToken, error) {

	var agent hermes.Agent
	err := col.DataSrc.DB.Get(&agent, " select * from agents where identity='"+email+"'")
	if err != nil {
		return hermes.AgentToken{}, err
	}

	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where type='activation' and agent_id=%d ", agent.Id))
	if err != nil {
		return hermes.AgentToken{}, err
	}

	_, err = hermes.AgentTokenColl.CreateToken(hermes.NewToken(agent.Id, "activation"))

	if err != nil {
		return hermes.AgentToken{}, err
	}

	var atoken hermes.AgentToken

	err = col.DataSrc.DB.Get(&atoken, " select * from agent_tokens where agent_id="+strconv.Itoa(agent.Id)+" and type='activation' and is_expired=false order by id desc limit 1 ")

	emailSender := application.GetSettings("email_provider")

	strSubject := Messages["ActiveUserSubject"]
	strMessage := fmt.Sprintf(Messages["ActiveUserBody"], email, atoken.Token)

	go hermes.SendEmail(strSubject, strMessage, email, emailSender, true)

	return hermes.AgentToken{}, nil
}

func (col *UserCollection) NotDeleteds(email string) (int, error) {
	var cnt int
	err := col.DataSrc.DB.Get(&cnt, fmt.Sprintf("select count(*) from users where email='%s' and is_deleted=false", email))
	if err != nil {
		return 0, err

	}
	return cnt, nil
}

func (col *UserCollection) TipsyLogin(user *User, url string) (*hermes.AgentToken, int, error) {
	intCnt, err := col.NotDeleteds(user.Email)
	if err != nil {
		return &hermes.AgentToken{}, 0, err
	}

	if intCnt == 0 {

		return &hermes.AgentToken{}, 0, hermes.ErrNotFound
	}

	user.Agent.Identity = user.Email
	var id int
	agent_token, err := hermes.AgentColl.Login(user.Agent, url)

	if err != nil {
		return &hermes.AgentToken{}, 0, err
	}

	result, err := UserColl.ListQuery("email="+user.Email+"&is_deleted=false", "")

	if err != nil {
		return &hermes.AgentToken{}, 0, err
	}
	userL := *result.(*[]User)
	rUser := userL[0]
	id = rUser.Id

	return &agent_token, id, err

}

func (col *UserCollection) EditProfile(token string, id int, user *User) (*User, error) {
	var err error

	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Update, id, "UPDATE", cnf.CheckAccess) {
		err = ownerToken(&token, id)
		if err != nil {
			return user,err
		}
	}
	//update user
	user.Modify_Date = time.Now()
	err = UserColl.Collection.Update(token, nil, id, user)

	if err != nil {
		return &User{}, err
	}

	//update user.agent by agent_id
	err = hermes.AgentColl.Collection.Update(token, nil, user.Agent_Id, &user.Agent)

	if err != nil {
		return &User{}, err
	}

	return user, nil
}

func (col *UserCollection) ExistsByEmail(token string, email string) (bool, error) {
	users := []User{}
	err := col.DataSrc.DB.Select(&users, fmt.Sprintf("select * from users where lower(email)=lower('%s') and is_deleted=false", email))
	if err != nil {
		return false, err
	}

	if len(users) > 0 {
		return true, nil
	}

	return false, nil
}

// override
func (col *UserCollection) Delete(t *sql.Tx, token string, id int) error {
	var err error
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Delete, id, "DELETE", cnf.CheckAccess) {

		err = ownerToken(&token, id)
		if err != nil {
			return err
		}
	}

	result, err := UserColl.Get(token, id, "")

	user := result.(*User)
	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update users set is_deleted=true where id=%d ; update agents set is_deleted=true where id=%d; ", id, user.Agent_Id))

	if err != nil {
		return err
	}

	_, err = col.DataSrc.DB.Exec(fmt.Sprintf(" update agent_tokens set is_expired=true where agent_id=%d ", user.Agent_Id))
	if err != nil {
		return err
	}
	return nil
}

func (col *UserCollection) SetAverageRate(rated_user_id int) (float64, error) {
	var err error
	var rateAverage, temp float64
	query := fmt.Sprintf("SELECT sum(rate_amount) / count(rate_amount) as average FROM rates WHERE rated_id=%d", rated_user_id)
	err = application.DataSrc.DB.Get(&temp, query)

	if err != nil {
		return 0, err
	}
	rateAverage = Round(temp, 0.5)
	return rateAverage, nil
}
func (col *UserCollection) UpdateRates(user *User) (float64, error) {
	var count int
	var err error
	var average_rate float64
	query := fmt.Sprintf("SELECT count(rate_amount) FROM rates WHERE rated_id=%d", user.Id)
	err = application.DataSrc.DB.Get(&count, query)
	if err != nil {
		return -1, err
	}
	if count > 0 {

		average_rate, err = col.SetAverageRate(user.Id)
		if err != nil {
			return -1, err
		}
	}
	return average_rate, nil
}
