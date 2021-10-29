package subscription

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/url"
	"testing"
)

type SubCreateRequest struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CallbackUrl  string `json:"callback_url"`
	VerifyToken  string `json:"verify_token"`
}

type Subscription struct {
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"`
	CallbackUrl  string `json:"callback_url"`
	VerifyToken  string `json:"verify_token"`
	client       *http.Client
}

func NewSubscription(clientID, clientSecret, callbackUrl string, client *http.Client) Subscription {

	return Subscription{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackUrl:  callbackUrl,
		VerifyToken:  uuid.New().String(),
		client:       client,
	}
}

type RoundTripperFunc func(*http.Request) (*http.Response, error)

func (fn RoundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return fn(req)
}

func (s Subscription) ViewSubscriptions() string {
	reqUrl := fmt.Sprintf("https://www.strava.com/api/v3/push_subscriptions?client_id=%s&client_secret=%s", s.ClientID, s.ClientSecret)
	resp, err := s.client.Get(reqUrl)
	if err != nil {
		fmt.Printf("error getting subscriptions %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("got statuscode not like 200, statuscode: %s", resp.StatusCode)
	}
	//resp.Body
	return ""
}

func Test_list_of_subscriptions(t *testing.T) {
	client := http.Client{}
	client.Transport = RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		rep := http.Response{
			Status:     "OK",
			Body:       ioutil.NopCloser(bytes.NewBufferString(`[]`)),
			StatusCode: http.StatusOK,
			Header:     http.Header{},
		}
		rep.Header.Set("Content-Type", "application/json; charset=utf-8")

		return &rep, nil
	})
	s := NewSubscription("cid", "cs", "callback", &client)
	result := s.ViewSubscriptions()
	assert.Equal(t, "[]", result)
}

/*
1) view list of subscription. If there are no subscription create new subscription
2) Call subscription create request with callbackUrl
3) callbackUrl will get a GET request with verify_token and challenge
4) callback need to response to GET with 200, application/json and body with json {"hub.challenge": "id"}
*/

func NewSubCreateRequest(clientID, clientSecret, callbackUrl, verifyToken string) SubCreateRequest {

	return SubCreateRequest{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		CallbackUrl:  callbackUrl,
		VerifyToken:  verifyToken,
	}
}

func Test_create_subscription_request(t *testing.T) {
	hostUrl := "https://www.strava.com/api/v3/push_subscriptions"
	clientID := "5"
	clientSecret := "topSecret***"
	callbackUrl := "http://localhost"
	verifyToken := "strava"

	subReq := NewSubCreateRequest(clientID, clientSecret, callbackUrl, verifyToken)

	bodyData, err := json.Marshal(subReq)
	assert.NoError(t, err)

	client := http.Client{}
	client.Transport = RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
		if req.Body == nil {
			t.Error("Body == nil; want a Body")
		}
		rep := http.Response{
			Status:     "OK",
			StatusCode: http.StatusOK,
		}
		return &rep, nil
	})
	req, err := http.NewRequest(http.MethodPost, hostUrl, bytes.NewBuffer(bodyData))
	assert.NoError(t, err)
	reps, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, 200, reps.StatusCode)
}

type HubValidation struct {
	Mode        string `json:"hub.mode"`
	Challenge   string `json:"hub.challenge"`
	VerifyToken string `json:"hub.verify_token"`
}

func Test_subscription_callback_validation(t *testing.T) {

	testReq := "https://mycallbackurl.com?hub.verify_token=STRAVA&hub.challenge=15f7d1a91c1f40f8a748fd134752feb3&hub.mode=subscribe"
	parsReq, err := url.Parse(testReq)
	assert.NoError(t, err)
	assert.Equal(t, "mycallbackurl.com", parsReq.Host)
	m, err := url.ParseQuery(parsReq.RawQuery)

	assert.NoError(t, err)
	assert.NotNil(t, m)
	assert.True(t, m.Has("hub.mode"))
	assert.True(t, m.Has("hub.challenge"))
	assert.True(t, m.Has("hub.verify_token"))

	challenge := m.Get("hub.challenge")
	assert.Equal(t, "15f7d1a91c1f40f8a748fd134752feb3", challenge)
}

/*
Your callback address must respond within two seconds to the GET request from Strava’s subscription service.
The response should indicate status code 200 and should echo
the hub.challenge field in the response body as application/json content type:
{ “hub.challenge”:”15f7d1a91c1f40f8a748fd134752feb3” }

https://developers.strava.com/docs/webhooks/
*/
// $ GET https://mycallbackurl.com?hub.verify_token=STRAVA&hub.challenge=15f7d1a91c1f40f8a748fd134752feb3&hub.mode=subscribe
