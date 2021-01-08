package handlers

import (
	"context"
	"errors"
	"faceit/model"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Mock Clients

type mockDaoClient struct {
	wasCalled  bool
	calledFunc string
	failFunc   string
	payload    *model.User
	results    []*model.User
}

func NewMockDaoClient(payload *model.User, results []*model.User, failFunc string) *mockDaoClient {
	return &mockDaoClient{
		wasCalled:  false,
		calledFunc: "None",
		failFunc:   failFunc,
		payload:    payload,
		results:    results,
	}
}

func (m *mockDaoClient) Get(ctx context.Context, id string) (*model.User, error) {
	m.wasCalled = true
	m.calledFunc = "Get"
	if m.failFunc == "Get" {
		return nil, errors.New("unable to get")
	}
	return m.payload, nil
}

func (m *mockDaoClient) Insert(ctx context.Context, user *model.User) error {
	m.wasCalled = true
	m.calledFunc = "Insert"
	m.payload = user
	if m.failFunc == "Insert" {
		return errors.New("unable to insert")
	}
	return nil
}

func (m *mockDaoClient) Delete(ctx context.Context, userId string) error {
	m.wasCalled = true
	m.calledFunc = "Delete"
	if m.failFunc == "Delete" {
		return errors.New("unable to delete")
	}
	return nil
}

func (m *mockDaoClient) Filter(ctx context.Context, conditions []*model.FilterCondition) ([]*model.User, error) {
	m.wasCalled = true
	m.calledFunc = "Filter"
	if m.failFunc == "Filter" {
		return nil, errors.New("unable to filter")
	}
	return m.results, nil
}

func (m *mockDaoClient) GetAll(ctx context.Context) ([]*model.User, error) {
	m.wasCalled = true
	m.calledFunc = "GetAll"
	if m.failFunc == "GetAll" {
		return nil, errors.New("unable to get all")
	}
	return m.results, nil
}

type mockMsgClient struct {
	wasCalled bool
	fail      bool
}

func NewMockMsgClient(fail bool) *mockMsgClient {
	return &mockMsgClient{
		wasCalled: false,
		fail:      fail,
	}
}

func (m *mockMsgClient) Publish(ctx context.Context, msg *model.Message) error {
	m.wasCalled = true
	if m.fail {
		return errors.New("unable to publish")
	}
	return nil
}

// Handler Tests

func TestGet(t *testing.T) {
	payload := &model.User{
		Id:       "dummy-test-user",
		Forename: "Andreas",
		Surname:  "Hojsleth",
		Nickname: "Xyp9x",
		Password: "astralis",
		Email:    "ah@notarealemail.com",
		Country:  "DEN",
	}
	id := "dummy-test-user"
	expectedCode := 200
	db := NewMockDaoClient(payload, nil, "None")
	msg := NewMockMsgClient(false)
	handler := NewHandler(db, msg)
	uri := fmt.Sprintf("/users/%s", id)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	assert.Nil(t, err)

	code, res, err := handler.GetUser(req)
	assert.Nil(t, err)
	assert.Equal(t, expectedCode, code)
	assert.Equal(t, db.wasCalled, true)
	assert.Equal(t, msg.wasCalled, false)
	if !reflect.DeepEqual(payload, res) {
		t.Errorf("expected response payload %v should match %v", payload, res)
	}
}

func TestGetFail(t *testing.T) {
	id := "dummy-test-user"
	expectedCode := 404
	db := NewMockDaoClient(nil, nil, "Get")
	msg := NewMockMsgClient(false)
	handler := NewHandler(db, msg)
	uri := fmt.Sprintf("/users/%s", id)
	req, err := http.NewRequest(http.MethodGet, uri, nil)
	assert.Nil(t, err)

	code, _, err := handler.GetUser(req)
	assert.NotNil(t, err)
	assert.Equal(t, expectedCode, code)
	assert.Equal(t, db.wasCalled, true)
	assert.Equal(t, msg.wasCalled, false)
}

func TestAddUser(t *testing.T) {
	payload := `{
		"forename": "Nathan",
		"surname": "Schmitt",
		"nickname": "NBK-",
		"password": "og",
		"email": "ns@notarealemail.com",
		"country": "FRA"
	}`
	expectedUser := &model.User{
		Id:       "not-be-empty",
		Forename: "Nathan",
		Surname:  "Schmitt",
		Nickname: "NBK-",
		Password: "og",
		Email:    "ns@notarealemail.com",
		Country:  "FRA",
	}
	expectedCode := 201
	db := NewMockDaoClient(nil, nil, "None")
	msg := NewMockMsgClient(false)
	handler := NewHandler(db, msg)
	uri := "/users"
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload))
	assert.Nil(t, err)

	code, res, err := handler.AddUser(req)
	assert.Nil(t, err)
	assert.Equal(t, expectedCode, code)
	assert.Equal(t, db.wasCalled, true)
	assert.Equal(t, msg.wasCalled, true)
	compareUser(t, expectedUser, res)
}

