package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"
)

type testValue struct {
	Password string
	Hashed   string
}

var testValues = [...]testValue{
	{"angryMonkey",
		"ZEHhWB65gUlzdVwtDQArEyx+KVLzp/aTaRaPlBzYRIFj6vjFdqEb0Q5B8zVKCZ0vKbZPZklJz0Fd7su2A+gf7Q=="},
	{"anussan",
		"rNcCtJrJyZxKXPcj6jq/euxudiKkDihJMeoPzi7Ar1+NQXz+1Sg9B4EWpIyXWGK0dxrOe1PQw3uYGNlVEYgeQw=="},
	{"testserver",
		"3+PI3hk8rOJBMeUNaKiVbUbzZRsMLDUEhKfb2una4eT+HZsuRnDAelq6O/YNelqYhh5pxusX51YNLyOsdwC7Dg=="},
}

func TestServer(t *testing.T) {
	server := &http.Server{Addr: fmt.Sprintf(":%d", 80)}
	done := make(chan struct{})
	go func() {
		server.SetKeepAlivesEnabled(false)

		close(done)
	}()

	for i, val := range testValues {
		// Generate POST request
		data := url.Values{}
		data.Add("password", val.Password)
		post, _ := http.NewRequest("POST", "/hash", bytes.NewBufferString(data.Encode()))
		post.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		postRes := httptest.NewRecorder()
		handleHashRequest_rootOnly(postRes, post)
		res := strings.Replace(postRes.Body.String(), "\n", "", -1)
		if res != strconv.Itoa(i+1) {
			t.Errorf("Unexpected return value: %s", postRes.Body.String())

		}
		//Generate the get Request

		time.Sleep(time.Second * 5)
		getReq, _ := http.NewRequest("GET", fmt.Sprintf("/hash/%d", i+1), nil)
		getRes := httptest.NewRecorder()
		makeFunc_handleHashRequest(getRes, getReq)
		getbodyres := strings.Replace(getRes.Body.String(), "\n", "", -1)
		if getbodyres != val.Hashed {
			t.Errorf("Unexpected hashed string for password %s. Expected %s, Got %s", val.Password, val.Hashed, getRes.Body.String())
		}

	}

	// Checking the stats
	stat, _ := http.NewRequest("GET", "/stats", nil)
	statRes := httptest.NewRecorder()
	handleStatsRequest(statRes, stat)
	type stats struct {
		Total   int     "total"
		Average float32 "average"
	}
	var st stats
	err := json.Unmarshal(statRes.Body.Bytes(), &st)
	if err != nil {
		t.Errorf("Could not parse stats json. Err: %v", err)
	}
	if st.Total != int(len(testValues)) {
		t.Errorf("Unexpected total stat: %d", st.Total)
	}

	if st.Average > 1000 {
		t.Errorf("Stats average time too high: %f", st.Average)
	}

	<-done

}
