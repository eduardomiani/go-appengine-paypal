package paypal

import (
  "time"
)

type CurrencyCode string
var (
  USD CurrencyCode = "USD"
)

func (cc CurrencyCode) String() string {
  switch cc {
    case USD:
    return "USD"
    default:
    return ""
  }
}

//ExpressCheckout Represents all basic informations to set a express checkout on PayPal
type ExpressCheckout struct {
  Amount           float64
  CurrencyCode     CurrencyCode
  ReturnURL        string
  CancelURL        string
  BillingAgreement *BillingAgreement
}

type BillingAgreement struct {
  //TODO Move Type property to a seperated type
  Type        string
  Description string
}

//RecurringPayment Represents all the required and optional informations 
//to create a RecurringPayment profile on PayPal
type RecurringPayment struct {
  Email        string
  PayerID      string
  StartDate    time.Time
  Description  string
  CurrencyCode CurrencyCode
  Billing      *RecurringPaymentBilling
}

//RecurringPaymentBilling Have the informations of recurring payment billing, such as Period,
//Frequency and Installment value
type RecurringPaymentBilling struct {
  Period            string
  Frequency         int
  AmountInstallment float64
  AutoBill          bool
}