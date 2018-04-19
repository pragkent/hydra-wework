package wework

import (
	"fmt"
	"net/url"
)

type GetUserResponse struct {
	Code        int    `json:"errcode,omitempty"`
	Message     string `json:"errmsg,omitempty"`
	UserID      string `json:"userid,omitempty"`
	EnglishName string `json:"english_name,omitempty"`
	Email       string `json:"email,omitempty"`
}

func (c *Client) GetUser(uid string) (*GetUserResponse, error) {
	q := url.Values{}
	q.Set("userid", uid)

	u := "https://qyapi.weixin.qq.com/cgi-bin/user/get?" + q.Encode()

	var resp GetUserResponse
	if err := c.getJSON(u, &resp); err != nil {
		return nil, err
	}

	if resp.Code != 0 {
		return nil, fmt.Errorf("Get user error: %v %v", resp.Code, resp.Message)
	}

	return &resp, nil
}
