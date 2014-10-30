package paypal

import (
  "fmt"
  "io/ioutil"
  "net/http"
  "net/url"
)

var ErrInvalidIPN = fmt.Errorf("paypal (ipn): Invalid IPN")

// Ipn represents a IPN Message sent by PayPal.
type Ipn struct {
  TxnID         string
  PaymentStatus string
  Values        url.Values
}

// CheckIPN validates a IPN Message sent by PayPal and confirms that
// the application receives the message with success.
func CheckIPN(c *http.Client, r *http.Request, sandbox bool) (*Ipn, error) { 
  if err := r.ParseForm(); err != nil {
    return nil, err
  }
  values := r.PostForm
  endpoint := fmt.Sprintf("%s%s", getEndpoint(sandbox), "?cmd=_notify-validate")
  resp, err := c.PostForm(endpoint, values)
  if err != nil {
    return nil, err
  }
  if resp.StatusCode < 200 || resp.StatusCode > 299 {
    return nil, fmt.Errorf("paypal (ipn): Invalid status returned from paypal %d", resp.StatusCode)
  }
  
  b, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, err
  }
  
  if string(b) != "VERIFIED" {
    return nil, ErrInvalidIPN
  }
  
  ipn := &Ipn{
    r.Form["txn_id"][0],
    r.Form["payment_status"][0],
    r.Form,
  }
  return ipn, nil
}