package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/twilio/twilio-go/client/jwt"
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
	router.POST("/voice/incomming-calls", handleInboundCall)

	router.Run(":3000")
}

func generateToken(c *gin.Context) {
	accountSid := os.Getenv("TWILIO_ACCOUNT_SID")
	apiKey := os.Getenv("TWILIO_API_KEY")
	apiSecret := os.Getenv("TWILIO_API_SECRET")
	applicationSid := os.Getenv("TWILIO_TWIML_APP_SID")
	identity := "+815012345678"

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

func handleInboundCall(c *gin.Context) {
	twiml := `<?xml version="1.0" encoding="UTF-8"?>
<Response>
    <Dial>
        <Client>+815012345678</Client>
    </Dial>
</Response>`

	c.Header("Content-Type", "text/xml")
	c.String(http.StatusOK, twiml)
}
