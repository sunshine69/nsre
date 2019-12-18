package cmd

import (
	"bytes"
	"net/http"
	"fmt"
	"io/ioutil"
	"context"
	"log"
	"encoding/base64"
	"crypto/rand"
	"github.com/json-iterator/go"
	"time"
)

//OauthGoogleLogin -
func OauthGoogleLogin(w http.ResponseWriter, r *http.Request) {
	oauthState := generateStateOauthCookie(w)
	u := GoogleOauthConfig.AuthCodeURL(oauthState)
	http.Redirect(w, r, u, http.StatusTemporaryRedirect)
}

//OauthGoogleCallback -
func OauthGoogleCallback(w http.ResponseWriter, r *http.Request) {
	oauthState, _ := r.Cookie("oauthstate")
	if r.FormValue("state") != oauthState.Value {
		log.Println("invalid oauth google state")
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	data, err := GetUserDataFromGoogle(r.FormValue("code"))
	if err != nil {
		log.Println(err.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}

	var json = jsoniter.ConfigCompatibleWithStandardLibrary

	//UserInfo -
	type UserInfo struct {
		Id string `json:"id"`
		Email string `json:"email"`
		Verified_email bool `json:"verified_email"`
		Picture string `json:"picture"`
		Hd string `json:"hd"`
	}
	user := UserInfo{}
	data1 := bytes.Replace(data, []byte("UserInfo: "), []byte("") , 1)

	if e := json.Unmarshal(data1, &user); e != nil {
		log.Println(e.Error())
		http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
		return
	}
	
	log.Printf("TRY_LOGIN - user email: %s, ID: %s\n", user.Email, user.Id)

	if (! user.Verified_email) || ! CheckAuthorizedUser(user.Email) {
		log.Println("Unauthorised for user not in authorised domain or email not verified")
		http.Redirect(w, r, "/auth/google/login", http.StatusTemporaryRedirect)
		return
	}

	session, err := SessionStore.Get(r, "auth-session")
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

	session.Values["oauthstate"] = oauthState.Value
	session.Values["useremail"] = user.Email

	if err := session.Save(r, w); err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
	}

	http.Redirect(w, r, "/searchlog", http.StatusSeeOther)
}

func generateStateOauthCookie(w http.ResponseWriter) string {
	var expiration = time.Now().Add(20 * time.Minute)
	b := make([]byte, 16)
	rand.Read(b)
	state := base64.URLEncoding.EncodeToString(b)
	cookie := http.Cookie{Name: "oauthstate", Value: state, Expires: expiration}
	http.SetCookie(w, &cookie)
	return state
}

//GetUserDataFromGoogle -
func GetUserDataFromGoogle(code string) ([]byte, error) {
	token, err := GoogleOauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return nil, fmt.Errorf("code exchange wrong: %s", err.Error())
	}
	response, err := http.Get(OauthGoogleUrlAPI + token.AccessToken)
	if err != nil {
		return nil, fmt.Errorf("failed getting user info: %s", err.Error())
	}
	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("failed read response: %s", err.Error())
	}
	return contents, nil
}