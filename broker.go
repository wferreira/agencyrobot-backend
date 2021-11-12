package main

import (
	"fmt"
	"log"
	"net/url"
	"time"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)


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