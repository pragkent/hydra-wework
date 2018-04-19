package server

import (
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

const requiredScope = "openid"

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

	srv := &Server{
		cfg:   c,
		mux:   mux.NewRouter(),
		hcli:  hcli,
		wcli:  wework.NewClient(c.WeworkCorpID, c.WeworkAgentID, c.WeworkSecret),
		store: sessions.NewCookieStore([]byte(c.SessionKey)),
	}

	srv.mux.HandleFunc("/consent", srv.ConsentHandler)
	srv.mux.HandleFunc("/wework/auth", srv.AuthHandler)
	srv.mux.HandleFunc("/wework/callback", srv.CallbackHandler)

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
		http.Redirect(w, r, authURL(r), http.StatusFound)
		return
	}

	extraVars := make(map[string]interface{})
	if err := s.getTokenExtraVars(uid, extraVars); err != nil {
		log.Printf("Get token extra vars error: %v", err)
		http.Error(w, "Get token extra vars error", http.StatusInternalServerError)
		return
	}

	response, err = s.hcli.AcceptOAuth2ConsentRequest(reqID, swagger.ConsentRequestAcceptance{
		Subject:          uid,
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

func authURL(r *http.Request) string {
	return "/wework/auth?consent=" + consentID(r)
}

func (s *Server) scheme() string {
	if s.cfg.HTTPS {
		return "https"
	} else {
		return "http"
	}
}

func (s *Server) getTokenExtraVars(uid string, vars map[string]interface{}) error {
	userResp, err := s.wcli.GetUser(uid)
	if err != nil {
		return fmt.Errorf("Get wework user failed. %v", err)
	}

	vars["username"] = userResp.UserID
	vars["email"] = userResp.Email
	vars["name"] = userResp.EnglishName

	gs, _, err := s.hcli.ListGroups(uid, 100, 0)
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
	callbackURL := fmt.Sprintf("%s://%s/wework/callback", s.scheme(), r.Host)
	state := r.URL.Query().Get("consent")

	u := s.wcli.GetOAuthURL(callbackURL, state)
	http.Redirect(w, r, u, http.StatusFound)
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

	u := "/consent?consent=" + r.URL.Query().Get("state")
	http.Redirect(w, r, u, http.StatusFound)
}

func (s *Server) session(r *http.Request) *sessions.Session {
	session, _ := s.store.Get(r, "identity_session")
	return session
}
