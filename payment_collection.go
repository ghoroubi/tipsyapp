package main

import (
	"time"

	"github.com/6thplaneta/hermes"
	"reflect"
	"fmt"
)

/////////////////////////////////////////////////////////////////////////
type Payment struct {
	Id            int       `json:"id" hermes:"dbspace:payments"`
	Creation_Date time.Time `json:"creation_date" hermes:"type:date default:$now"`
	Payer_Id      int       `json:"payer_id" validate:"required" hermes:"index" `
	Payed_Id      int       `json:"payed_id" validate:"required" hermes:"index"`
	Payer         User      `db:"-" json:"payer" hermes:"many2one:User"`
	Payed         User      `db:"-" json:"payed,omitempty" validate:"structonly" hermes:"many2one:User"`
	Amount        float32   `json:"amount,omitempty" hermes:"index,editable"`
	Status        string    `json:"status,omitempty" hermes:"index,editable,searchable"`
	Reference     string    `json:"reference,omitempty" hermes:"index,searchable,editable"`
	// Tip -> 1
	Pay_Type string `json:"pay_type" hermes:"searchable,editable,default:'Tip'"`
	// ApplePay -> 1 AndroidPay -> 2
	AdyenInfo AdyenRequest `db:"-" json:"adyen_info"`
}

/////////////////////////////////////////////////////////////////////////
//  collection
func NewPaymentCollection() (*PaymentCollection, error) {
	coll, err := hermes.NewDBCollection(&Payment{}, application.DataSrc)
	typ := reflect.TypeOf(Payment{})
	OColl := &PaymentCollection{coll}
	hermes.CollectionsMap[typ] = OColl
	return &PaymentCollection{coll}, err
}

type PaymentCollection struct {
	*hermes.Collection
}

func (col *PaymentCollection) List(token string, params *hermes.Params, pg *hermes.Paging, populate, project string) (interface{}, error) {

	obj, err := PaymentColl.Collection.List(token, params, pg, populate, project)
	if err != nil {
		return nil, err
	}
	return obj, nil
}
func (col *PaymentCollection) CreatePayment(token string, payer_id, payed_id int, amount float32, reference string) (int, error) {
	var last_id int
	paymentType := "Tip"
	paymentStatus := "Pending"
	Logger.Info("Payments", "New Payment Transaction Started;"+reference)
	var err error
	cnf := col.Conf()
	if !hermes.Authorize(token, cnf.Authorizations.Create, payer_id, "CREATE", cnf.CheckAccess) {
		err = givePermission(&token, payer_id)
		if err != nil {
			return 0, err
		}
	}
	query1 := "INSERT INTO payments(creation_date,payer_id,payed_id,amount,pay_type,reference,status)"
	query2 := fmt.Sprintf("VALUES(Now(),%d,%d,%f,'%s','%s','%s')", payer_id, payed_id, amount, paymentType, reference, paymentStatus)
	insertQuery := query1 + query2
	application.DataSrc.DB.QueryRow(insertQuery).Scan(&last_id)
	updatePayer := fmt.Sprintf("UPDATE users SET credit_balance=credit_balance+%f WHERE id=%d", amount, payer_id)
	_, err = application.DataSrc.DB.Exec(updatePayer)
	if err != nil {
		return 0, err
	}
	return last_id, nil
}
func (col *PaymentCollection) VerifyPayment(token, reference string) error {

	cnf := col.Conf()
	payerUserId, err := getUserIdByToken(token)
	if err != nil {
		return err
	}
	if !hermes.Authorize(token, cnf.Authorizations.Create, payerUserId, "CREATE", cnf.CheckAccess) {
		err = givePermission(&token, payerUserId)
		if err != nil {
			return err
		}
	}
	paymentUpdateQuery := fmt.Sprintf("UPDATE payments SET status='Verified' WHERE reference='%s'", reference)
	_, err = application.DataSrc.DB.Exec(paymentUpdateQuery)
	if err != nil {
		return err
	}
	// Getting Payer and Payed Users Id for updateing credit balance for them
	var payments []Payment
	payedUserQuery := fmt.Sprintf("SELECT * FROM payments WHERE reference='%s' limit 1", reference)
	err = application.DataSrc.DB.Select(&payments, payedUserQuery)
	if err != nil {
		return err
	}
	payment := payments[0]
	payed_id, payer_id := payment.Payed_Id, payment.Payer_Id
	//////////////////////////////////////////////////////////////////////////////////
	updatePayed := fmt.Sprintf("UPDATE users SET credit_balance=credit_balance + %f WHERE id=%d", payment.Amount, payed_id)
	_, err = application.DataSrc.DB.Exec(updatePayed)
	if err != nil {
		return err
	}
	//////////////////////////////////////////////////////////////////////////////////
	updatePayer := fmt.Sprintf("UPDATE users SET credit_balance=credit_balance-%f WHERE id=%d", payment.Amount, payer_id)
	_, err = application.DataSrc.DB.Exec(updatePayer)
	if err != nil {
		return err
	}
	return nil
}
