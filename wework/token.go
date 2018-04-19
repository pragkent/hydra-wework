package wework

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

type GetAccessTokenResponse struct {
	Code        int    `json:"errcode,omitempty"`
	Message     string `json:"errmsg,omitempty"`
	AccessToken string `json:"access_token,omitempty"`
	ExpiresIn   int    `json:"expires_in,omitempty"`
}

func (c *Client) refreshAccessToken() (string, error) {
	c.tokenHolder.mu.Lock()
	defer c.tokenHolder.mu.Unlock()

	now := time.Now()
	if now.Before(c.tokenHolder.expiresAt) {
		return c.tokenHolder.token, nil
	}

	resp, err := c.requestAccessToken()
	if err != nil {
		return "", err
	}

	c.tokenHolder.token = resp.AccessToken
	c.tokenHolder.expiresAt = now.Add(time.Duration(resp.ExpiresIn) * time.Second)

	return c.tokenHolder.token, err
}

func (c *Client) requestAccessToken() (*GetAccessTokenResponse, error) {
	q := url.Values{}
	q.Set("corpid", c.corpID)
	q.Set("corpsecret", c.agentSecret)

	u, _ := url.Parse("https://qyapi.weixin.qq.com/cgi-bin/gettoken")
	u.RawQuery = q.Encode()

	httpResp, err := http.Get(u.String())
	if err != nil {
		return nil, fmt.Errorf("http.Get error: %v", err)
	}

	defer httpResp.Body.Close()

	body, err := ioutil.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("http response read error: %v", err)
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("illegal http status code: %v", httpResp.StatusCode)
	}

	var resp GetAccessTokenResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("json.Unmarshal error: %v", err)
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("token.get api error: %+v", resp)
	}

	return &resp, nil
}
