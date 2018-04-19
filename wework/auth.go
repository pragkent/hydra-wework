package wework

import (
	"fmt"
	"net/url"
)

func (c *Client) GetOAuthURL(redirectURI, state string) string {
	q := url.Values{}
	q.Set("appid", c.corpID)
	q.Set("agentid", c.agentID)
	q.Set("redirect_uri", redirectURI)
	q.Set("state", state)

	return "https://open.work.weixin.qq.com/wwopen/sso/qrConnect?" + q.Encode()
}

type GetUserInfoResponse struct {
	Code    int    `json:"errcode,omitempty"`
	Message string `json:"errmsg,omitempty"`
	UserID  string `json:"UserId,omitempty"`
}

func (c *Client) GetUserInfo(code string) (string, error) {
	q := url.Values{}
	q.Set("code", code)

	u := "https://qyapi.weixin.qq.com/cgi-bin/user/getuserinfo?" + q.Encode()

	var resp GetUserInfoResponse
	if err := c.getJSON(u, &resp); err != nil {
		return "", err
	}

	if resp.Code != 0 {
		return "", fmt.Errorf("Get user info error: %v %v", resp.Code, resp.Message)
	}

	return resp.UserID, nil
}
