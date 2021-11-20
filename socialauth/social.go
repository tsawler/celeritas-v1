package socialauth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/alexedwards/scs/v2"
	"github.com/go-chi/chi/v5"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"github.com/markbates/goth/providers/google"
	"log"
	"net/http"
	"os"
)

type SocialLogin struct {
	Session *scs.SessionManager
}

func (s *SocialLogin) InitSocialAuth(r *http.Request) {
	m := make(map[string]string)

	m["github"] = "Github"

	scope := []string{"user"}

	goth.UseProviders(
		github.New(os.Getenv("GITHUB_KEY"), os.Getenv("GITHUB_SECRET"), "http://localhost:4000/auth/github/callback", scope...),
		google.New(os.Getenv("GOOGLE_KEY"), os.Getenv("GOOGLE_SECRET"), "http://localhost:4000/auth/google/callback", scope...),
	)

	var providers []goth.Provider

	goth.UseProviders(providers...)

	key := os.Getenv("KEY")
	maxAge := 86400 * 30
	st := sessions.NewCookieStore([]byte(key))
	st.MaxAge(maxAge)
	st.Options.Path = "/"
	st.Options.HttpOnly = true
	st.Options.Secure = false

	gothic.Store = st
}

func (s *SocialLogin) SocialLogin(w http.ResponseWriter, r *http.Request) {
	// save provider type in session
	provider := chi.URLParam(r, "provider")

	s.InitSocialAuth(r)
	s.Session.Put(r.Context(), "social_provider", provider)

	if _, err := gothic.CompleteUserAuth(w, r); err == nil {
		// already logged in
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		// attempt login
		gothic.BeginAuthHandler(w, r)
	}
}

// SocialMediaCallback is called after the user agrees to try to log in;
// note that goth cookies are internal and only used as part of the auth flow, so
// our application must maintain its own session/authentication state
// from the data provided back after calling gothic.CompleteUserAuth.
func (s *SocialLogin) SocialMediaCallback(w http.ResponseWriter, r *http.Request) {
	s.InitSocialAuth(r)
	gUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		s.Session.Put(r.Context(), "error", err.Error())
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
		return
	}

	// TODO -- look  up user by email address
	s.Session.Put(r.Context(), "userID", 1)
	s.Session.Put(r.Context(), "social_token", gUser.AccessToken)
	s.Session.Put(r.Context(), "social_email", gUser.Email)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

type Payload struct {
	AccessToken string `json:"access_token"`
}

func (s *SocialLogin) SocialMediaLogout(w http.ResponseWriter, r *http.Request) {
	s.InitSocialAuth(r)

	provider := s.Session.Get(r.Context(), "social_provider").(string)

	if provider == "github" {
		// call remote api and revoke token
		clientID := os.Getenv("GITHUB_CLIENT_ID")
		clientSecret := os.Getenv("GITHUB_SECRET")
		token := s.Session.Get(r.Context(), "social_token").(string)

		payload := Payload{
			AccessToken: token,
		}

		jsonReq, _ := json.Marshal(payload)
		req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("https://%s:%s@api.github.com/applications/%s/grant", clientID, clientSecret, clientID), bytes.NewBuffer(jsonReq))
		if err != nil {
			log.Println("Error building request", err)
		}

		client := &http.Client{}
		_, err = client.Do(req)
		if err != nil {
			log.Println("Error calling client.Do()", err)
		}
	}

	s.Session.RenewToken(r.Context())
	s.Session.Remove(r.Context(), "userID")
	s.Session.Remove(r.Context(), "remember_token")
	s.Session.Destroy(r.Context())
	s.Session.RenewToken(r.Context())

	gothic.Logout(w, r)

	w.Header().Set("Location", "/users/login")
	w.WriteHeader(http.StatusTemporaryRedirect)
}
