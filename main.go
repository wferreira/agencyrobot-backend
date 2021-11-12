package main

import (
	"net/http"
	"os"

	"encoding/gob"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
)

var CLOUDMQTT_URL = os.Getenv("AR_CLOUDMQTT_URL")
var CLOUDMQTT_TOPIC = os.Getenv("AR_CLOUDMQTT_TOPIC")
var CLOUDMQTT_USER = os.Getenv("AR_CLOUDMQTT_USER")
var CLOUDMQTT_PWD = os.Getenv("AR_CLOUDMQTT_PWD")

var GOOGLE_AUTH_CLIENTID = os.Getenv("AR_GOOGLE_AUTH_CLIENTID")
var GOOGLE_AUTH_CLIENTSECRET = os.Getenv("AR_GOOGLE_AUTH_CLIENTSECRET")
var GOOGLE_AUTH_REDIRECTURL = os.Getenv("AR_GOOGLE_AUTH_REDIRECTURL")

const (
	GOOGLE_TOKEN = "GOOGLE_TOKEN"
)

var client mqtt.Client

func main() {
	//initBrockerClient()

	gob.Register(oauth2.Token{})

	//init Gin
	r := gin.Default()

	//init user session store
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	//unsecured routes
	r.POST("/api/google_signin", googleSignin)
	r.POST("/api/google_token", googleToken)

	//secured routes
	private := r.Group("/private")
	private.Use(AuthRequired)
	{
		private.GET("/api/command/:cmd", command)
	}

	r.Run()

}

// AuthRequired is a simple middleware to check the session
func AuthRequired(c *gin.Context) {
	session := sessions.Default(c)
	user := session.Get(GOOGLE_TOKEN)
	if user == nil {
		// Abort the request with the appropriate error code
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	// Continue down the chain to handler etc
	c.Next()
}

func command(c *gin.Context) {
	command := c.Param("cmd")

	client.Publish(CLOUDMQTT_TOPIC, 0, false, command)

	c.JSON(200, gin.H{
		"executed command": command,
	})
}

