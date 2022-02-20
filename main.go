package main

import (
	"encoding/json"
	"log"
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
var CLOUDMQTT_TOPICS = os.Getenv("AR_CLOUDMQTT_TOPICS")
var CLOUDMQTT_USER = os.Getenv("AR_CLOUDMQTT_USER")
var CLOUDMQTT_PWD = os.Getenv("AR_CLOUDMQTT_PWD")

var GOOGLE_AUTH_CLIENTID = os.Getenv("AR_GOOGLE_AUTH_CLIENTID")
var GOOGLE_AUTH_CLIENTSECRET = os.Getenv("AR_GOOGLE_AUTH_CLIENTSECRET")
var GOOGLE_AUTH_REDIRECTURL = os.Getenv("AR_GOOGLE_AUTH_REDIRECTURL")

var JITSI_APP_ID = os.Getenv("AR_JITSI_APP_ID")
var JITSI_API_KEY = os.Getenv("AR_JITSI_API_KEY")

const (
	GOOGLE_TOKEN = "GOOGLE_TOKEN"
)

type Robot struct {
	Id    string `json:"id"`
	Name  string `json:"name"`
	Topic string `json:"topic"`
}

type RobotConfiguration struct {
	JitsiToken string
	RoomName   string
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
		log.Println("Unable to parse TOPICS", err)
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
		private.GET("/robots", listRobots)
		private.GET("/robot/init/:robotId", initRobotConfiguration)
		private.GET("/robot/:robotId/command/:cmd", command)
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

func initRobotConfiguration(c *gin.Context) {
	robotId := c.Param("robotId")
	robot := getRobotById(robotId)

	//init Jitsi session for robot camera
	robotJwtToken, err := initJitsiSession(robot.Name, robot.Id)
	if err != nil {
		log.Fatal(err)
		return
	}
	robotConfiguration := new(RobotConfiguration)
	robotConfiguration.JitsiToken = robotJwtToken
	robotConfiguration.RoomName = JITSI_APP_ID + "/" + robot.Id

	//send Configuration to the robot
	//client.Publish(robot.Topic, 0, false, robotConfiguration)

	//init Jitsi session for end user
	userJwtToken, err := initJitsiSession("William Ferreira", robot.Id)
	if err != nil {
		log.Fatal(err)
		return
	}
	userRobotConfiguration := new(RobotConfiguration)
	userRobotConfiguration.JitsiToken = userJwtToken
	userRobotConfiguration.RoomName = JITSI_APP_ID + "/" + robot.Id

	c.JSON(200, userRobotConfiguration)
}

func command(c *gin.Context) {
	robotId := c.Param("robotId")
	command := c.Param("cmd")

	robot := getRobotById(robotId)

	client.Publish(robot.Topic, 0, false, command)

	c.JSON(200, gin.H{
		"executed command": command,
	})

	//TODO c.JSON(http.StatusBadRequest, gin.H{"error": "unknown robot id"})

}

func listRobots(c *gin.Context) {
	c.JSON(200, robots)
}

func getRobotById(robotId string) Robot {
	for _, robot := range robots {
		if robot.Id == robotId {
			return robot
		}
	}
	return Robot{}
}
