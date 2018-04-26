// Copyright © 2018 Banco Bilbao Vizcaya Argentaria S.A.  All rights reserved.
// Use of this source code is governed by an Apache 2 License
// that can be found in the LICENSE file

/*
	Package agent implements the command line interface to interact with the
	API rest
*/
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"verifiabledata/api/apihttp"
	"verifiabledata/balloon"
	"verifiabledata/balloon/hashing"
	"verifiabledata/balloon/history"
	"verifiabledata/balloon/hyper"
	"verifiabledata/log"
)

type HttpClient struct {
	endpoint string
	apiKey   string
	log      *log.Logger
	http.Client
}

func NewHttpClient(endpoint, apiKey string) *HttpClient {
	return &HttpClient{
		endpoint,
		apiKey,
		log.NewDebug(),
		*http.DefaultClient,
	}
}

func (c HttpClient) doReq(method, path string, data []byte) ([]byte, error) {
	req, err := http.NewRequest(method, c.endpoint+path, bytes.NewBuffer(data))
	if err != nil {
		panic(err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Api-Key", c.apiKey)

	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	bodyBytes, _ := ioutil.ReadAll(resp.Body)

	return bodyBytes, nil
}

func (c HttpClient) Add(event string) (*apihttp.Snapshot, error) {

	data := []byte(fmt.Sprintf("{\"event\": %+q}", event))

	body, err := c.doReq("POST", "/events", data)
	if err != nil {
		return nil, err
	}

	var snapshot apihttp.Snapshot

	json.Unmarshal(body, &snapshot)

	return &snapshot, nil

}

func (c HttpClient) Membership(key []byte, version uint64) *balloon.Proof {

	data := []byte(fmt.Sprintf("{\"key\": %+q, \"version\": %d}", key, version))
	body, err := c.doReq("POST", "/proofs/membership", data)
	if err != nil {
		panic(err)
	}

	var proof apihttp.MembershipProof

	json.Unmarshal(body, &proof)

	return toBalloonProof(&proof)

}

func toBalloonProof(p *apihttp.MembershipProof) *balloon.Proof {
	htlh := history.LeafHasherF(hashing.Sha256Hasher)
	htih := history.InteriorHasherF(hashing.Sha256Hasher)

	hylh := hyper.LeafHasherF(hashing.Sha256Hasher)
	hyih := hyper.InteriorHasherF(hashing.Sha256Hasher)

	historyProof := history.NewProof(apihttp.ToHistoryNode(p.Proofs.HistoryAuditPath), htlh, htih)
	hyperProof := hyper.NewProof("", p.Proofs.HyperAuditPath, hylh, hyih)

	return balloon.NewProof(p.IsMember, hyperProof, historyProof, p.QueryVersion, p.ActualVersion)

}

func (c HttpClient) Verify(event []byte, cm *balloon.Commitment, b *balloon.Proof) bool {
	return b.Verify(cm, event)
}
