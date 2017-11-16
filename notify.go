package main

import (
	"bytes"
	"encoding/json"
	apns "github.com/anachronistic/apns"
	"io/ioutil"
	"net/http"
	"strconv"
)

type GCMResult struct {
	Registration_Id string
	Message_Id      string
}

type GCMResponse struct {
	Multicast_Id  int64
	Success       int
	Failure       int
	Canonical_Ids int
	Results       []GCMResult
}

type DeviceInfo struct {
	Cm_Id    string
	Platform string
}

func NotifyUser(userid int, message, content string) error {

	var gcms []DeviceInfo
	err := application.DataSrc.DB.Select(&gcms, "select cm_id,platform from devices where id in (select device_id from agent_tokens where is_expired=false and type='login' and agent_id=(select agent_id from users where id="+strconv.Itoa(userid)+"))")

	if err != nil {
		return err
	}
	settings := App.GetSettings("public")
	for i := 0; i < len(gcms); i++ {
		if gcms[i].Platform == "Android" {
			gcm_api_key, _ := settings["gcm_api_key"].(string)

			var jsonStr = []byte(`{"to": "` + gcms[i].Cm_Id + `","data": {"message":"` + message + `","content":` + content + `}}`)
			req, err := http.NewRequest("POST", "https://fcm.googleapis.com/fcm/send", bytes.NewBuffer(jsonStr))
			req.Header.Set("Authorization", gcm_api_key)
			req.Header.Set("Content-Type", "application/json")

			client := &http.Client{}
			res, err := client.Do(req)

			if err != nil {
				App.Logger.Error("GCM Notify Error" + err.Error())

			}

			body, err := ioutil.ReadAll(res.Body)

			var data GCMResponse
			json.Unmarshal(body, &data)
			if data.Canonical_Ids != 0 {
				//update old Registration_Id with new one

				_, err = application.DataSrc.DB.Exec("update devices set cm_id='" + data.Results[0].Registration_Id + "' where cm_id= '" + gcms[i].Cm_Id + "'")

				if err != nil {

					//return err
					application.Logger.Error(err.Error())

				}
			}
		} else if gcms[i].Platform == "iOS" {
			payload := apns.NewPayload()
			payload.Alert = message
			payload.Badge = 1
			payload.Sound = "bingbong.aiff"

			pn := apns.NewPushNotification()
			pn.DeviceToken = gcms[i].Cm_Id

			pn.AddPayload(payload)
			pn.Set("body", content)
			mode, _ := settings["mode"].(string)

			//developmnt
			var client *apns.Client
			if mode == "development" {
				client = apns.NewClient("gateway.sandbox.push.apple.com:2195", "certificate.pem", "key.unencrypted.pem")
			} else if mode == "release" {
				client = apns.NewClient("gateway.push.apple.com:2195", "certificate_r.pem", "key.unencrypted_r.pem")
			}

			resp := client.Send(pn)
			if resp.Error != nil {

				application.Logger.Error(resp.Error.Error())
				//return resp.Error
			}

		}
	}
	return nil
}
