package paypal

import (
  "bytes"
  "testing" 
  "net/http"
  "net/http/httptest"
)

func TestCheckValidIPN(t *testing.T) {
  c := new(http.Client)
  s := paypalIpnTestServer()
  defer s.Close()
  getEndpoint = func(sandbox bool) string {
    return s.URL
  }
  
  r, err := http.NewRequest("POST", "http://localhost", bytes.NewBufferString(IpnExample))
  if err != nil {
    t.Fatal(err)
  }
  r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
  
  ipn, err := CheckIPN(c, r, true)
  if err != nil {
    t.Errorf(err.Error())
  }
  if ipn == nil {
    t.Fatalf("A non-nil ipn is expected")
  }
  if ipn.TxnID != "61E67681CH3238416" {
    t.Errorf("Unexpected txn_id %s, expected 61E67681CH3238416", ipn.TxnID)
  }
}

func paypalIpnTestServer() *httptest.Server {
  return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("VERIFIED"))
	}))
}