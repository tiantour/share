package wechat

import (
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/google/go-querystring/query"
	"github.com/tiantour/fetch"
	"github.com/tiantour/imago"
	"github.com/tiantour/rsae"
)

var (
	// AppID appid
	AppID string

	// AppSecret app secret
	AppSecret string
)

// Share share
type Share struct {
	AppID       string `json:"appid" url:"-"`
	JSapiTicket string `json:"jsapi_ticket" url:"jsapi_ticket"`
	Noncestr    string `json:"noncestr" url:"noncestr"`
	Timestamp   string `json:"timestamp" url:"timestamp"`
	URL         string `json:"url" url:"url"`
	Signature   string `json:"signature" url:"-"`
}

// Ticket ticket
type Ticket struct {
	Ticket    string `json:"ticket"`
	ExpiresIn int    `json:"expires_in"`
	ErrCode   int    `json:"errcode"`
	ErrMsg    string `json:"errmsg"`
}

// NewShare new share
func NewShare() *Share {
	return &Share{}
}

// Message message
func (s *Share) Message(url string) (*Share, error) {
	ticket, err := s.Ticket()
	if err != nil {
		return nil, err
	}
	result := Share{
		Noncestr:  imago.NewRandom().Text(16),
		Timestamp: strconv.FormatInt(time.Now().Unix(), 10),
		URL:       url,
	}
	result.JSapiTicket = ticket.Ticket
	sign, err := s.sign(&result)
	if err != nil {
		return nil, err
	}
	result.Signature = sign
	result.AppID = AppID
	return &result, nil
}

// Ticket ticket
func (s *Share) Ticket() (*Ticket, error) {
	token, err := NewToken().Access()
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/ticket/getticket?access_token=%s&type=jsapi",
		token,
	)
	return s.do(url)
}

// do
func (s *Share) do(url string) (*Ticket, error) {
	result := Ticket{}
	body, err := fetch.Cmd(fetch.Request{
		Method: "GET",
		URL:    url,
	})
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return nil, err
	}
	if result.ErrCode != 0 {
		return nil, errors.New(result.ErrMsg)
	}
	return &result, err
}

// sign
func (s *Share) sign(args *Share) (string, error) {
	params, err := query.Values(args)
	if err != nil {
		return "", err
	}
	query, err := url.QueryUnescape(params.Encode())
	if err != nil {
		return "", err
	}
	sign := rsae.NewSHA().SHA1(query)
	return hex.EncodeToString(sign), nil
}
