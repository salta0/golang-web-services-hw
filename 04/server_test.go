package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type serverTestCase struct {
	Name        string
	AccessToken string
	Query       string
	OrderField  string
	OrderType   string
	Limit       string
	Offset      string
	RespBody    string
	Code        int
}

var withoutOrderExpectedBody = `[` +
	`{"Id":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"},` +
	`{"Id":1,"Name":"Hilda Mayer","Age":21,"About":"Sit commodo consectetur minim amet ex. Elit aute mollit fugiat labore sint ipsum dolor cupidatat qui reprehenderit. Eu nisi in exercitation culpa sint aliqua nulla nulla proident eu. Nisi reprehenderit anim cupidatat dolor incididunt laboris mollit magna commodo ex. Cupidatat sit id aliqua amet nisi et voluptate voluptate commodo ex eiusmod et nulla velit.\n","Gender":"female"},` +
	`{"Id":2,"Name":"Brooks Aguilar","Age":25,"About":"Velit ullamco est aliqua voluptate nisi do. Voluptate magna anim qui cillum aliqua sint veniam reprehenderit consectetur enim. Laborum dolore ut eiusmod ipsum ad anim est do tempor culpa ad do tempor. Nulla id aliqua dolore dolore adipisicing.\n","Gender":"male"},` +
	`{"Id":3,"Name":"Everett Dillard","Age":27,"About":"Sint eu id sint irure officia amet cillum. Amet consectetur enim mollit culpa laborum ipsum adipisicing est laboris. Adipisicing fugiat esse dolore aliquip quis laborum aliquip dolore. Pariatur do elit eu nostrud occaecat.\n","Gender":"male"},` +
	`{"Id":4,"Name":"Owen Lynn","Age":30,"About":"Elit anim elit eu et deserunt veniam laborum commodo irure nisi ut labore reprehenderit fugiat. Ipsum adipisicing labore ullamco occaecat ut. Ea deserunt ad dolor eiusmod aute non enim adipisicing sit ullamco est ullamco. Elit in proident pariatur elit ullamco quis. Exercitation amet nisi fugiat voluptate esse sit et consequat sit pariatur labore et.\n","Gender":"male"}]`
var allRecordsExpectedBody = `[` +
	`{"Id":13,"Name":"Whitley Davidson","Age":40,"About":"Consectetur dolore anim veniam aliqua deserunt officia eu. Et ullamco commodo ad officia duis ex incididunt proident consequat nostrud proident quis tempor. Sunt magna ad excepteur eu sint aliqua eiusmod deserunt proident. Do labore est dolore voluptate ullamco est dolore excepteur magna duis quis. Quis laborum deserunt ipsum velit occaecat est laborum enim aute. Officia dolore sit voluptate quis mollit veniam. Laborum nisi ullamco nisi sit nulla cillum et id nisi.\n","Gender":"male"},` +
	`{"Id":33,"Name":"Twila Snow","Age":36,"About":"Sint non sunt adipisicing sit laborum cillum magna nisi exercitation. Dolore officia esse dolore officia ea adipisicing amet ea nostrud elit cupidatat laboris. Proident culpa ullamco aute incididunt aute. Laboris et nulla incididunt consequat pariatur enim dolor incididunt adipisicing enim fugiat tempor ullamco. Amet est ullamco officia consectetur cupidatat non sunt laborum nisi in ex. Quis labore quis ipsum est nisi ex officia reprehenderit ad adipisicing fugiat. Labore fugiat ea dolore exercitation sint duis aliqua.\n","Gender":"female"},` +
	`{"Id":18,"Name":"Terrell Hall","Age":27,"About":"Ut nostrud est est elit incididunt consequat sunt ut aliqua sunt sunt. Quis consectetur amet occaecat nostrud duis. Fugiat in irure consequat laborum ipsum tempor non deserunt laboris id ullamco cupidatat sit. Officia cupidatat aliqua veniam et ipsum labore eu do aliquip elit cillum. Labore culpa exercitation sint sint.\n","Gender":"male"},` +
	`{"Id":26,"Name":"Sims Cotton","Age":39,"About":"Ex cupidatat est velit consequat ad. Tempor non cillum labore non voluptate. Et proident culpa labore deserunt ut aliquip commodo laborum nostrud. Anim minim occaecat est est minim.\n","Gender":"male"},` +
	`{"Id":9,"Name":"Rose Carney","Age":36,"About":"Voluptate ipsum ad consequat elit ipsum tempor irure consectetur amet. Et veniam sunt in sunt ipsum non elit ullamco est est eu. Exercitation ipsum do deserunt do eu adipisicing id deserunt duis nulla ullamco eu. Ad duis voluptate amet quis commodo nostrud occaecat minim occaecat commodo. Irure sint incididunt est cupidatat laborum in duis enim nulla duis ut in ut. Cupidatat ex incididunt do ullamco do laboris eiusmod quis nostrud excepteur quis ea.\n","Gender":"female"}]`
