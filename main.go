package main

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
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

var client mqtt.Client

func main() {
	initBrockerClient()

	//initGoogleOAuth()

	r := gin.Default()

	r.GET("/api/command/:cmd", command)
	r.POST("/api/signin", signin)

	r.Run()

}

func command(c *gin.Context) {
	command := c.Param("cmd")

	client.Publish(CLOUDMQTT_TOPIC, 0, false, command)

	c.JSON(200, gin.H{
		"executed command": command,
	})
}

func signin(c *gin.Context) {

	var config = &oauth2.Config{
		ClientID:     GOOGLE_AUTH_CLIENTID,
		ClientSecret: GOOGLE_AUTH_CLIENTSECRET,
		Endpoint:     google.Endpoint,
		RedirectURL:  GOOGLE_AUTH_REDIRECTURL,
		Scopes:       []string{"openid"},
	}

	url := config.AuthCodeURL("state", oauth2.AccessTypeOffline)
	fmt.Printf("Visit the URL for the auth dialog: %v", url)

	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
	//c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
	//c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
	//c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

	c.JSON(200, gin.H{
		"redirectTo": url,
	})

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
