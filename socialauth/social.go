package socialauth

import (
	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/github"
	"net/http"
	"os"
)

type SocialLogin struct {
	Session *scs.SessionManager
}

func (s *SocialLogin) InitSocialAuth() {
	githubKey := os.Getenv("GITHUB_KEY")
	githubSecret := os.Getenv("GITHUB_SECRET")

	scope := []string{"user"}

	provider := github.New(
		githubKey,
		githubSecret,
		"http://localhost:4000/auth/github/callback",
		scope...,
	)

	var providers []goth.Provider
	providers = append(providers, provider)

	goth.UseProviders(providers...)

	key := os.Getenv("KEY")
	maxAge := 86400 * 30
	store := sessions.NewCookieStore([]byte(key))
	store.MaxAge(maxAge)
	store.Options.Path = "/"
	store.Options.HttpOnly = true
	store.Options.Secure = false

	gothic.Store = store
}

func (s *SocialLogin) GithubLogin(w http.ResponseWriter, r *http.Request) {
	s.InitSocialAuth()
	if _, err := gothic.CompleteUserAuth(w, r); err == nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
	} else {
		gothic.BeginAuthHandler(w, r)
	}
}

// SocialMediaCallback is called after the user agrees to try to log in;
// note that goth cookies are internal and only used as part of the auth flow, so
// our application must maintain its own session/authentication state
// from the data provided back after calling gothic.CompleteUserAuth.
func (s *SocialLogin) SocialMediaCallback(w http.ResponseWriter, r *http.Request) {
	s.InitSocialAuth()
	gUser, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		s.Session.Put(r.Context(), "error", err.Error())
		http.Redirect(w, r, "/users/login", http.StatusSeeOther)
		return
	}
	s.Session.Put(r.Context(), "userID", 1)
	s.Session.Put(r.Context(), "gUser", gUser.Email)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *SocialLogin) SocialMediaLogout(w http.ResponseWriter, r *http.Request) {
	s.InitSocialAuth()
	s.Session.RenewToken(r.Context())
	s.Session.Remove(r.Context(), "userID")
	s.Session.Remove(r.Context(), "remember_token")
	s.Session.Destroy(r.Context())
	s.Session.RenewToken(r.Context())

	_ = gothic.Logout(w, r)

	http.Redirect(w, r, "/users/login", http.StatusTemporaryRedirect)
}
