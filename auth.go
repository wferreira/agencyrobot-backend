package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/gin-contrib/sessions"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

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

	session.Set(GOOGLE_TOKEN, tok)

	getUserInfo(c)
}

func getUserInfo(c *gin.Context) {
	session := sessions.Default(c)

	var config = &oauth2.Config{
		ClientID:     GOOGLE_AUTH_CLIENTID,
		ClientSecret: GOOGLE_AUTH_CLIENTSECRET,
		Endpoint:     google.Endpoint,
		RedirectURL:  GOOGLE_AUTH_REDIRECTURL,
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.profile"},
	}

	tok := session.Get(GOOGLE_TOKEN).(*oauth2.Token)

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

	if err := session.Save(); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to save session"})
		return
	}
	c.JSON(http.StatusOK, googleUser)

}