package ledger

import (
	"net/http"
	"time"
)

type Token string

func (t Token) apply(r *http.Request) {
	r.Header.Set("X-Auth-Token", string(t))
}

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Remember bool   `json:"remember"`
}

type SessionInfo struct {
	ID       string        `json:"id"`
	UserID   string        `json:"user_id"`
	LoggedAt time.Time     `json:"logged_at"`
	TTL      time.Duration `json:"ttl"`
}

type LoginResponse struct {
	Token   Token       `json:"token"`
	User    User        `json:"user"`
	Session SessionInfo `json:"session"`
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

func (c Client) Login(data Credentials) (*LoginResponse, error) {
	rsp := new(LoginResponse)
	return rsp, c.post("/auth", data, rsp, "")
}

func (c Client) Register(data RegisterRequest) (*LoginResponse, error) {
	rsp := new(LoginResponse)
	return rsp, c.post("/auth/register", data, rsp, "")
}

func (c Client) Session(t Token) (*SessionInfo, error) {
	rsp := new(SessionInfo)
	return rsp, c.get("/auth/session", rsp, t)
}

func (c Client) Logout(t Token) error {
	return c.delete("/auth/session", t)
}
