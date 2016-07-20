package main

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/boltdb/bolt"
	goboardbackend "github.com/dguihal/goboard/backend"
	goboardcookie "github.com/dguihal/goboard/cookie"
	goboarduser "github.com/dguihal/goboard/user"
	"github.com/gorilla/mux"
)

type AdminHandler struct {
	GoboardHandler

	adminToken string
}

func NewAdminHandler(db *bolt.DB) (a *AdminHandler) {
	a = &AdminHandler{}

	a.db = db

	a.supportedOps = []SupportedOp{
		{"/admin/user/{login}", "DELETE"}, // Delete a user
		{"/admin/post/{id}", "DELETE"},    // Delete a post
	}

	a.adminToken = "plop"
	return
}

func (a *AdminHandler) ServeHTTP(w http.ResponseWriter, rq *http.Request) {
	reqAdminToken := rq.Header.Get("Token-Id")
	if !a.checkAdminToken(reqAdminToken) {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	vars := mux.Vars(rq)

	switch rq.Method {
	case "DELETE":
		if strings.HasPrefix(rq.URL.Path, "/admin/user/") {
			login := vars["login"]
			a.DeleteUser(w, login)
		} else if strings.HasPrefix(rq.URL.Path, "/admin/post/") {
			postId := vars["postId"]
			a.DeletePost(w, postId)
		} else {
			w.WriteHeader(http.StatusNotFound)
			return
		}
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func (a *AdminHandler) DeleteUser(w http.ResponseWriter, login string) {

	if err := goboarduser.DeleteUser(a.db, login); err != nil {
		if uerr, ok := err.(*goboarduser.UserError); ok {
			if uerr.ErrCode == goboarduser.UserDoesNotExistsError {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte(fmt.Sprintf("User %s Not found", login)))
				return
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				fmt.Println(err.Error())
			}
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Println(err.Error())
		}
	}

	if err := goboardcookie.DeleteCookiesForUser(a.db, login); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	return

}

func (a *AdminHandler) DeletePost(w http.ResponseWriter, postId string) {

	id, err := strconv.ParseUint(postId, 10, 64)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
	}

	if err := goboardbackend.DeletePost(a.db, id); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Println(err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	return

}

func (a *AdminHandler) checkAdminToken(token string) bool {
	return token == a.adminToken
}
