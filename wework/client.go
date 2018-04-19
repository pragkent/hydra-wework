package wework

import (
	"sync"
	"time"
)

type Client struct {
	corpID      string
	agentID     string
	agentSecret string
	tokenHolder *tokenHolder
}

type tokenHolder struct {
	mu        *sync.Mutex
	token     string
	expiresAt time.Time
}

func NewClient(corpID, agentID, agentSecret string) *Client {
	return &Client{
		corpID:      corpID,
		agentID:     agentID,
		agentSecret: agentSecret,
		tokenHolder: &tokenHolder{
			mu: &sync.Mutex{},
		},
	}
}
