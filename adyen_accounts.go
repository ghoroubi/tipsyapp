package main

import (
	"github.com/6thplaneta/hermes"
	"reflect"
	"io"
	"encoding/xml"
	"fmt"
	"os"
	"github.com/gin-gonic/gin"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"bytes"
)

//////////////// Adyen Accounts Structures ////////////////
var (
	AdyenAccountHolderCreateEndPoint,
	AdyenAccountHolderUpdateEndPoint string
)
//=======================================================//
type AdyenAccountHolder struct {
	Id                    int                  `json:"id" hermes:"dbspace:adyen"`
	AccountHolderCode     string               `db:"unique" json:"accountHolderCode" validate:"required" hermes:"index,searchable"`
	User_Id               int                  `json:"user_id" hermes:"index"`
	User                  User                 `db:"-" json:"user" validate:"structonly" hermes:"one2one:AdyenAccountHolder"`
	AccountHolder_Details AccountHolderDetails `db:"-" json:"accountHolderDetails"`
	LegalEntity           string               `json:"legalEntity" validate:"required"`
	/*Entity type. Allowed values:
	Business
	Individual*/
}
type AccountHolderDetails struct {
	Address            Address_            `db:"-" json:"address,omitempty"`
	BankAccountDetails BankAccountDetail `db:"-" json:"bankAccountDetails,omitempty"`
	BusinessDetails    BusinessDetail      `db:"-" json:"businessDetails,omitempty"`
	Email              string              `json:"email" validate:"required"`
	// (e.g. "0031 6 11 22 33 44", "+316/1122-3344", "(0031) 611223344")
	FullPhoneNumber      string             `json:"fullPhoneNumber,omitempty"`
	IndividualDetails    IndividualDetails_ `db:"-" json:"individualDetails,omitempty"`
	PhoneNumber          PhoneNumber_       `db:"-" json:"phoneNumber,omitempty"`
	MerchantCategoryCode string             `json:"merchantCategoryCode,omitempty"`
	WebAddress           string             `json:"webAddress,omitempty"`
}
type Address_ struct {
	City              string `json:"city" validate:"required"`
	Country           string `json:"country" validate:"required"`
	HouseNumberOrName string `json:"houseNumberOrName" validate:"required"`
	PostalCode        string `json:"postalCode"`
	StateOrProvince   string `json:"stateOrProvince"`
	Street            string `json:"street" validate:"required"`
}
type PhoneNumber_ struct {
	PhoneCountryCode string `json:"phoneCountryCode"`
	PhoneNumber      string `json:"phoneNumber"`
	PhoneType        string `json:"phoneType"`
}
type BankAccountDetail struct {
	AccountNumber          string `json:"accountNumber"`
	AccountType            string `json:"accountType"`
	BankAccountName        string `json:"bankAccountName"`
	BankAccountUUID        string `json:"bankAccountUuid"`
	BankBicSwift           string `json:"bankBicSwift"`
	BankCity               string `json:"bankCity"`
	BankCode               string `json:"bankCode"`
	BankName               string `json:"bankName"`
	BranchCode             string `json:"branchCode"`
	CheckCode              string `json:"checkCode"`
	CountryCode            string `json:"countryCode"`
	CurrencyCode           string `json:"currencyCode"`
	Iban                   string `json:"iban"`
	OwnerCity              string `json:"ownerCity"`
	OwnerCountryCode       string `json:"ownerCountryCode"`
	OwnerDateOfBirth       string `json:"ownerDateOfBirth"`
	OwnerHouseNumberOrName string `json:"ownerHouseNumberOrName"`
	OwnerName              string `json:"ownerName" validate:"required"`
	OwnerNationality       string `json:"ownerNationality"`
	OwnerPostalCode        string `json:"ownerPostalCode"`
	OwnerState             string `json:"ownerState"`
	OwnerStreet            string `json:"ownerStreet"`
	PrimaryAccount         string `json:"primarySccount"`
	TaxId                  string `json:"taxId"`
	UrlForVerification     string `json:"urlForVerification"`
}
type Name_ struct {
	FirstName string `json:"firstName"`
	Gender    string `json:"gender"`
	/*MALE
	FEMALE
	UNKNOWN*/
	Infix    string `json:"infix"`
	LastName string `json:"lastName"`
}
type IndividualDetails_ struct {
	Name         Name_         `db:"-" json:"name"`
	PersonalData PersonalData_ `db:"-" json:"personalData"`
}
type PersonalData_ struct {
	DateOfBirth string `json:"dateOfBirth"`
	// In ISO format yyyy-mm-dd (e.g. 2000-01-31).
	IdNumber    string `json:"idNumber"`
	Nationality string `json:"nationality"`
}
type BusinessDetail struct {
	DoingBusinessAs   string               `json:"doingBusinessAs"`
	LegalBusinessName string               `json:"legalBusinessName"`
	Shareholders      ShareholderContact `db:"-" json:"shareholders"`
	TaxId             string               `json:"taxId"`
}
type ShareholderContact struct {
	ShareholderCode string        `json:"shareholderCode" validate:"required"`
	Name            Name_         `db:"-" json:"name"`
	PersonalData    PersonalData_ `db:"-" json:"personalData"`
	Address         Address_      `db:"-" json:"address"`
	Email           string        `json:"email"`
	WebAddress      string        `json:"webAddress"`
	FullPhoneNumber string        `json:"fullPhoneNumber"`
	PhoneNumber     PhoneNumber_  `db:"-" json:"phoneNumber"`
}

