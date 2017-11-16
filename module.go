package main

import (
	"io/ioutil"
	"github.com/6thplaneta/hermes"
)

var tipsyMdl *TipsyMdl

type TipsyMdl struct {
	hermes.Module
}

var FileColl *FileCollection
var UserColl *UserCollection
var PaymentColl *PaymentCollection
var RateColl *RateCollection
var BarColl *BarCollection
var AdyenColl *AdyenCollection
var application *hermes.App
var SecretKey string
var XMLFileName string

func (sm *TipsyMdl) Init(app *hermes.App) error {
	var err error

	settings := app.GetSettings("agents")

	if settings["secret"] == nil {
		return ErrConfigNotFound
	}
	SecretKey = settings["secret"].(string)
	XMLFileName = "./AccountEndpoints.xml"
	InitEndpoints()
	application = app
	// AdyenCollection
	AdyenColl, err = NewAdyenCollection()
	if err != nil {
		return err
	}
	AdyenColl.Conf().SetAuth("Create AdyenAccountHolder",
		"Get AdyenAccountHolder",
		"List AdyenAccountHolder",
		"Update AdyenAccountHolder",
		"",
		"")
	// UserCollection
	UserColl, err = NewUserCollection()
	if err != nil {
		return err
	}
	UserColl.Conf().SetAuth("Create User", "Get User", "List User", "Update User", "Delete User", "Update User")
	// FileCollection
	FileColl, err = NewFileCollection()
	if err != nil {
		return err
	}
	FileColl.Conf().SetAuth("Create File", "Get File", "List File", "", "Delete File", "Rel File")

	//PaymentCollection
	PaymentColl, err = NewPaymentCollection()
	if err != nil {
		return err
	}
	PaymentColl.Conf().SetAuth("Create Payment", "Get Payment", "List Payment", "Update Payment", "Delete Payment", "")
	// NeedCollection

	// Bar Collection
	BarColl, err = NewBarCollection()
	if err != nil {
		return err
	}
	BarColl.Conf().SetAuth("Create Bar", "Get Bar", "List Bar", "Update Bar", "Delete Bar", "Rel Bar")

	// RateCollection
	RateColl, err = NewRateCollection()
	if err != nil {
		return err
	}
	//RateColl.Conf().SetAuth("Create Rate","Get Rate","List Rate","Update Rate","Delete Rate","Rel Rate")
	// Adding DB required FKeys

	err = hermes.AddPostgresForeignKey(
		application.DataSrc.DB,
		User{},
		"agent_id",
		hermes.Agent{},
		"id",
		false,
	)
	if err != nil {
		app.Logger.Error("Error in adding key:" + err.Error())
		return err
	}
	err = hermes.AddPostgresForeignKey(
		application.DataSrc.DB,
		Payment{},
		"payer_id",
		User{},
		"id",
		false,
	)
	if err != nil {
		app.Logger.Error("Error in adding key:" + err.Error())
		return err
	}
	err = hermes.AddPostgresForeignKey(
		application.DataSrc.DB,
		Payment{},
		"payed_id",
		User{},
		"id",
		false,
	)
	if err != nil {
		app.Logger.Error("Error in adding key:" + err.Error())
		return err
	}
	err = hermes.AddPostgresForeignKey(
		application.DataSrc.DB,
		AdyenAccountHolder{},
		"user_id",
		User{},
		"id",
		false,
	)
	if err != nil {
		app.Logger.Error("Error in adding key:" + err.Error())
		return err
	}
	SetRouter()
	addRolePermissions()
	return nil
}

func SetRouter() {

	// Users
	userCont := NewUserController(UserColl, "/users")
	tipsyMdl.GET("/users", userCont.List)
	tipsyMdl.GET("/users/items/:id", userCont.Get)
	tipsyMdl.PUT("/users/items/:id", userCont.Update)
	tipsyMdl.POST("/users", userCont.Create)
	tipsyMdl.DELETE("/users/items/:id", userCont.Delete)
	tipsyMdl.GET("/existsUser/items/:email", userCont.ExistsByEmail)
	tipsyMdl.GET("/forgetPassword/:identity", userCont.RequestPasswordToken)
	tipsyMdl.GET("/newActivationToken/:email", userCont.NewActivationToken)
	tipsyMdl.POST("/createAccountHolder",userCont.CreateAccountHolder)

	tipsyMdl.PUT("/editProfile/items/:id", userCont.EditProfile)
	tipsyMdl.POST("/tlogin", userCont.TipsyLogin)
	tipsyMdl.POST("/flogin/:ftoken", userCont.FBLogin)
	// Rates
	rateController := NewRateController(RateColl, "/rates")
	tipsyMdl.POST("/rates", rateController.CreateRate)
	tipsyMdl.GET("/rates/items/:id", rateController.GetRate)
	// Bars
	barCont := NewBarController(BarColl, "/bars")
	tipsyMdl.RegisterController(barCont)
	tipsyMdl.SetCrudRoutes(barCont, nil)
	tipsyMdl.POST("/newBar", barCont.NewBar)
	tipsyMdl.GET("/getStaffs", barCont.GetStaffs)
	tipsyMdl.GET("/getBarId", barCont.GetBarId)
	//Payments
	paymentCont := NewPaymentController(PaymentColl, "/payments")
	tipsyMdl.RegisterController(paymentCont)
	tipsyMdl.SetCrudRoutes(paymentCont, nil)
	tipsyMdl.POST("/setupPayment", paymentCont.SetupPayment)
	tipsyMdl.POST("/verifyPayment", paymentCont.VerifyPayment)

}
func addRolePermissions() {
	usrRoleId, err := hermes.AddRole("User")
	if err != nil {
		panic(err)
	}
	errPrm := hermes.AddRolePermission(usrRoleId,
		"Get User", "Get Bar", "Get Payment", "Get Rate",
		"Update User", "Update Bar", "Update Profile",
	)
	if errPrm != nil {
		panic(errPrm)
	}

	adminId, err := hermes.AddRole("Admin")
	if err != nil {
		panic(err)
	}
	errPrm = hermes.AddRolePermission(adminId,
		"Get User", "Create User", "List User", "Update User",
		"Get Bar", "Create Bar", "List Bar", "Update Bar",
		"Get Rate", "Create Rate", "List Rate", "Delete Rate", "Update Rate",
		"Get Payment", "Create Payment", "List Payment", "Delete Payment", "Update Payment",
	)
	if errPrm != nil {
		panic(errPrm)
	}
}
func RunScripts() {
	//postgres
	dat, err := ioutil.ReadFile("./scripts/postgres")
	if err != nil {
		panic(err)
	}
	psQuery := string(dat)
	_, err = application.DataSrc.DB.Exec(psQuery)

}
