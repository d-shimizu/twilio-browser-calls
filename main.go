package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go"
	"github.com/twilio/twilio-go/client/jwt"
	twilioApi "github.com/twilio/twilio-go/rest/api/v2010"
)

type VoiceResponse struct {
	Identity string `json:"identity"`
	Token    string `json:"token"`
}

type TwilioGrant struct {
	OutgoingApplicationSid string                 `json:"outgoing_application_sid,omitempty"`
	IncomingAllow          bool                   `json:"incoming_allow,omitempty"`
	OutgoingAllow          bool                   `json:"outgoing_allow,omitempty"`
	Params                 map[string]interface{} `json:"params,omitempty"`
}

func main() {
	router := gin.Default()

	router.Static("/static", "./static")
	router.LoadHTMLGlob("templates/*")

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", nil)
	})

	router.POST("/voice/token", generateToken)
	router.POST("/voice/incomming-calls", handleIncommingCall)
	router.POST("/voice/outbound-calls", handleOutboundCall)

	router.Run(":3000")
}

func generateToken(c *gin.Context) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")
	applicationSid := os.Getenv("TWILIO_TWIML_APP_SID")
	identity := "user123"

	params := jwt.AccessTokenParams{
		AccountSid:    accountSid,
		SigningKeySid: apiKey,
		Secret:        apiSecret,
		Identity:      identity,
	}

	jwtToken := jwt.CreateAccessToken(params)
	voiceGrant := &jwt.VoiceGrant{
		Incoming: jwt.Incoming{Allow: true},
		Outgoing: jwt.Outgoing{
			ApplicationSid: applicationSid,
		},
	}

	jwtToken.AddGrant(voiceGrant)
	token, err := jwtToken.ToJwt()
	if err != nil {
		fmt.Println(err)
	}

	c.JSON(http.StatusOK, VoiceResponse{
		Identity: identity,
		Token:    token,
	})
}

func handleIncommingCall(c *gin.Context) {
	twiml := `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Say language="ja-JP">着信がありました</Say>
    <Dial timeout="20">
        <Client>user123</Client>
    </Dial>
    <Say language="ja-JP">応答がありませんでした</Say>
</Response>`

	//c.Header("Content-Type", "application/xml")
	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, twiml)
}

func handleOutboundCall(c *gin.Context) {
	from := os.Getenv("TWILIO_FROM_PHONE_NUMBER")
	//from := c.PostForm("From")
	to := c.PostForm("To")
	if to == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "To parameter is required"})
		return
	}

	client := twilio.NewRestClient()
	params := &twilioApi.CreateCallParams{}
	params.SetTo(to)
	params.SetFrom(from)
	params.SetUrl(fmt.Sprintf("%s/voice/callbacks", os.Getenv("APP_BASE_URL")))

	resp, err := client.Api.CreateCall(params)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Call Status: " + *resp.Status)
		fmt.Println("Call Sid: " + *resp.Sid)
		fmt.Println("Call Direction: " + *resp.Direction)
	}
}
