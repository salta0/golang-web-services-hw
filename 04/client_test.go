package main

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"
)

type clientTestCase struct {
	Name       string
	AccessKey  string
	Query      string
	OrderField string
	OrderType  int
	Limit      int
	Offset     int
	Result     *SearchResponse
	Error      error
}

func errToJSON(err string) []byte {
	errJSON, mErr := json.Marshal(struct{ Error string }{Error: err})
	if mErr != nil {
		panic(mErr)
	}

	return errJSON
}

func usersToJSON(users []User) []byte {
	usersJSON, mErr := json.Marshal(users)
	if mErr != nil {
		panic(mErr)
	}

	return usersJSON
}

func handleSearchDummy(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("AccessToken") != "access_token" {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error": "Unauhtorized"}`)
		return
	}

	switch {
	case r.FormValue("order_field") == "Foo":
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errToJSON("ErrorBadOrderField"))
	case r.FormValue("order_by") == "5":
		w.WriteHeader(http.StatusBadRequest)
		w.Write(errToJSON("ErrorBadOrderBy"))
	case r.FormValue("query") == "Bad JSON":
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`"{"id":1,"name":"Owen Lynn",`))
	case r.FormValue("query") == "InternalServerError":
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(errToJSON("InternalServerError"))
	case r.FormValue("query") == "":
		w.WriteHeader(http.StatusOK)
		users := []User{
			{Id: 0, Name: "Boyd Wolf", Age: 22, About: "Velit ullamco est aliqua.\n"},
			{Id: 1, Name: "Owen Lynn", Age: 25, About: "Ex et excepteur anim in eiusmod.\n"},
			{Id: 2, Name: "Beulah Stark", Age: 19, About: "Lorem magna dolore et velit ut officia.\n"},
		}
		w.Write(usersToJSON(users))
	case r.FormValue("query") == "Boyd":
		w.WriteHeader(http.StatusOK)
		users := []User{
			{Id: 0, Name: "Boyd Wolf", Age: 22, About: "Velit ullamco est aliqua.\n"},
		}
		w.Write(usersToJSON(users))
	case r.FormValue("query") == "No Found":
		w.WriteHeader(http.StatusOK)
		users := []User{}
		w.Write(usersToJSON(users))
	case r.FormValue("query") == "Timeout":
		time.Sleep(2 * time.Second)
		w.WriteHeader(http.StatusOK)
		users := []User{}
		w.Write(usersToJSON(users))
	case r.FormValue("query") == "Bad JSON error":
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error"}`))
	}
}

