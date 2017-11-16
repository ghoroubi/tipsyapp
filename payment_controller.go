package main

import (
	"github.com/6thplaneta/hermes"
	"github.com/gin-gonic/gin"
	"net/http"
	"io/ioutil"
	"github.com/zhutik/adyen-api-go"
	"time"
	"encoding/json"
	bytes2 "bytes"
)

const (
	ADYEN_USERNAME     = "ws@Company.DigiPays"
	ADYEN_PASSWORD     = "af>7Fdy&m<6U#fZHPp\\=aPfni"
	ADYEN_CLIENT_TOKEN = "AQEjhmfuXNWTK0Qc+iSUm2Mxl+WCWxAOSglB1/fm53It0bnEX9kQwV1bDb7kfNy1WIxIIkxgBw==-z7yb0PM9m7y54XWz0J7vlvgkwEHnxg7F3xB1/GCj9Qo=-hTg8c8uxRmJ4er33"
	ADYEN_ACCOUNT      = "DigiPaysCOM"
	CHECKOUT_API_KEY   = "AQEjhmfuXNWTK0Qc+iSUm2Mxl+WCWxAOSglB1/fm53It0bnEX9kQwV1bDb7kfNy1WIxIIkxgBw==-z7yb0PM9m7y54XWz0J7vlvgkwEHnxg7F3xB1/GCj9Qo=-hTg8c8uxRmJ4er33"
	SETUP_ENDPOINT ="https://checkout-test.adyen.com/services/PaymentSetupAndVerification/v31/setup"
	VERIFY_ENDPOINT="https://checkout-test.adyen.com/services/PaymentSetupAndVerification/v31/verify"
)

type AdyenRequest struct {
	Amount           adyen.Amount `json:"amount"`
	Channel          string       `json:"channel"`
	CountryCode      string       `json:"countryCode"`
	MerchantAccount  string       `json:"merchantAccount"`
	Reference        string       `json:"reference"`
	ReturnUrl        string       `json:"returnUrl"`
	ShopperLocal     string       `json:"shopperLocal"`
	ShopperReference string       `json:"shopperReference"`
	Token            string       `json:"token"`

}
type AdyenSetupResponse struct {
	AuthoriseFromSDK          string          `json:"authoriseFromSDK"`
	DisableRecurringDetailURL string          `json:"disableRecurringDetailUrl"`
	GenerationtTime           time.Time       `json:"generationtime"`
	HTML                      string          `json:"html"`
	InitiationURL             string          `json:"initiationUrl"`
	LogoBaseURL               string          `json:"logoBaseUrl"`
	Origin                    string          `json:"origin"`
	OriginKey                 string          `json:"originKey"`
	Payment                   AdyenPayment    `json:"payment"`
	PaymentData               string          `json:"paymentData"`
	PaymentMethods            []PaymentMethod `json:"paymentMethods"`
	PublicKey                 string          `json:"publicKey"`
	PublicKeyToken            string          `json:"publicKeyToken"`
}

type AdyenPayment struct {
	Amount           adyen.Amount `json:"amount"`
	CountryCode      string       `json:"countryCode"`
	Reference        string       `json:"reference"`
	SessionValidity  time.Time    `json:"sessionValidity"`
	ShopperLocale    string       `json:"shopperLocale"`
	ShopperReference string       `json:"shopperReference"`
}
type PaymentMethod struct {
	Group             Groups         `json:"group,omitempty"`
	InputDetails      []InputDetail  `json:"inputDetails"`
	Name              string         `json:"name"`
	PaymentMethodData string         `json:"paymentMethodData"`
	Type              string         `json:"type"`
	Configuration     Configurations `json:"configuration,omitempty"`
}
type Configurations struct {
	CanIgnoreCookies string `json:"canIgnoreCookies"`
}
type Groups struct {
	Name              string `json:"name"`
	PaymentMethodData string `json:"paymentMethodData"`
	Type              string `json:"type"`
}
type InputDetail struct {
	Key      string `json:"key"`
	Type     string `json:"type"`
	Optional string `json:"optional,omitempty"`
}

