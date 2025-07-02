package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"slices"
	"strconv"
)

type ErrorApi struct {
	Error string `json:"error"`
}

func (p *ProfileParams) Fill(r *http.Request) error {
	login := r.FormValue("login")

	if login == "" {
		return errors.New("login must be not empty")
	}
	p.Login = login
	return nil
}
func (p *CreateParams) Fill(r *http.Request) error {
	login := r.FormValue("login")

	if login == "" {
		return errors.New("login must be not empty")
	}
	if len(login) < 10 {
		return errors.New("login len must be >= 10")
	}
	p.Login = login
	full_name := r.FormValue("full_name")
	p.Name = full_name
	status := r.FormValue("status")
	if status == "" {
		status = "user"
	}
	if !(slices.Contains([]string{"user", "moderator", "admin"}, status)) {
		return errors.New("status must be one of [user, moderator, admin]")
	}
	p.Status = status
	age, err := strconv.Atoi(r.FormValue("age"))
	if err != nil {
		return errors.New("age must be int")
	}
	if age < 0 {
		return errors.New("age must be >= 0")
	}
	if age > 128 {
		return errors.New("age must be <= 128")
	}
	p.Age = age
	return nil
}
func (p *OtherCreateParams) Fill(r *http.Request) error {
	username := r.FormValue("username")
	if len(username) < 3 {
		return errors.New("username len must be >= 3")
	}

	if username == "" {
		return errors.New("username must be not empty")
	}
	p.Username = username
	account_name := r.FormValue("account_name")
	p.Name = account_name
	class := r.FormValue("class")
	if class == "" {
		class = "warrior"
	}
	if !(slices.Contains([]string{"warrior", "sorcerer", "rouge"}, class)) {
		return errors.New("class must be one of [warrior, sorcerer, rouge]")
	}
	p.Class = class
	level, err := strconv.Atoi(r.FormValue("level"))
	if err != nil {
		return errors.New("level must be int")
	}
	if level < 1 {
		return errors.New("level must be >= 1")
	}
	if level > 50 {
		return errors.New("level must be <= 50")
	}
	p.Level = level
	return nil
}
func (srv *MyApi) handleProfile(w http.ResponseWriter, r *http.Request) {
	params := ProfileParams{}
	err := params.Fill(r)
	if err != nil {
		paramsErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(paramsErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusBadRequest)
		return
	}
	resp, err := srv.Profile(context.Background(), params)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if apiError, ok := err.(ApiError); ok {
			httpCode = apiError.HTTPStatus
		}
		handleErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), httpCode)
		return
	}
	respStruct := struct {
		Error    string `json:"error"`
		Response *User  `json:"response"`
	}{
		Error:    "",
		Response: resp,
	}
	jsonResp, err := json.Marshal(respStruct)
	if err != nil {
		panic(err)
	}
	w.Write(jsonResp)
}
func (srv *MyApi) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Auth") != "100500" {
		authErr := &ErrorApi{Error: "unauthorized"}
		response, err := json.Marshal(authErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusForbidden)
		return
	}
	params := CreateParams{}
	err := params.Fill(r)
	if err != nil {
		paramsErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(paramsErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusBadRequest)
		return
	}
	resp, err := srv.Create(context.Background(), params)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if apiError, ok := err.(ApiError); ok {
			httpCode = apiError.HTTPStatus
		}
		handleErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), httpCode)
		return
	}
	respStruct := struct {
		Error    string   `json:"error"`
		Response *NewUser `json:"response"`
	}{
		Error:    "",
		Response: resp,
	}
	jsonResp, err := json.Marshal(respStruct)
	if err != nil {
		panic(err)
	}
	w.Write(jsonResp)
}
func (srv *OtherApi) handleCreate(w http.ResponseWriter, r *http.Request) {
	if r.Header.Get("X-Auth") != "100500" {
		authErr := &ErrorApi{Error: "unauthorized"}
		response, err := json.Marshal(authErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusForbidden)
		return
	}
	params := OtherCreateParams{}
	err := params.Fill(r)
	if err != nil {
		paramsErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(paramsErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusBadRequest)
		return
	}
	resp, err := srv.Create(context.Background(), params)
	if err != nil {
		httpCode := http.StatusInternalServerError
		if apiError, ok := err.(ApiError); ok {
			httpCode = apiError.HTTPStatus
		}
		handleErr := &ErrorApi{Error: err.Error()}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), httpCode)
		return
	}
	respStruct := struct {
		Error    string     `json:"error"`
		Response *OtherUser `json:"response"`
	}{
		Error:    "",
		Response: resp,
	}
	jsonResp, err := json.Marshal(respStruct)
	if err != nil {
		panic(err)
	}
	w.Write(jsonResp)
}
func (srv *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		srv.handleProfile(w, r)
	case "/user/create":
		if r.Method != "POST" {
			methodErr := &ErrorApi{Error: "bad method"}
			response, err := json.Marshal(methodErr)
			if err != nil {
				panic(err)
			}
			http.Error(w, string(response), http.StatusNotAcceptable)
			return
		}
		srv.handleCreate(w, r)
	default:
		handleErr := &ErrorApi{Error: "unknown method"}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusNotFound)
	}
}
func (srv *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		if r.Method != "POST" {
			methodErr := &ErrorApi{Error: "bad method"}
			response, err := json.Marshal(methodErr)
			if err != nil {
				panic(err)
			}
			http.Error(w, string(response), http.StatusNotAcceptable)
			return
		}
		srv.handleCreate(w, r)
	default:
		handleErr := &ErrorApi{Error: "unknown method"}
		response, err := json.Marshal(handleErr)
		if err != nil {
			panic(err)
		}
		http.Error(w, string(response), http.StatusNotFound)
	}
}
