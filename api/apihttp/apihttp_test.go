// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

package apihttp

import (
	"bytes"
	"encoding/json"
	// "io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"verifiabledata/balloon"
	"verifiabledata/balloon/history"
)

type fakeBalloon struct {
	addch  chan *balloon.Commitment
	stopch chan bool
	proof  chan *balloon.MembershipProof
}

func (b fakeBalloon) Add(event []byte) chan *balloon.Commitment {
	return b.addch
}

func (b fakeBalloon) Close() chan bool {
	return b.stopch
}

func (b fakeBalloon) GenMembershipProof(event []byte, version uint) chan *balloon.MembershipProof {
	return b.proof
}

func newAddOpFakeBalloon(addch chan *balloon.Commitment) fakeBalloon {
	closech := make(chan bool)
	proofch := make(chan *balloon.MembershipProof)

	return fakeBalloon{
		addch,
		closech,
		proofch,
	}
}

func newMembershipOpFakeBalloon(proofch chan *balloon.MembershipProof) fakeBalloon {
	addch := make(chan *balloon.Commitment)
	closech := make(chan bool)

	return fakeBalloon{
		addch,
		closech,
		proofch,
	}
}

func TestHealthCheckHandler(t *testing.T) {
	// Create a request to pass to our handler. We don't have any query parameters for now, so we'll
	// pass 'nil' as the third parameter.
	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the response body is what we expect.
	expected := `{"version":0,"status":"ok"}`
	if rr.Body.String() != expected {
		t.Errorf("handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestInsertEvent(t *testing.T) {
	// Create a request to pass to our handler. We pass a message as a data.
	// If it's nil it will fail.
	data := []byte(`{"message": "this is a sample event"}`)

	req, err := http.NewRequest("POST", "/events", bytes.NewBuffer(data))
	if len(data) == 0 {
		t.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	addch := make(chan *balloon.Commitment)
	handler := InsertEvent(newAddOpFakeBalloon(addch))

	go func() {
		addch <- &balloon.Commitment{
			[]byte{0x0},
			[]byte{0x1},
			0,
		}
	}()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusCreated)
	}

	// Check the body response
	snapshot := &Snapshot{}
	json.Unmarshal([]byte(rr.Body.String()), snapshot)

	if snapshot.HyperDigest != "0000000000000000000000000000000000000000000000000000000000000001" {
		t.Errorf("HyperDigest is not consistent: %s", snapshot.HyperDigest)
	}

	if snapshot.HistoryDigest != "0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("HistoryDigest is not consistent %s", snapshot.HistoryDigest)
	}

	if snapshot.Version != 0 {
		t.Errorf("Version is not consistent")
	}
	if snapshot.Event != "this is a sample event" {
		t.Errorf("Event is not consistent ")
	}

}

func TestMembership(t *testing.T) {
	// Create a request to pass to our handler. We pass a parameters as query params.
	// If it's nil it will fail.
	req, err := http.NewRequest("GET", "/proof/membership", nil)
	if err != nil {
		t.Fatal(err)
	}

	q := req.URL.Query()

	key := "this is a sample event"
	q.Set("key", key)

	version := uint(1)
	q.Set("version", "1")
	req.URL.RawQuery = q.Encode()

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	proof := make(chan *balloon.MembershipProof)
	handler := Membership(newMembershipOpFakeBalloon(proof))

	go func() {
		proof <- &balloon.MembershipProof{
			true,
			[][]byte{[]byte{0x0}},
			[]history.Node{history.Node{[]byte{0x0}, 0, 1}},
			version,
			version + 1,
		}
	}()

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}

	// Check the body response
	membershipProof := &MembershipProof{}
	json.Unmarshal([]byte(rr.Body.String()), membershipProof)

	if membershipProof.Key != key {
		t.Errorf("Key is not consistent ")
	}

	if membershipProof.IsMember != true {
		t.Errorf("IsMember is not consistent ")
	}

	if len(membershipProof.Proofs.HyperAuditPath) != 1 {
		t.Errorf("Proofs.HyperAuditPath is not consistent ")
	}

	if membershipProof.Proofs.HyperAuditPath[0] !=
		"0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Proofs.HyperAuditPath is not consistent ")
	}

	if len(membershipProof.Proofs.HistoryAuditPath) != 1 {
		t.Errorf("Proofs.HistoryAuditPath is not consistent ")
	}

	if membershipProof.Proofs.HistoryAuditPath[0].Digest !=
		"0000000000000000000000000000000000000000000000000000000000000000" {
		t.Errorf("Proofs.HistoryAuditPath is not consistent ")
	}

	if membershipProof.Proofs.HistoryAuditPath[0].Index != 0 {
		t.Errorf("Proofs.HistoryAuditPath is not consistent ")
	}

	if membershipProof.Proofs.HistoryAuditPath[0].Layer != 1 {
		t.Errorf("Proofs.HistoryAuditPath is not consistent ")
	}

	if membershipProof.QueryVersion != version {
		t.Errorf("QueryVersion is not consistent ")
	}

	if membershipProof.ActualVersion != version+1 {
		t.Errorf("ActualVersion is not consistent ")
	}

}

func TestAuthHandlerMiddleware(t *testing.T) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", "this-is-my-api-key")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := AuthHandlerMiddleware(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Check the status code is what we expect.
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("handler returned wrong status code: got %v want %v",
			status, http.StatusOK)
	}
}

func BenchmarkNoAuth(b *testing.B) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		b.Fatal(err)
	}

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Define our http client
	client := &http.Client{}

	for i := 0; i < b.N; i++ {
		client.Do(req)
	}
}

func BenchmarkAuth(b *testing.B) {

	req, err := http.NewRequest("GET", "/health-check", nil)
	if err != nil {
		b.Fatal(err)
	}

	// Set Api-Key header
	req.Header.Set("Api-Key", "this-is-my-api-key")

	// We create a ResponseRecorder (which satisfies http.ResponseWriter) to record the response.
	rr := httptest.NewRecorder()
	handler := AuthHandlerMiddleware(HealthCheckHandler)

	// Our handlers satisfy http.Handler, so we can call their ServeHTTP method
	// directly and pass in our Request and ResponseRecorder.
	handler.ServeHTTP(rr, req)

	// Define our http client
	client := &http.Client{}

	for i := 0; i < b.N; i++ {
		client.Do(req)
	}
}
