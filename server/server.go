package server

import (
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
	"github.com/ory/hydra/sdk/go/hydra"
	"github.com/ory/hydra/sdk/go/hydra/swagger"
	"github.com/pragkent/hydra-wework/wework"
)

const (
	requiredScope = "openid"

	pathConsent  = "/wework/consent"
	pathAuth     = "/wework/auth"
	pathCallback = "/wework/callback"
)

type Server struct {
	cfg   *Config
	mux   *mux.Router
	hcli  hydra.SDK
	wcli  *wework.Client
	store sessions.Store
}

func New(c *Config) (*Server, error) {
	hcli, err := hydra.NewSDK(&hydra.Configuration{
		ClientID:     c.HydraClientID,
		ClientSecret: c.HydraClientSecret,
		EndpointURL:  c.HydraURL,
		Scopes:       []string{"hydra.consent", "hydra.warden.groups"},
	})

	if err != nil {
		return nil, err
	}

	store := sessions.NewCookieStore([]byte(c.CookieSecret))
	store.MaxAge(86400)

	srv := &Server{
		cfg:   c,
		mux:   mux.NewRouter(),
		hcli:  hcli,
		wcli:  wework.NewClient(c.WeworkCorpID, c.WeworkAgentID, c.WeworkSecret),
		store: store,
	}

	srv.mux.HandleFunc(pathConsent, srv.ConsentHandler)
	srv.mux.HandleFunc(pathAuth, srv.AuthHandler)
	srv.mux.HandleFunc(pathCallback, srv.CallbackHandler)

	return srv, nil
}

func (s *Server) ListenAndServe() error {
	lis, err := net.Listen("tcp", s.cfg.BindAddr)
	if err != nil {
		return err
	}

	log.Printf("Listening on %v", lis.Addr())
	return http.Serve(lis, s.mux)
}

func (s *Server) ConsentHandler(w http.ResponseWriter, r *http.Request) {
	reqID := consentID(r)
	if reqID == "" {
		log.Printf("Consent request id is missing")
		http.Error(w, "Consent request id is missing", http.StatusBadRequest)
		return
	}

	request, response, err := s.hcli.GetOAuth2ConsentRequest(reqID)
	if err != nil {
		log.Printf("Get consent request failed. %v", err)
		http.Error(w, "Get consent request failed", http.StatusBadRequest)
		return
	}

	if response.StatusCode != http.StatusOK {
		log.Printf("Get consent request unexpected http status: %v", response.Status)
		http.Error(w, "Get consent request error", http.StatusBadRequest)
		return
	}

	session := s.session(r)
	uid, ok := session.Values["uid"].(string)
	if !ok || uid == "" {
		log.Printf("User not signed in")
		http.Redirect(w, r, getAuthURL(consentID(r)), http.StatusFound)
		return
	}

	extraVars := make(map[string]interface{})
	if err := s.getTokenExtraVars(uid, extraVars); err != nil {
		log.Printf("Get token extra vars error: %v", err)
		http.Error(w, "Get token extra vars error", http.StatusInternalServerError)
		return
	}

	response, err = s.hcli.AcceptOAuth2ConsentRequest(reqID, swagger.ConsentRequestAcceptance{
		Subject:          subjectOf(uid),
		GrantScopes:      getScopes(request.RequestedScopes),
		AccessTokenExtra: extraVars,
		IdTokenExtra:     extraVars,
	})

	if err != nil {
		log.Printf("Accept consent request failed. %v", err)
		http.Error(w, "Accept consent request error", http.StatusInternalServerError)
		return
	}

	if response.StatusCode != http.StatusNoContent {
		log.Printf("Accept consent request unexpected http status: %v", response.Status)
		http.Error(w, "Accept consent request error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, request.RedirectUrl, http.StatusFound)
}

func subjectOf(uid string) string {
	return "user:" + uid
}

func consentID(r *http.Request) string {
	return r.URL.Query().Get("consent")
}

func getScopes(scopes []string) []string {
	if !contains(scopes, requiredScope) {
		scopes = append(scopes, requiredScope)
	}

	return scopes
}

func contains(values []string, s string) bool {
	for _, i := range values {
		if i == s {
			return true
		}
	}

	return false
}

func getAuthURL(consentID string) string {
	return fmt.Sprintf("%s?consent=%s", pathAuth, consentID)
}

func (s *Server) getTokenExtraVars(uid string, vars map[string]interface{}) error {
	userResp, err := s.wcli.GetUser(uid)
	if err != nil {
		return fmt.Errorf("Get wework user failed. %v", err)
	}

	if userResp.Status != wework.UserActive {
		return errors.New("User is not active")
	}

	vars["username"] = userResp.UserID
	vars["email"] = userResp.Email
	vars["name"] = userResp.EnglishName

	gs, _, err := s.hcli.ListGroups(subjectOf(uid), 100, 0)
	if err != nil {
		return fmt.Errorf("Get hydra groups failed. %v", err)
	}

	var groups []string
	for _, g := range gs {
		groups = append(groups, g.Id)
	}

	vars["groups"] = groups

	return nil
}

func (s *Server) AuthHandler(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("consent")
	callbackURL := getWeworkCallbackURL(s.cfg.HTTPS, r.Host)

	u := s.wcli.GetOAuthURL(callbackURL, state)
	http.Redirect(w, r, u, http.StatusFound)
}

func getWeworkCallbackURL(https bool, host string) string {
	scheme := "https"
	if !https {
		scheme = "http"
	}

	return fmt.Sprintf("%s://%s%s", scheme, host, pathCallback)
}

func (s *Server) CallbackHandler(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	uid, err := s.wcli.GetUserInfo(code)
	if err != nil {
		log.Printf("Get user info failed. %v", err)
		http.Error(w, "Get user info failed", http.StatusInternalServerError)
		return
	}

	log.Printf("User signed in as %v", uid)
	session := s.session(r)
	session.Values["uid"] = uid
	session.Save(r, w)

	consentURL := getConsentURL(r.URL.Query().Get("state"))
	http.Redirect(w, r, consentURL, http.StatusFound)
}

func getConsentURL(consentID string) string {
	return fmt.Sprintf("%s?consent=%s", pathConsent, consentID)
}

func (s *Server) session(r *http.Request) *sessions.Session {
	session, _ := s.store.Get(r, "identity_session")
	return session
}
