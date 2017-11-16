package main

import (
	"net/http"
	"github.com/6thplaneta/hermes"
	"github.com/6thplaneta/utils.v1"
	"github.com/zhutik/adyen-api-go"
)

var App *hermes.App
var Logger *log.Logger

type NotifyStruct struct {
	Conversation_Id string `json:"conversation_id"`
	Request_Id      int    `json:"request_id"`
}

func main() {
	InitMessages()
	App = hermes.NewApp("./conf.yml")
	App.InitLogs("")
	App.Logger.Level = 6
	App.Router.Use(hermes.LoggerMiddleware(App.Logger, nil))
	App.Router.Use(hermes.AuthMiddleware([]string{"POST:/tipsy/tlogin",
		"POST:/tipsy/flogin",
		"GET:/tipsy/forgetPassword",
		"POST:/tipsy/users", "GET:/auth/logout",
		"POST:/auth/changePasswordByToken", "GET:/upload/image",
		"GET:/auth/activeUser", "GET:/tipsy/newActivationToken"}))
	uploadPath := App.Conf.GetString("public.upload")
	App.Router.POST("/upload/image", hermes.UploadMiddleware("file", uploadPath, uploadImageMW))
	App.Router.StaticFS("/upload", http.Dir(uploadPath))
	// auth should be first module
	// hermes.DisableAuth()
	App.Mount(hermes.AuthorizationModule, "/auth")
	App.Mount(tipsyMdl, "/tipsy")
	RunScripts()
	InitAdyen()
	App.Run()
}

func init() {
	tipsyMdl = &TipsyMdl{}
	var err error
	Logger, err = log.NewLogger("./logs/", 500, nil)
	if err != nil {
		panic(err.Error())
		return
	}
}
func InitAdyen() {
	Instance = adyen.New(
		adyen.Testing,
		ADYEN_USERNAME,
		ADYEN_PASSWORD,
		ADYEN_CLIENT_TOKEN,
		ADYEN_ACCOUNT,
	)
}
