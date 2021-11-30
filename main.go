package main

import (
	"net/http"
	"encoding/json"
	"os"
	"log"

	"encoding/gob"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"

	"golang.org/x/oauth2"
)

var CLOUDMQTT_URL = os.Getenv("AR_CLOUDMQTT_URL")
var CLOUDMQTT_TOPICS = os.Getenv("AR_CLOUDMQTT_TOPICS")
var CLOUDMQTT_USER = os.Getenv("AR_CLOUDMQTT_USER")
var CLOUDMQTT_PWD = os.Getenv("AR_CLOUDMQTT_PWD")

var GOOGLE_AUTH_CLIENTID = os.Getenv("AR_GOOGLE_AUTH_CLIENTID")
var GOOGLE_AUTH_CLIENTSECRET = os.Getenv("AR_GOOGLE_AUTH_CLIENTSECRET")
var GOOGLE_AUTH_REDIRECTURL = os.Getenv("AR_GOOGLE_AUTH_REDIRECTURL")

const (
	GOOGLE_TOKEN = "GOOGLE_TOKEN"
)

type Robot struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Topic     string `json:"topic"`
}


var client mqtt.Client
var robots []Robot

func main() {
	initBrockerClient()

	gob.Register(oauth2.Token{})

	//init Gin
	r := gin.Default()

	//init Topic List (one topic per robot)
	err := json.Unmarshal([]byte(CLOUDMQTT_TOPICS), &robots)
	if err != nil {
		log.Println("Unable to parse TOPICS", err);
		return
	}

	//init user session store
	store := cookie.NewStore([]byte("secret"))
	r.Use(sessions.Sessions("mysession", store))

	//unsecured routes
	api := r.Group("/api")
	api.Use()
	{
		r.POST("/api/google_signin", googleSignin)
		r.POST("/api/google_token", googleToken)
	}

	//secured routes
	private := api.Group("/private")
	private.Use(AuthRequired)
	{
		private.GET("/user/infos", getUserInfo)
		private.GET("/command/:robotId/:cmd", command)
		private.GET("/robots", listRobots)
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
	robotId := c.Param("robotId")
	command := c.Param("cmd")

	for _, robot := range robots {
        if robot.Id == robotId {
			client.Publish(robot.Topic, 0, false, command)

			c.JSON(200, gin.H{
				"executed command": command,
			})
			return;
        }
    }

	c.JSON(http.StatusBadRequest, gin.H{"error": "unknown robot id"})
	
}

func listRobots(c *gin.Context) {
	c.JSON(200, robots)
}

