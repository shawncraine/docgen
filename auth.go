package main

import (
	"bytes"
	"crypto/rand"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/spf13/viper"
)

var (
	sess = make(map[string]string, 0)
)

const cookieName = "docgen_doc_box"

func uid() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		fmt.Println("Error: ", err)
		return fmt.Sprintf("%v", time.Now().Unix())
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", b[0:4], b[4:6], b[6:8], b[8:10], b[10:])
}

func verifyAuth(username, password string) bool {
	pass := getAuthSecret(username)
	if pass == "" {
		return false
	}
	return pass == password
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(cookieName)
		if err != nil {
			if r.Header.Get("X-Requested-With") == "xmlhttprequest" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("401"))
				return
			}
			if err == http.ErrNoCookie {
				http.Redirect(w, r, "/login?code=1", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/login?code=1", http.StatusSeeOther)
			return
		}
		sessionToken := c.Value
		if sess[sessionToken] == "" && r.URL.Path != "/login" {
			if r.Header.Get("X-Requested-With") == "xmlhttprequest" {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("401"))
				return
			}
			http.Redirect(w, r, "/login?code=1", http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	}

	return http.HandlerFunc(fn)
}

func getAuthSecret(usrname string) string {
	return (viper.GetStringMapString("auth_secrets"))[strings.ToLower(usrname)]
}

func getLogin(w http.ResponseWriter, r *http.Request) {
	// check already has valid session
	c, err := r.Cookie(cookieName)
	if err != nil {
		log.Println("getLogin:", err)
	}
	if sess[c.Value] != "" {
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	type app struct {
		Name    string
		Message string
	}
	data := struct {
		Assets Assets
		Data   app
	}{
		Assets: assets,
		Data: app{
			Name: viper.GetString("app.name"),
		},
	}
	if f := r.URL.Query().Get("code"); f != "" {
		switch f {
		default:
			data.Data.Message = ""
		case "1":
			data.Data.Message = "Unauthorized! Please login"
		case "2":
			data.Data.Message = "Invalid username or password!"
		}
	}
	buf := new(bytes.Buffer)
	if err := tmDoc.ExecuteTemplate(buf, "login", data); err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func postLogin(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		log.Fatalln(err)
	}

	usrname := strings.TrimSpace(strings.ToLower(r.Form.Get("username")))
	password := strings.TrimSpace(r.Form.Get("password"))
	if verifyAuth(usrname, password) {
		token := uid()
		sess[token] = usrname
		http.SetCookie(w, &http.Cookie{
			Name:    cookieName,
			Value:   token,
			Expires: time.Now().Add(viper.GetDuration("app.session_expire") * time.Minute),
		})
		http.Redirect(w, r, "/home", http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/login?code=2", http.StatusSeeOther)
}

func logout(w http.ResponseWriter, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err != nil {
		log.Println("logut:", err)
	}
	delete(sess, c.Value)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func hasCRUDpermission(username string) bool {
	pp := viper.GetStringSlice("app.crud")
	for _, p := range pp {
		if p == username {
			return true
		}
	}
	return false
}
