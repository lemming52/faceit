package componenttests

import (
	"encoding/json"
	"faceit/model"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func getHost() string {
	return "http://localhost:3000"
}

func TestGetUser(t *testing.T) {
	id := "testing"
	codeWant := 200
	expected := &model.User{
		Id:       "testing",
		Nickname: "lemming52",
		Country:  "NZ",
		Forename: "andrew",
		Surname:  "s",
		Password: "correcthorsebatterystaple",
		Email:    "lemming52@github.com",
	}

	uri := fmt.Sprintf("%s/users/%s", getHost(), id)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, codeWant, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results := &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)

	if !reflect.DeepEqual(expected, results) {
		t.Errorf("user should match %s %s", expected, results)
	}
}

func TestAddUser(t *testing.T) {
	codeWant := 201
	expected := &model.User{
		Nickname: "shox",
		Country:  "FRA",
		Forename: "Richard",
		Surname:  "Papillion",
		Password: "vitality",
		Email:    "rp@notarealemail.com",
	}
	payload := `{
		"forename": "Richard",
		"surname": "Papillion",
		"nickname": "shox",
		"password": "vitality",
		"email": "rp@notarealemail.com",
		"country": "FRA"
	}`
	// Insert
	uri := fmt.Sprintf("%s/users", getHost())
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload))
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, codeWant, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results := &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)

	expected.Id = results.Id
	if !reflect.DeepEqual(expected, results) {
		t.Errorf("inserteduser should match %s %s", expected, results)
	}

	// Check if stored using Get endpoint
	uri = fmt.Sprintf("%s/users/%s", getHost(), expected.Id)
	req, err = http.NewRequest(http.MethodGet, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, codeWant, res.StatusCode)

	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results = &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)
	if !reflect.DeepEqual(expected, results) {
		t.Errorf("retrieved user should match %s %s", expected, results)
	}

	// Cleanup, implicitly test delete endpoint
	deleteCode := 204
	req, err = http.NewRequest(http.MethodDelete, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making delete request")
	assert.Equal(t, deleteCode, res.StatusCode)
}

func TestDelete(t *testing.T) {
	payload := `{
		"forename": "Gabriel",
		"surname": "Toledo",
		"nickname": "FalleN",
		"password": "MIBR",
		"email": "gt@notarealemail.com",
		"country": "BRA"
	}`

	// Seed, implicitly test insert
	uri := fmt.Sprintf("%s/users", getHost())
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload))
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results := &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)
	id := results.Id

	// Delete
	uri = fmt.Sprintf("%s/users/%s", getHost(), id)
	req, err = http.NewRequest(http.MethodDelete, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making request")

	// Check Deleted
	notFoundCode := 404
	req, err = http.NewRequest(http.MethodGet, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, notFoundCode, res.StatusCode)
}

func TestUpdate(t *testing.T) {
	payload := `{
		"forename": "Jacky",
		"surname": "Yip",
		"nickname": "Stewie2K",
		"password": "MIBR",
		"email": "jy@notarealemail.com",
		"country": "USA"
	}`
	updatePayload := `{
		"forename": "Jacky",
		"surname": "Yip",
		"nickname": "Stewie2K",
		"password": "Liquid",
		"email": "jy@notarealemail.com",
		"country": "USA"
	}`
	expected := &model.User{
		Forename: "Jacky",
		Surname:  "Yip",
		Nickname: "Stewie2K",
		Password: "Liquid",
		Email:    "jy@notarealemail.com",
		Country:  "USA",
	}

	// Seed, implicitly test insert
	uri := fmt.Sprintf("%s/users", getHost())
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload))
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results := &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)
	expected.Id = results.Id

	// Update
	expectedCode := 200
	uri = fmt.Sprintf("%s/users/%s", getHost(), expected.Id)
	req, err = http.NewRequest(http.MethodPut, uri, strings.NewReader(updatePayload))
	res, err = client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, expectedCode, res.StatusCode)

	// Check Updated
	req, err = http.NewRequest(http.MethodGet, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, expectedCode, res.StatusCode)

	body, err = ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	results = &model.User{}
	err = json.Unmarshal(body, results)
	assert.Nil(t, err)

	if !reflect.DeepEqual(expected, results) {
		t.Errorf("user should match %s %s", expected, results)
	}

	// Cleanup, implicitly test delete endpoint
	deleteCode := 204
	req, err = http.NewRequest(http.MethodDelete, uri, nil)
	res, err = client.Do(req)
	assert.Nil(t, err, "error making delete request")
	assert.Equal(t, deleteCode, res.StatusCode)
}

func TestFilter(t *testing.T) {
	expected := []*model.User{
		{
			Id:       "0144a93a-c655-49f9-8a86-57533a083333",
			Forename: "Nathan",
			Surname:  "Schmitt",
			Nickname: "NBK-",
			Password: "og",
			Email:    "ns@notarealemail.com",
			Country:  "FRA",
		},
	}
	expectedCount := 1
	expectedCode := 200
	uri := fmt.Sprintf("%s/users", getHost())
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	q := req.URL.Query()
	q.Add("country", "FRA")
	req.URL.RawQuery = q.Encode()
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, expectedCode, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	response := &model.FilterResponse{}
	err = json.Unmarshal(body, response)
	assert.Nil(t, err)

	assert.Equal(t, expectedCount, response.Count)
	if !reflect.DeepEqual(expected, response.Results) {
		t.Errorf("filtered results should match %v %v", expected[0], response.Results[0])
	}
}

// Not strictly a requirement, but functions as a test to check if any documents have escaped cleanup
func TestGetAll(t *testing.T) {
	expectedCount := 5
	expectedCode := 200
	uri := fmt.Sprintf("%s/users", getHost())
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	client := &http.Client{}
	res, err := client.Do(req)
	assert.Nil(t, err, "error making request")
	assert.Equal(t, expectedCode, res.StatusCode)

	body, err := ioutil.ReadAll(res.Body)
	assert.Nil(t, err)
	response := &model.FilterResponse{}
	err = json.Unmarshal(body, response)
	assert.Nil(t, err)
	assert.Equal(t, expectedCount, response.Count)
}
