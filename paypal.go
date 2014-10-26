package paypal

import (
	"net/http"
	"net/url"
	"fmt"
	"io/ioutil"
	"strings"
  "time"
)

const (
	NVP_SANDBOX_URL         = "https://api-3t.sandbox.paypal.com/nvp"
	NVP_PRODUCTION_URL      = "https://api-3t.paypal.com/nvp"
  CHECKOUT_SANDBOX_URL    = "https://www.sandbox.paypal.com/cgi-bin/webscr"
  CHECKOUT_PRODUCTION_URL = "https://www.paypal.com/cgi-bin/webscr"
	NVP_VERSION             = "84"
)

type PayPalClient struct {
	username      string
	password      string
	signature     string
	usesSandbox   bool
	client        *http.Client
}

type PayPalResponse struct {
	Ack           string
	CorrelationId string
	Timestamp     string
	Version       string
	Build         string
	Values        url.Values
  usedSandbox   bool
}

type PayPalError struct {
	Ack           string
	ErrorCode     string
	ShortMessage  string
	LongMessage   string
	SeverityCode  string
}

func (e *PayPalError) Error() string {
	var message string
	if len(e.ErrorCode) != 0 && len(e.ShortMessage) != 0 {
		message = "PayPal Error " + e.ErrorCode + ": " + e.ShortMessage
	} else if len(e.Ack) != 0 {
		message = e.Ack
	} else {
		message = "PayPal is undergoing maintenance.\nPlease try again later."
	}

  return message
}

// CheckoutUrl creates the checkout url given a PayPalResponse after call SetExpressCheckout
func (r *PayPalResponse) CheckoutUrl() string {
  query := url.Values{}
  query.Set("cmd", "_express-checkout")
  query.Add("token", r.Values["TOKEN"][0])
  checkoutUrl := CHECKOUT_PRODUCTION_URL
  if r.usedSandbox {
    checkoutUrl = CHECKOUT_SANDBOX_URL
  }
  return fmt.Sprintf("%s?%s", checkoutUrl, query.Encode())
}

// NewDefaultClient Creates a new PayPalClient for comunication with PayPal.
// This function uses the standard http.Client of go
func NewDefaultClient(username, password, signature string, usesSandbox bool) *PayPalClient {
	return &PayPalClient{username, password, signature, usesSandbox, new(http.Client)}
}

// NewClient Creates a new PayPalClient for comunication with PayPal.
// Can receive a different implementation of http.Client for comunication.
func NewClient(username, password, signature string, usesSandbox bool, client *http.Client) *PayPalClient {
	return &PayPalClient{username, password, signature, usesSandbox, client}
}

// PerformRequest performs the request given url values
func (c *PayPalClient) PerformRequest(values url.Values) (*PayPalResponse, error) {
	values.Add("USER", c.username);
	values.Add("PWD", c.password);
	values.Add("SIGNATURE", c.signature);
	values.Add("VERSION", NVP_VERSION);

	endpoint := NVP_PRODUCTION_URL
	if c.usesSandbox {
		endpoint = NVP_SANDBOX_URL
	}
  
	formResponse, err := c.client.PostForm(endpoint, values)
	if err != nil { return nil, err }
	defer formResponse.Body.Close()

	body, err := ioutil.ReadAll(formResponse.Body)
	if err != nil { return nil, err }

	responseValues, err := url.ParseQuery(string(body))
	response := &PayPalResponse{usedSandbox: c.usesSandbox}
	if err == nil {
		response.Ack = responseValues.Get("ACK")
		response.CorrelationId = responseValues.Get("CORRELATIONID")
		response.Timestamp = responseValues.Get("TIMESTAMP")
		response.Version = responseValues.Get("VERSION")
		response.Build = responseValues.Get("2975009")
		response.Values = responseValues

		errorCode := responseValues.Get("L_ERRORCODE0")
		if len(errorCode) != 0 || strings.ToLower(response.Ack) == "failure" || strings.ToLower(response.Ack) == "failurewithwarning" {
			pError := new(PayPalError)
			pError.Ack = response.Ack
			pError.ErrorCode = errorCode
			pError.ShortMessage = responseValues.Get("L_SHORTMESSAGE0")
			pError.LongMessage = responseValues.Get("L_LONGMESSAGE0")
			pError.SeverityCode = responseValues.Get("L_SEVERITYCODE0")

			err = pError
		}
	}

	return response, err
}

// SetExpressCheckout make the ExpressChckout operation, setting the informations of payment
func (client *PayPalClient) SetExpressCheckout(ec *ExpressCheckout) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "SetExpressCheckout")
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", ec.Amount))
  values.Add("PAYMENTREQUEST_0_CURRENCYCODE", ec.CurrencyCode.String());
	values.Add("RETURNURL", ec.ReturnURL);
	values.Add("CANCELURL", ec.CancelURL);
	values.Add("REQCONFIRMSHIPPING", "0");
	values.Add("NOSHIPPING", "1");
	values.Add("SOLUTIONTYPE", "Sole");

  ba := ec.BillingAgreement
  values.Add("L_BILLINGTYPE0", ba.Type)
  values.Add("L_BILLINGAGREEMENTDESCRIPTION0", ba.Description)
  
	return client.PerformRequest(values)
}

// RecurringPaymentProfile creates a RecurringPayment Profile with all the required data.
// This function must be called only after call SetExpressCheckout to get a valid token.
func (client *PayPalClient) RecurringPaymentProfile(token string, rp *RecurringPayment) (*PayPalResponse, error) {
  values := url.Values{}
	values.Set("METHOD", "CreateRecurringPaymentsProfile")
  values.Set("TOKEN", token)
  values.Set("PAYERID", rp.PayerID)
  values.Set("EMAIL", rp.Email)
  //TODO Check if this format date is correct
  values.Set("PROFILESTARTDATE", rp.StartDate.UTC().Format(time.RFC3339))
  values.Set("DESC", rp.Description)
  values.Set("CURRENCYCODE", rp.CurrencyCode.String())
  values.Set("BILLINGPERIOD", rp.Billing.Period)
  values.Set("BILLINGFREQUENCY", fmt.Sprintf("%d", rp.Billing.Frequency))
  values.Set("AMT", fmt.Sprintf("%.2f", rp.Billing.AmountInstallment))
  if rp.Billing.AutoBill {
    values.Set("AUTOBILLOUTAMT", "AddToNextBilling")
  }  
  return client.PerformRequest(values)
}

// paymentType can be "Sale" or "Authorization" or "Order" (ship later)
//TODO Create type param for this function
func (c *PayPalClient) DoExpressCheckout(token, payerId, paymentType, currencyCode string, finalPaymentAmount float64) (*PayPalResponse, error) {
	values := url.Values{}
	values.Set("METHOD", "DoExpressCheckoutPayment")
	values.Add("TOKEN", token)
	values.Add("PAYERID", payerId)
	values.Add("PAYMENTREQUEST_0_PAYMENTACTION", paymentType)
	values.Add("PAYMENTREQUEST_0_CURRENCYCODE", currencyCode);
	values.Add("PAYMENTREQUEST_0_AMT", fmt.Sprintf("%.2f", finalPaymentAmount))

	return c.PerformRequest(values)
}

func (client *PayPalClient) ExpressCheckoutDetails(token string) (*PayPalResponse, error) {
  values := url.Values{}
	values.Add("TOKEN", token)
	values.Set("METHOD", "GetExpressCheckoutDetails")
	return client.PerformRequest(values)
}