var clientTestCases = []clientTestCase{
	{
		Name:       "Bad access token",
		AccessKey:  "",
		Query:      "Lorem",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("Bad AccessToken"),
	},
	{
		Name:       "Bad order field",
		AccessKey:  "access_token",
		Query:      "Lorem",
		OrderField: "Foo",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("OrderFeld Foo invalid"),
	},
	{
		Name:       "Bad order type",
		AccessKey:  "access_token",
		Query:      "Lorem",
		OrderField: "Age",
		OrderType:  5,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("unknown bad request error: ErrorBadOrderBy"),
	},
	{
		Name:       "Find users",
		AccessKey:  "access_token",
		Query:      "Boyd",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result: &SearchResponse{
			Users: []User{
				{
					Id:    0,
					Name:  "Boyd Wolf",
					Age:   22,
					About: "Velit ullamco est aliqua.\n",
				},
			},
			NextPage: false,
		},
		Error: nil,
	},
	{
		Name:       "Several pages",
		AccessKey:  "access_token",
		Query:      "",
		OrderField: "Name",
		OrderType:  0,
		Limit:      2,
		Offset:     0,
		Result: &SearchResponse{
			Users: []User{
				{Id: 0, Name: "Boyd Wolf", Age: 22, About: "Velit ullamco est aliqua.\n"},
				{Id: 1, Name: "Owen Lynn", Age: 25, About: "Ex et excepteur anim in eiusmod.\n"},
			},
			NextPage: true,
		},
		Error: nil,
	},
	{
		Name:       "No user find",
		AccessKey:  "access_token",
		Query:      "No Found",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		Error: nil,
	},
	{
		Name:       "Bad json response",
		AccessKey:  "access_token",
		Query:      "Bad JSON",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("cant unpack result json: invalid character 'i' after top-level value"),
	},
	{
		Name:       "Search server 500 error",
		AccessKey:  "access_token",
		Query:      "InternalServerError",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("SearchServer fatal error"),
	},
	{
		Name:       "Timeout",
		AccessKey:  "access_token",
		Query:      "Timeout",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("timeout for limit=6&offset=0&order_by=0&order_field=Name&query=Timeout"),
	},
	{
		Name:       "Negative limit",
		AccessKey:  "access_token",
		Query:      "No Found",
		OrderField: "Name",
		OrderType:  0,
		Limit:      -5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("limit must be > 0"),
	},
	{
		Name:       "Negative offset",
		AccessKey:  "access_token",
		Query:      "No Found",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     -5,
		Result:     nil,
		Error:      errors.New("offset must be > 0"),
	},
	{
		Name:       "Limit more 25",
		AccessKey:  "access_token",
		Query:      "No Found",
		OrderField: "Name",
		OrderType:  0,
		Limit:      26,
		Offset:     0,
		Result: &SearchResponse{
			Users:    []User{},
			NextPage: false,
		},
		Error: nil,
	},
	{
		Name:       "Response with invalid json error",
		AccessKey:  "access_token",
		Query:      "Bad JSON error",
		OrderField: "Name",
		OrderType:  0,
		Limit:      5,
		Offset:     0,
		Result:     nil,
		Error:      errors.New("cant unpack error json: invalid character '}' after object key"),
	},
}

func TestSearchClient(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleSearchDummy))
	for _, tCase := range clientTestCases {
		client := SearchClient{
			AccessToken: tCase.AccessKey,
			URL:         ts.URL,
		}
		request := SearchRequest{
			Query:      tCase.Query,
			OrderField: tCase.OrderField,
			OrderBy:    tCase.OrderType,
			Limit:      tCase.Limit,
			Offset:     tCase.Offset,
		}
		res, err := client.FindUsers(request)

		t.Run(tCase.Name+" result", func(t *testing.T) {
			if (res == nil && tCase.Result != nil) || (res != nil && tCase.Result == nil) {
				t.Errorf("Expected result: %v\n Got result: %v", &tCase.Result, &res)
				return
			}

			if res == nil && tCase.Result == nil {
				return
			}

			if !slices.Equal(res.Users, tCase.Result.Users) {
				t.Errorf("Expected users: %v\n Got users: %v", tCase.Result.Users, res.Users)
			}

			if res.NextPage != tCase.Result.NextPage {
				t.Errorf("Expected NextPage: %v\n Got NextPage: %v", tCase.Result.NextPage, res.NextPage)
			}
		})
		t.Run(tCase.Name+" error", func(t *testing.T) {
			if (res == nil && tCase.Result != nil) || (res != nil && tCase.Result == nil) {
				t.Errorf("Expected error: %v\n Got error: %v", tCase.Error, err)
				return
			}

			if err == nil && tCase.Error == nil {
				return
			}

			if err.Error() != tCase.Error.Error() {
				t.Errorf("Expected %v\n Got %v", tCase.Error, err)
			}
		})
	}
	client := SearchClient{
		AccessToken: "access_token",
		URL:         "bad_url",
	}
	request := SearchRequest{
		Query:      "",
		OrderField: "Name",
		OrderBy:    0,
		Limit:      5,
		Offset:     0,
	}
	res, err := client.FindUsers(request)
	t.Run("Bad url"+" error", func(t *testing.T) {
		expErr := errors.New(
			"unknown error Get \"" +
				"bad_url?limit=6&offset=0&order_by=0&order_field=Name&query=\": " +
				"unsupported protocol scheme \"\"")
		if err.Error() != expErr.Error() {
			t.Errorf("Expected error: %v\n Got error: %v", expErr, err)
		}

		if res != nil {
			t.Errorf("Expected result: %v\n Got result: %v", nil, res)
		}
	})

	ts.Close()
}
