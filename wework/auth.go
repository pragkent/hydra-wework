package wework

import (
	"errors"
	"fmt"
	"net/url"
)

const (
	oauthURL     = "https://open.weixin.qq.com/connect/oauth2/authorize"
	qrConnectURL = "https://open.work.weixin.qq.com/wwopen/sso/qrConnect"
	userInfoURL  = "https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo"
)

func (c *Client) GetQRConnectURL(redirectURI, state string) string {
	q := url.Values{}
	q.Set("appid", c.corpID)
	q.Set("agentid", c.agentID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)

	return fmt.Sprintf("%s?%s", qrConnectURL, q.Encode())
}

func (c *Client) GetOAuthURL(redirectURI, state string) string {
	q := url.Values{}
	q.Set("appid", c.corpID)
	q.Set("redirect_uri", redirectURI)
	q.Set("response_type", "code")
	q.Set("scope", "snsapi_base")
	q.Set("agentid", c.agentID)
	q.Set("state", state)

	return fmt.Sprintf("%s?%s#wechat_redirect", oauthURL, q.Encode())
}

type GetUserInfoResponse struct {
	Code    int    `json:"errcode,omitempty"`
	Message string `json:"errmsg,omitempty"`
	UserID  string `json:"UserId,omitempty"`
}

func (c *Client) GetUserInfo(code string) (string, error) {
	q := url.Values{}
	q.Set("code", code)

	u := fmt.Sprintf("%s?%s", userInfoURL, q.Encode())

	var resp GetUserInfoResponse
	if err := c.getJSON(u, &resp); err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("Get user info error: %v %v", resp.Code, resp.Message)
	}

	if resp.UserID == "" {
		return "", errors.New("User is not wework member")
	}

	return resp.UserID, nil
}