// compareUser is a convenience func for testing user equivalence with ID generation
// in a full scenario i'd using something like gosert https://github.com/mina-akimi/gosert
func compareUser(t *testing.T, expected *model.User, given interface{}) {
	res := given.(*model.User)
	assert.NotEqual(t, "", res.Id)
	assert.Equal(t, expected.Forename, res.Forename)
	assert.Equal(t, expected.Surname, res.Surname)
	assert.Equal(t, expected.Nickname, res.Nickname)
	assert.Equal(t, expected.Email, res.Email)
	assert.Equal(t, expected.Password, res.Password)
	assert.Equal(t, expected.Country, res.Country)
}

// For this test I demonstrate how the mocks i've written can be configured to fail at specific points
// I won't do the same for the other endpoints for the sake of the test, but it's simple to expand these
// tests to cover the eventualities
func TestAddUserFail(t *testing.T) {
	tests := []struct {
		name         string
		payload      string
		failFunc     string
		dbCalled     bool
		expectedCode int
	}{
		{
			name: "fail insert",
			payload: `{
				"forename": "Richard",
				"surname": "Papillion",
				"nickname": "shox",
				"password": "vitality",
				"email": "rp@notarealemail.com",
				"country": "FRA"
			}`,
			failFunc:     "Insert",
			dbCalled:     true,
			expectedCode: 500,
		}, {
			name:         "fail unmarshal",
			payload:      `incorrect structure}`,
			failFunc:     "None",
			dbCalled:     false,
			expectedCode: 400,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockDaoClient(nil, nil, tt.failFunc)
			msg := NewMockMsgClient(false)
			handler := NewHandler(db, msg)
			uri := "/users"
			req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(tt.payload))
			assert.Nil(t, err)

			code, res, err := handler.AddUser(req)
			assert.NotNil(t, err)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, db.wasCalled, tt.dbCalled)
			assert.Equal(t, msg.wasCalled, false)
			assert.Nil(t, res)
		})
	}
}

func TestDelete(t *testing.T) {
	payload := &model.User{
		Id:       "dummy-test-user",
		Forename: "Christopher",
		Surname:  "Alesund",
		Nickname: "GeT_RiGhT",
		Password: "nip",
		Email:    "ca@notarealemail.com",
		Country:  "SWE",
	}
	id := "dummy-test-user"
	expectedCode := 204
	db := NewMockDaoClient(payload, nil, "None")
	msg := NewMockMsgClient(false)
	handler := NewHandler(db, msg)
	uri := fmt.Sprintf("/users/%s", id)
	req, err := http.NewRequest(http.MethodDelete, uri, nil)
	assert.Nil(t, err)

	code, res, err := handler.RemoveUser(req)
	assert.Nil(t, err)
	assert.Equal(t, expectedCode, code)
	assert.Equal(t, db.wasCalled, true)
	assert.Equal(t, msg.wasCalled, true)
	assert.Nil(t, res)
}

