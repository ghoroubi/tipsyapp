package main

import (
	"reflect"
	"strconv"
	"github.com/6thplaneta/hermes"
	"net/http"
	"fmt"
	)

func setCreatorId(token string, obj interface{}, field string) error {
	uId, err := getUserIdByToken(token)
	if err != nil {
		return err
	}
	rval := reflect.ValueOf(obj)
	if rval.Kind() == reflect.Ptr {
		rval = rval.Elem()
	}
	rval.FieldByName(field).Set(reflect.ValueOf(uId))
	return nil
}

func ownerToken(token *string, userId int) error {

	agent, err := hermes.AgentColl.GetByLoginToken(*token)
	if err != nil {
		return err
	}

	ruser, err := UserColl.Get(hermes.SystemToken, userId, "")

	if err != nil {
		return err
	}
	if ruser.(*User).Agent_Id == agent.Id {
		*token = hermes.SystemToken
	}

	return nil
}

func getUserIdByToken(token string) (int, error) {
	agent, err := hermes.AgentColl.GetByLoginToken(token)
	if err != nil {
		return 0, err
	}

	result, err := UserColl.ListQuery("agent_id="+strconv.Itoa(agent.Id), "")

	userL := *result.(*[]User)
	if err != nil {
		return 0, err
	}
	if len(userL) == 0 {
		return 0, hermes.ErrNotFound

	}
	user := userL[0]
	return user.Id, nil
}
func givePermission(token *string, userId int) error {

	agent, err := hermes.AgentColl.GetByLoginToken(*token)

	if err != nil {
		return err
	}

	ruser, err := UserColl.Get(*token, userId, "")

	if err != nil {
		return err
	}
	if ruser.(*User).Agent_Id == agent.Id {
		*token = hermes.SystemToken
	}

	return nil
}
func ReadHttpBody(response *http.Response) string {

	bodyBuffer := make([]byte, 5000)
	var str string

	count, err := response.Body.Read(bodyBuffer)

	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {

		if err != nil {

		}

		str += string(bodyBuffer[:count])
	}

	return str

}
func addAdmin() {
	settings := application.GetSettings("admin")

	email := settings["email"].(string)
	password := settings["password"].(string)

	usr := User{}
	usr.Display_Name = "admin"
	usr.Email = email
	usr.Agent = hermes.Agent{}
	usr.Agent.Identity = email
	usr.Agent.Password = password
	usr.Agent.Is_Active = true

	result, _ := UserColl.Create(hermes.SystemToken, nil, &usr)

	if result != nil {
		admin_user := result.(*User)
		_, err := UserColl.DataSrc.DB.Exec(fmt.Sprintf("update agents set is_active=true where id=%d;update roles_agents set role_id=2 where agent_id=%d;", admin_user.Agent_Id, admin_user.Agent_Id))
		if err != nil {
			application.Logger.Error("Add Admin Err:" + err.Error())
		}
	}
}
func Round(x, unit float64) float64 {
	return float64(int64(x/unit+0.5)) * unit
}