var serverTestCases = []serverTestCase{
	{
		Name:        "Valid params",
		AccessToken: "access_token",
		Query:       "Boyd",
		OrderField:  "Name",
		OrderType:   "1",
		Limit:       "5",
		Offset:      "0",
		RespBody:    `[{"Id":0,"Name":"Boyd Wolf","Age":22,"About":"Nulla cillum enim voluptate consequat laborum esse excepteur occaecat commodo nostrud excepteur ut cupidatat. Occaecat minim incididunt ut proident ad sint nostrud ad laborum sint pariatur. Ut nulla commodo dolore officia. Consequat anim eiusmod amet commodo eiusmod deserunt culpa. Ea sit dolore nostrud cillum proident nisi mollit est Lorem pariatur. Lorem aute officia deserunt dolor nisi aliqua consequat nulla nostrud ipsum irure id deserunt dolore. Minim reprehenderit nulla exercitation labore ipsum.\n","Gender":"male"}]`,
		Code:        http.StatusOK,
	},
	{
		Name:        "Unauthorized",
		AccessToken: "bad_token",
		Query:       "yey",
		OrderField:  "Name",
		OrderType:   "1",
		Limit:       "5",
		Offset:      "0",
		RespBody:    expErrToJson("ErrorUnauthorized"),
		Code:        http.StatusUnauthorized,
	},
	{
		Name:        "All records",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "Name",
		OrderType:   "1",
		Limit:       "5",
		Offset:      "0",
		RespBody:    allRecordsExpectedBody,
		Code:        http.StatusOK,
	},
	{
		Name:        "Without order",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "",
		OrderType:   "",
		Limit:       "5",
		Offset:      "0",
		RespBody:    withoutOrderExpectedBody,
		Code:        http.StatusOK,
	},
	{
		Name:        "Limit over records count",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "",
		OrderType:   "",
		Limit:       "5",
		Offset:      "35",
		RespBody:    `[]`,
		Code:        http.StatusOK,
	},
	{
		Name:        "Invalid order field",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "About",
		OrderType:   "-1",
		Limit:       "5",
		Offset:      "0",
		RespBody:    expErrToJson("ErrorBadOrderField"),
		Code:        http.StatusBadRequest,
	},
	{
		Name:        "Invalid order type",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "Age",
		OrderType:   "3",
		Limit:       "5",
		Offset:      "0",
		RespBody:    expErrToJson("ErrorBadOrderBy"),
		Code:        http.StatusBadRequest,
	},
	{
		Name:        "Invalid limit",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "Age",
		OrderType:   "1",
		Limit:       "0",
		Offset:      "0",
		RespBody:    expErrToJson("ErrorBadLimit"),
		Code:        http.StatusBadRequest,
	},
	{
		Name:        "Invalid offset",
		AccessToken: "access_token",
		Query:       "",
		OrderField:  "Age",
		OrderType:   "1",
		Limit:       "5",
		Offset:      "-1",
		RespBody:    expErrToJson("ErrorBadOffset"),
		Code:        http.StatusBadRequest,
	},
}

func expErrToJson(sErr string) string {
	res, err := json.Marshal(struct{ Error string }{Error: sErr})
	if err != nil {
		panic(err)
	}
	return string(res)
}

func TestHandleSearch(t *testing.T) {
	for _, tCase := range serverTestCases {
		params := url.Values{}
		params.Add("query", tCase.Query)
		params.Add("order_field", tCase.OrderField)
		params.Add("order_by", tCase.OrderType)
		params.Add("limit", tCase.Limit)
		params.Add("offset", tCase.Offset)

		req, err := http.NewRequest("GET", "/"+"?"+params.Encode(), nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Add("AccessToken", tCase.AccessToken)

		rr := httptest.NewRecorder()
		handleSearch(rr, req)

		t.Run(tCase.Name+" response code", func(t *testing.T) {
			if rr.Code != tCase.Code {
				t.Errorf("\nExpected status code %v, \ngot %v", tCase.Code, rr.Code)
			}
		})

		t.Run(tCase.Name+" response body", func(t *testing.T) {
			if rr.Body.String() != tCase.RespBody {
				t.Errorf("\nExpected response body %v,\n got %v", tCase.RespBody, rr.Body.String())
			}
		})
	}
}