/*
Checkout API Key:
AQEjhmfuXNWTK0Qc+iSUm2Mxl+WCWxAOSglB1/fm53It0bnEX9kQwV1bDb7kfNy1WIxIIkxgBw==-z7yb0PM9m7y54XWz0J7vlvgkwEHnxg7F3xB1/GCj9Qo=-hTg8c8uxRmJ4er33
_______________________________________________________________________

Checkout Demo Server API Key:
0101348667EE5CD5932B441CFA24949B633197E5825BC65908805E15415FC1719B9F7F58BDE85BC7D4C1D2337A6FD275C103EBB4AE134010C15D5B0DBEE47CDCB5588C48224C6007
_______________________________________________________________________
WS Username:
ws@Company.DigiPays

WS User Password :
af>7Fdy&m<6U#fZHPp\=aPfni
_______________________________________________________________________
Skin:
Code: bTiNscCu
HMAC Key_Test:
1D3ED90234D4C357B11A7098B51AF696C9C97EBF4AA6A9D187F7595152189087
Valid Account:
DigiPaysCOM

________________________________________________________________________

*/

type PaymentController struct{
	*hermes.Controller
}

var Instance *adyen.Adyen
func NewPaymentController(coll hermes.Collectionist,base string) *PaymentController{
	return &PaymentController{hermes.NewController(coll,base)}

}
func (cont *PaymentController)List(c *gin.Context){
	if hermes.RedirectPage(c){
		return
	}
	token:=c.Request.Header.Get("Authorization")
	params:=cont.ReadParams(c)
	paging,err:=hermes.ReadPaging(c)
	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
	}
	result, err := cont.Coll.List(token, params, paging, c.Query("$populate"), "")

	if err != nil {
		hermes.HandleHttpError(c, err, application.Logger)
	}
	c.JSON(http.StatusOK, result)

}
func (cont *PaymentController) SetupPayment(c *gin.Context) {
	var err error
	req := &Payment{}
	c.BindJSON(&req)
	adyenRequest:=req.AdyenInfo
	adyenRequestBody,err:=json.Marshal(adyenRequest)
	if err != nil {
		c.JSON(500,err.Error())
	}
	reader:=bytes2.NewReader(adyenRequestBody)
	httpRequest,err:=http.NewRequest("POST",SETUP_ENDPOINT,reader)
	if err != nil {
		c.JSON(500,err.Error())
	}
	httpRequest.Header.Set("content-type","application/json")
	httpRequest.Header.Set("x-api-key",ADYEN_CLIENT_TOKEN)
	client := &http.Client{}
	resp,err:=client.Do(httpRequest)
	if err != nil {
		c.JSON(500,err.Error())
	}
	bytes,err:=ioutil.ReadAll(resp.Body)
	if err != nil {
		c.JSON(500,err.Error())
	}
	c.Writer.Write(bytes)
	PaymentColl.CreatePayment(
		c.Request.Header.Get("Authorization"),
		req.Payer_Id,req.Payed_Id,
		req.AdyenInfo.Amount.Value,
		req.AdyenInfo.Reference,
		)
}
func (cont *PaymentController) VerifyPayment(c *gin.Context) {
	var err error
	var reference string
	token:=c.Request.Header.Get("Authorization")
	//c.BindJSON(&reference)
	reference=c.Query("ref")
	httpRequest,err:=http.NewRequest("POST",VERIFY_ENDPOINT,c.Request.Body)

	if err != nil {
		c.JSON(500,err.Error())
	}
	httpRequest.Header.Set("content-type","application/json")
	httpRequest.Header.Set("x-api-key",ADYEN_CLIENT_TOKEN)
	client := &http.Client{}
	resp,err:=client.Do(httpRequest)
	if err != nil {
		c.JSON(500,err.Error())
	}
	bytes,_:=ioutil.ReadAll(resp.Body)
	c.Writer.Write(bytes)
	PaymentColl.VerifyPayment(token,reference)

}