//=======================================================//
///////////////////Collection//////////////////////////////
type AdyenCollection struct {
	*hermes.Collection
}

func NewAdyenCollection() (*AdyenCollection, error) {
	coll, err := hermes.NewDBCollection(&AdyenAccountHolder{}, application.DataSrc)

	typ := reflect.TypeOf(AdyenAccountHolder{})
	OColl := &AdyenCollection{coll}
	hermes.CollectionsMap[typ] = OColl

	return &AdyenCollection{coll}, err
}

type EndPoint struct {
	XMLName xml.Name `xml:"Endpoint"`
	Path    string   `xml:"key,attr"`
	Url     string   `xml:"value,attr"`
}
type EndPoints struct {
	XMLName      xml.Name   `xml:"Endpoints"`
	XMLEndpoints []EndPoint `xml:"Endpoint"`
}
////////// A Simple Struct To Split Tipsy User and Adyen Account //////////

type SimpleAdyen struct {
	AccountHolderCode     string               `db:"unique" json:"accountHolderCode" validate:"required" hermes:"index,searchable"`
	AccountHolder_Details AccountHolderDetails `db:"-" json:"accountHolderDetails"`
	LegalEntity           string               `json:"legalEntity" validate:"required"`
	/*Entity type;
	Allowed values:
	Business
	Individual*/
}
///////////////////////////////////////////////////////////////////////////
func ParseXML(xmlFile io.Reader) (map[string]string, error) {
	var ep EndPoints
	Urls := make(map[string]string)
	err := xml.NewDecoder(xmlFile).Decode(&ep)
	if err != nil {
		return nil, err
	}
	for k, v := range ep.XMLEndpoints {
		Urls[v.Path] = ep.XMLEndpoints[k].Url
	}
	return Urls, nil
}

////////////////////////////////////////////////
func InitEndpoints() {
	xmlFile, _ := os.Open(XMLFileName)
	ep, _ := ParseXML(xmlFile)
	AdyenAccountHolderCreateEndPoint=ep["createAccountHolder"]
	AdyenAccountHolderUpdateEndPoint=ep["Update"]
	fmt.Println("Create:",AdyenAccountHolderCreateEndPoint)
}
func (cont *UserController) CreateAccountHolder(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")
	var err error
	id, err := getUserIdByToken(token)
	if err != nil {
		c.JSON(500, err.Error())
		return
	}
	cnf := UserColl.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Update, id, "UPDATE", cnf.CheckAccess) {
		err := ownerToken(&token, id)
		if err != nil {
			c.JSON(500,err.Error())
			return
		}
	}

	adyenAcc:=AdyenAccountHolder{}
	c.BindJSON(&adyenAcc)
	query:=fmt.Sprintf("UPDATE users SET is_staff=TRUE WHERE id=%d",id)
	_,err=application.DataSrc.DB.Exec(query)
	if err != nil {
		c.JSON(500,err.Error())
	}
	//col.Update(token,nil,user.Id,user)
	simpleAdyen:=SimpleAdyen{
		adyenAcc.AccountHolderCode,
		adyenAcc.AccountHolder_Details,
		adyenAcc.LegalEntity,
		}
	///////////////////////////////////////////////////////////////////////////
	adyenRequestBody,err:=json.Marshal(simpleAdyen)
	if err != nil {
		c.JSON(500,err.Error())
	}
	reader:=bytes.NewReader(adyenRequestBody)
	httpRequest,err:=http.NewRequest("POST",AdyenAccountHolderCreateEndPoint,reader)
	if err != nil {
		c.JSON(500,err.Error())
	}
	httpRequest.Header.Set("content-type","application/json")
	httpRequest.Header.Set("x-api-key",ADYEN_CLIENT_TOKEN)
	client := &http.Client{}
	resp,err:=client.Do(httpRequest)
	if err!=nil {
		fmt.Println(err.Error())
	}
	bytes2,err:=ioutil.ReadAll(resp.Body)
	if err!=nil {
		fmt.Println(err.Error())
	}
	c.Writer.Write(bytes2)
}
