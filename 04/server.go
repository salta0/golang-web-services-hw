package main

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"net/http"
	"os"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

const AccessToken = "access_token"

// Serve XML users
type UserRows struct {
	XMLName xml.Name  `xml:"root"`
	Rows    []UserRow `xml:"row"`
}
type UserRow struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

func (u *UserRow) MakeUser() User {
	return User{
		Id:     u.ID,
		Name:   u.FirstName + " " + u.LastName,
		About:  u.About,
		Age:    u.Age,
		Gender: u.Gender,
	}
}

// Serve request params
type SearchParams struct {
	Query       string
	OrderField  string
	OrderBy     string
	Limit       int
	Offset      int
	AccessToken string
}

func (s *SearchParams) parseParamsFrom(r *http.Request) error {
	s.AccessToken = r.Header.Get("AccessToken")
	s.Query = r.FormValue("query")
	s.OrderField = r.FormValue("order_field")
	s.OrderBy = r.FormValue("order_by")
	limit, err := strconv.Atoi(r.FormValue("limit"))
	if err != nil {
		return err
	}
	s.Limit = limit
	offset, err := strconv.Atoi(r.FormValue("offset"))
	if err != nil {
		return err
	}
	s.Offset = offset

	return nil
}
func (s *SearchParams) Validate() error {
	permittedOrderField := []string{"", "Id", "Age", "Name"}
	if !(slices.Contains(permittedOrderField, s.OrderField)) {
		return errors.New("ErrorBadOrderField")
	}
	permittedOrderBy := []string{"-1", "0", "1"}
	if s.OrderField != "" && !(slices.Contains(permittedOrderBy, s.OrderBy)) {
		return errors.New("ErrorBadOrderBy")
	}

	if s.Limit <= 0 {
		return errors.New("ErrorBadLimit")
	}
	if s.Offset < 0 {
		return errors.New("ErrorBadOffset")
	}
	return nil
}

// Serve find users
func findUsers(params *SearchParams) ([]User, error) {
	data, err := os.ReadFile("dataset.xml")
	if err != nil {
		return []User{}, err
	}

	var users UserRows
	err = xml.Unmarshal(data, &users)
	if err != nil {
		return []User{}, err
	}
	var result []User
	for _, row := range users.Rows {
		if applyNameFilter(row, params.Query) || applyAboutFilter(row, params.Query) {
			result = append(result, row.MakeUser())
		}
	}

	if len(result) == 0 {
		return []User{}, nil
	}

	return result, nil
}

func applyNameFilter(user UserRow, value string) bool {
	if value == "" {
		return true
	} else {
		return strings.Contains(user.FirstName+" "+user.LastName, value)
	}
}
func applyAboutFilter(user UserRow, value string) bool {
	if value == "" {
		return true
	} else {
		return strings.Contains(user.About, value)
	}
}
func applyOrder(users []User, params *SearchParams) []User {
	if params.OrderBy == "0" || params.OrderField == "" {
		return users
	}

	stringUserGetter := map[string]func(User) string{
		"Name": func(u User) string { return u.Name },
	}
	intUserGetter := map[string]func(User) int{
		"Id":  func(u User) int { return u.Id },
		"Age": func(u User) int { return u.Age },
	}

	orderField := params.OrderField
	switch {
	case orderField == "Id" || orderField == "Age":
		if params.OrderBy == "-1" {
			sort.Slice(users, func(i, j int) bool {
				return intUserGetter[orderField](users[i]) < intUserGetter[orderField](users[j])
			})
		} else if params.OrderBy == "1" {
			sort.Slice(users, func(i, j int) bool {
				return intUserGetter[orderField](users[i]) > intUserGetter[orderField](users[j])
			})
		}
	case orderField == "Name":
		if params.OrderBy == "-1" {
			sort.Slice(users, func(i, j int) bool {
				return stringUserGetter[orderField](users[i]) < stringUserGetter[orderField](users[j])
			})
		} else if params.OrderBy == "1" {
			sort.Slice(users, func(i, j int) bool {
				return stringUserGetter[orderField](users[i]) > stringUserGetter[orderField](users[j])
			})
		}
	}

	return users
}
func applyPagination(users []User, params *SearchParams) []User {
	userLen := len(users)
	offset := params.Offset
	limit := params.Limit
	if (offset - 1) >= userLen {
		return []User{}
	} else if (offset-1) <= userLen && (offset-1)+limit >= userLen {
		return users[offset:userLen]
	}

	return users[offset:(offset + limit)]
}

// Search http handler
func handleSearch(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := &SearchParams{}
	params.parseParamsFrom(r)

	if params.AccessToken != AccessToken {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write(jsonError("ErrorUnauthorized"))
		return
	}

	err := params.Validate()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(jsonError(err.Error()))
		return
	}

	users, err := findUsers(params)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonError("ErrorInternalServerError"))
		return
	}

	orderedUsers := applyOrder(users, params)
	paginatedUsers := applyPagination(orderedUsers, params)

	jsonRes, err := json.Marshal(paginatedUsers)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write(jsonError("ErrorInternalServerError"))
		return
	}
	w.Write(jsonRes)
}

func jsonError(sErr string) []byte {
	res, err := json.Marshal(struct{ Error string }{Error: sErr})
	if err != nil {
		panic(err)
	}
	return res
}

// Search server
type SearchServer struct {
	AuthToken    string
	Port         string
	Handler      http.Handler
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

func (s *SearchServer) Run() {
	server := http.Server{
		Addr:         s.Port,
		Handler:      s.Handler,
		ReadTimeout:  s.ReadTimeout,
		WriteTimeout: s.WriteTimeout,
	}

	fmt.Println("Server started")
	server.ListenAndServe()
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/", handleSearch)

	server := &SearchServer{
		Port:         ":8080",
		Handler:      mux,
		ReadTimeout:  10 * time.Millisecond,
		WriteTimeout: 10 * time.Millisecond,
	}
	server.Run()
}
