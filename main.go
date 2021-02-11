package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var CLOUDMQTT_URL = os.Getenv("AR_CLOUDMQTT_URL")
var CLOUDMQTT_TOPIC = os.Getenv("AR_CLOUDMQTT_TOPIC")
var CLOUDMQTT_USER = os.Getenv("AR_CLOUDMQTT_USER")
var CLOUDMQTT_PWD = os.Getenv("AR_CLOUDMQTT_PWD")

var GOOGLE_AUTH_CLIENTID = os.Getenv("AR_GOOGLE_AUTH_CLIENTID")
var GOOGLE_AUTH_CLIENTSECRET = os.Getenv("AR_GOOGLE_AUTH_CLIENTSECRET")
var GOOGLE_AUTH_REDIRECTURL = os.Getenv("AR_GOOGLE_AUTH_REDIRECTURL")

const (
	GIN_USER_TOKEN = "GOOGLE_TOKEN"
)

var client mqtt.Client

func main() {
	//initBrockerClient()

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
	user := session.Get(GIN_USER_TOKEN)
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

func googleSignin(c *gin.Context) {

	var config = &oauth2.Config{
		ClientID:     GOOGLE_AUTH_CLIENTID,
		ClientSecret: GOOGLE_AUTH_CLIENTSECRET,
		Endpoint:     google.Endpoint,
		RedirectURL:  GOOGLE_AUTH_REDIRECTURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
	}

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)

	c.JSON(200, gin.H{
		"redirectTo": url,
	})

}

type GoogleTokenInput struct {
	Code string `json:"code" binding:"required"`
}

type GoogleUser struct {
	Sub        string
	Name       string
	GivenName  string
	FamilyName string
	Picture    string
	Locale     string
}

func googleToken(c *gin.Context) {
	session := sessions.Default(c)

	var input GoogleTokenInput
	c.ShouldBindJSON(&input)

	var config = &oauth2.Config{
		ClientID:     GOOGLE_AUTH_CLIENTID,
		ClientSecret: GOOGLE_AUTH_CLIENTSECRET,
		Endpoint:     google.Endpoint,
		RedirectURL:  GOOGLE_AUTH_REDIRECTURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
	}

	//get token
	tok, err := config.Exchange(oauth2.NoContext, input.Code)
	//log.Println(tok)

	if err != nil {
		log.Fatal(err)
		return
	}

	//get user infos
	response, err := config.Client(oauth2.NoContext, tok).Get("https://www.googleapis.com/oauth2/v3/userinfo")
	if err != nil {
		log.Fatal(err)
		return
	}

	defer response.Body.Close()

	var googleUser GoogleUser
	err2 := json.NewDecoder(response.Body).Decode(&googleUser)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err2.Error()})
		return
	}

	//get u
	//session.Set(GIN_USER_TOKEN, tok)
	session.Set(GIN_USER_TOKEN, "toto")

	if err := session.Save(); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, googleUser)

}

func initBrockerClient() {
	//uri, err := url.Parse(os.Getenv("CLOUDMQTT_URL"))
	uri, err := url.Parse(CLOUDMQTT_URL)
	if err != nil {
		log.Fatal(err)
	}

	client = connect("pub", uri)
}

func connect(clientId string, uri *url.URL) mqtt.Client {
	opts := createClientOptions(clientId, uri)
	client := mqtt.NewClient(opts)
	token := client.Connect()
	for !token.WaitTimeout(3 * time.Second) {
	}
	if err := token.Error(); err != nil {
		log.Fatal(err)
	}
	return client
}

func createClientOptions(clientId string, uri *url.URL) *mqtt.ClientOptions {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(fmt.Sprintf("tcp://%s", CLOUDMQTT_URL))
	opts.SetUsername(CLOUDMQTT_USER)
	opts.SetPassword(CLOUDMQTT_PWD)
	opts.SetClientID(clientId)
	return opts
}