func TestUpdateUser(t *testing.T) {
	previous := &model.User{
		Id:       "dummy-test-user",
		Forename: "Sean",
		Surname:  "Kaiwai",
		Nickname: "Gratisfaction",
		Password: "renegades",
		Email:    "sk@notarealemail.com",
		Country:  "NZ",
	}
	payload := `{
		"forename": "Sean",
		"surname": "Kaiwai",
		"nickname": "Gratisfaction",
		"password": "100T",
		"email": "sk@notarealemail.com",
		"country": "NZ"
	}`
	expectedUser := &model.User{
		Id:       "dummy-test-user",
		Forename: "Sean",
		Surname:  "Kaiwai",
		Nickname: "Gratisfaction",
		Password: "100T",
		Email:    "sk@notarealemail.com",
		Country:  "NZ",
	}
	expectedCode := 200
	db := NewMockDaoClient(previous, nil, "None")
	msg := NewMockMsgClient(false)
	handler := NewHandler(db, msg)
	uri := "/users/dummy-test-user"
	req, err := http.NewRequest(http.MethodPost, uri, strings.NewReader(payload))
	assert.Nil(t, err)

	code, res, err := handler.UpdateUser(req)
	assert.Nil(t, err)
	assert.Equal(t, expectedCode, code)
	assert.Equal(t, db.wasCalled, true)
	assert.Equal(t, msg.wasCalled, true)
	compareUser(t, expectedUser, db.payload)
	compareUser(t, expectedUser, res)
}

// more a test for the sake of tests; as the filter logic is in the db implementation
func TestFilter(t *testing.T) {
	payload := []*model.User{
		{
			Id:       "dummy-test-user",
			Forename: "Gabriel",
			Surname:  "Toledo",
			Nickname: "FalleN",
			Password: "MIBR",
			Email:    "gt@notarealemail.com",
			Country:  "BRA",
		}, {
			Id:       "dummy-test-user2",
			Forename: "Jacky",
			Surname:  "Yip",
			Nickname: "Stewie2K",
			Password: "liquid",
			Email:    "yj@notarealemail.com",
			Country:  "USA",
		},
	}
	expectedResponse := &model.FilterResponse{
		Count:   2,
		Results: payload,
	}
	tests := []struct {
		name         string
		conditions   map[string]string
		expectedCode int
		expectedFunc string
	}{
		{
			name:         "filter",
			conditions:   map[string]string{"country": "BRA"},
			expectedCode: 200,
			expectedFunc: "Filter",
		}, {
			name:         "get all",
			conditions:   map[string]string{},
			expectedCode: 200,
			expectedFunc: "GetAll",
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			db := NewMockDaoClient(nil, payload, "None")
			msg := NewMockMsgClient(false)
			handler := NewHandler(db, msg)
			uri := "/users"
			req, err := http.NewRequest(http.MethodGet, uri, nil)
			q := req.URL.Query()
			for field, query := range tt.conditions {
				q.Add(field, query)
			}
			req.URL.RawQuery = q.Encode()

			code, res, err := handler.FilterUsers(req)
			assert.Nil(t, err)
			assert.Equal(t, tt.expectedCode, code)
			assert.Equal(t, db.wasCalled, true)
			assert.Equal(t, tt.expectedFunc, db.calledFunc)
			assert.Equal(t, msg.wasCalled, false)
			if !reflect.DeepEqual(expectedResponse, res) {
				t.Errorf("expected response payload %v should match %v", expectedResponse, res)
			}
		})
	}
}

func TestPrepareFilter(t *testing.T) {
	tests := []struct {
		name              string
		query             string
		value             []string
		expectedCondition *model.FilterCondition
		expectedValid     bool
	}{
		{
			name:  "base",
			query: "country",
			value: []string{"GBR"},
			expectedCondition: &model.FilterCondition{
				Query: "country",
				Value: "GBR",
			},
			expectedValid: true,
		}, {
			name:              "invalid query",
			query:             "rank",
			value:             []string{"1"},
			expectedCondition: nil,
			expectedValid:     false,
		}, {
			name:              "missing",
			query:             "",
			value:             []string{},
			expectedCondition: nil,
			expectedValid:     false,
		},
	}
	for _, test := range tests {
		tt := test
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			res, ok := prepareFilter(tt.query, tt.value)
			assert.Equal(t, tt.expectedValid, ok)
			if !reflect.DeepEqual(tt.expectedCondition, res) {
				t.Errorf("filter condition should match expected %v %v", tt.expectedCondition, res)
			}
		})
	}
}
