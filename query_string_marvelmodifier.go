package querystring

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/google/martian"
	"github.com/google/martian/parse"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

func init() {
	parse.Register("querystring.MarvelModifier", marvelModifierFromJSON)
}

// MarvelModifier contains the private and public Marvel API key
type MarvelModifier struct {
	public, private string
}

// MarvelModifierJSON to Unmarshal the JSON configuration
type MarvelModifierJSON struct {
	Public  string               `json:"public"`
	Private string               `json:"private"`
	Scope   []parse.ModifierType `json:"scope"`
}

// ModifyRequest modifies the query string of the request with the given key and value.
func (m *MarvelModifier) ModifyRequest(req *http.Request) error {
	query := req.URL.Query()
	header := req.Header
	header.Add("sso","sso")
	ts := strconv.FormatInt(time.Now().Unix(), 10)
	hash := GetMD5Hash(ts + m.private + m.public)
	query.Set("apikey", m.public)
	query.Set("ts", ts)
	query.Set("hash", hash)
	//for key, value := range header{
	//	fmt.Println(key)
	//	fmt.Println(value)
	//}
	fmt.Println("判断是否带了sso的请求头")
	if _, ok := header["X-SSO-FullticketId"]; !ok{
		return errors.New("没有带sso的请求头")
	}
	fmt.Println(header["X-SSO-FullticketId"])
	req.URL.RawQuery = query.Encode()
	return nil
}

// GetMD5Hash returns the md5 hash from a string
func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

// MarvelNewModifier returns a request modifier that will set the query string
// at key with the given value. If the query string key already exists all
// values will be overwritten.
func MarvelNewModifier(public, private string) martian.RequestModifier {
	return &MarvelModifier{
		public:  public,
		private: private,
	}
}

// marvelModifierFromJSON takes a JSON message as a byte slice and returns
// a querystring.modifier and an error.
//
// Example JSON:
// {
//  "public": "apikey",
//  "private": "apikey",
//  "scope": ["request", "response"]
// }
func marvelModifierFromJSON(b []byte) (*parse.Result, error) {
	msg := &MarvelModifierJSON{}

	if err := json.Unmarshal(b, msg); err != nil {
		return nil, err
	}

	return parse.NewResult(MarvelNewModifier(msg.Public, msg.Private), msg.Scope)
}
func Check(){
	//sso_ticket := ctx.GetHeader(config.AuthHeader)
	sso_ticket_header := "X-SSO-FullticketId"
	//if sso_ticket_header == "" {
	//	return
	//}

	userinfo, err := ssoGetUserModel(sso_ticket_header)
	if err != nil {

		return
	}

	if userinfo.ErrorCode != 0 {
		return
	}
}

func ssoGetUserModel(ticket string) (*SsoTicketUserInfoResponse, error) {
	//token_url := config.SSOCheckUserUrl + "/api/v2/info"
	token_url := "http://10.200.60.36:8800/api/v2/info"

	client := &http.Client{}
	v := url.Values{}
	//pass the values to the request's body
	req, err := http.NewRequest("GET", token_url, strings.NewReader(v.Encode()))
	req.Header.Set("ticket", ticket)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var token_response SsoTicketUserInfoResponse
	err = json.Unmarshal(body, &token_response)
	return &token_response, nil

}

type SsoTicketUserInfoResponse struct {
	ErrorCode int      `json:"errorCode"`
	Data      *SsoData `json:"data"`
	Message   string   `json:"message"`
}

type SsoData struct {
	LoginEmail  string         `json:"LoginEmail"`
	AccountGuid string         `json:"AccountGuid"`
	DisplayName string         `json:"DisplayName"`
}