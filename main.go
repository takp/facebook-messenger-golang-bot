package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
	//"runtime"
	//"regexp"
	"strings"
)

var accessToken = os.Getenv("ACCESS_TOKEN")
var verifyToken = os.Getenv("VERIFY_TOKEN")
var port = os.Getenv("PORT")

const FacebookEndPoint = "https://graph.facebook.com/v2.6/me/messages"

type ReceivedMessage struct {
	Object string  `json:"object"`
	Entry  []Entry `json:"entry"`
}

type Entry struct {
	ID        int64       `json:"id"`
	Time      int64       `json:"time"`
	Messaging []Messaging `json:"messaging"`
}

type Messaging struct {
	Sender    Sender    `json:"sender"`
	Recipient Recipient `json:"recipient"`
	Timestamp int64     `json:"timestamp"`
	Message   Message   `json:"message"`
}

type Sender struct {
	ID int64 `json:"id"`
}

type Recipient struct {
	ID int64 `json:"id"`
}

type Message struct {
	MID  string `json:"mid"`
	Seq  int64  `json:"seq"`
	Text string `json:"text"`
}

type Payload struct {
	TemplateType string `json:"template_type"`
	Text string `json:"text"`
	Buttons Buttons `json:"buttons"`
}

type Buttons struct {
	Type string `json:"type"`
	Url string `json:"url"`
	Title string `json:"title"`
}

type Attachment struct {
	Type string `json:"type"`
	Payload Payload `json:"payload"`
}

type ButtonMessageBody struct {
	Attachment Attachment `json:"attachment"`
}

type ButtonMessage struct {
	Recipient Recipient `json:"recipient"`
	ButtonMessageBody ButtonMessageBody `json:"message"`
}

type SendMessage struct {
	Recipient Recipient `json:"recipient"`
	Message   struct {
			  Text string `json:"text"`
		  } `json:"message"`
}

func main() {
	http.HandleFunc("/", TopPageHandler)
	http.HandleFunc("/webhook", webhookHandler)
	address := fmt.Sprintf(":%s", port)
	http.ListenAndServe(address, nil)
}

func TopPageHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "This is go-bot application's top page.")
}

func webhookHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		verifyTokenAction(w, r)
	}
	if r.Method == "POST" {
		webhookPostAction(w, r)
	}
}

func verifyTokenAction(w http.ResponseWriter, r *http.Request) {
	if r.URL.Query().Get("hub.verify_token") == verifyToken {
		log.Print("verify token success.")
		fmt.Fprintf(w, r.URL.Query().Get("hub.challenge"))
	} else {
		log.Print("Error: verify token failed.")
		fmt.Fprintf(w, "Error, wrong validation token")
	}
}

func webhookPostAction(w http.ResponseWriter, r *http.Request) {
	var receivedMessage ReceivedMessage
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
	}
	if err = json.Unmarshal(body, &receivedMessage); err != nil {
		log.Print(err)
	}
	messagingEvents := receivedMessage.Entry[0].Messaging
	for _, event := range messagingEvents {
		senderID := event.Sender.ID
		if &event.Message != nil && event.Message.Text != "" {
			// TODO: Fix sendButtonMessage function
			//if messageForButton(event.Message.Text) {
			//	message := getReplyMessage(event.Message.Text)
			//	sendButtonMessage(senderID, message)
			//} else {
			//	message := getReplyMessage(event.Message.Text)
			//	sendTextMessage(senderID, message)
			//}
			message := getReplyMessage(event.Message.Text)
			sendTextMessage(senderID, message)
		}
	}
	fmt.Fprintf(w, "Success")
}

func messageForButton(text string) bool {
	return strings.Contains(text, "予約") || strings.Contains(text, "RESERVATION") || strings.Contains(text, "RESERVE") || strings.Contains(text, "BOOKING")
}

// TODO: Reply message is just sample and made by easy logic, need to enhance the logic.
func getReplyMessage(receivedMessage string) string {
	var message string
	receivedMessage = strings.ToUpper(receivedMessage)
	log.Print(" Received message: " + receivedMessage)

	if strings.Contains(receivedMessage, "予約") {
		message = "予約ありがとうございます。ご予約日を選んでください。"
	} else if strings.Contains(receivedMessage, "場所") {
		message = "BTSアソーク駅の近くです。"
	} else if strings.Contains(receivedMessage, "電話") {
		message = "02-123-4567"
	} else if strings.Contains(receivedMessage, "RESERVATION") {
		message = "Thank you for reservation! Choose the date."
	} else if strings.Contains(receivedMessage, "RESERVE") {
		message = "Thank you for reservation! Choose the date."
	} else if strings.Contains(receivedMessage, "BOOKING") {
		message = "Thank you for reservation! Choose the date."
	} else if strings.Contains(receivedMessage, "LOCATION") {
		message = "Near BTS Asok station."
	} else if strings.Contains(receivedMessage, "WHERE") {
		message = "Near BTS Asok station."
	} else if strings.Contains(receivedMessage, "TEL") {
		message = "02-123-4567"
	} else if receivedMessage == "HI" {
		message = "Hi! This is Golang Restaurant."
	} else if receivedMessage == "こんにちは" {
		message = "お問合せありがとうございます。Golangレストランです。"
	} else {
		message = "ちょっと意味が分かりません"
	}
	return message
}

// TODO: Fix Buttons json's format
func sendButtonMessage(senderID int64, text string) {
	recipient := new(Recipient)
	recipient.ID = senderID
	buttonMessage := new(ButtonMessage)
	buttonMessage.Recipient = *recipient
	buttonMessage.ButtonMessageBody.Attachment.Type = "template"
	buttonMessage.ButtonMessageBody.Attachment.Payload.TemplateType = "button"
	buttonMessage.ButtonMessageBody.Attachment.Payload.Text = text
	buttonMessage.ButtonMessageBody.Attachment.Payload.Buttons.Type = "web_url"
	buttonMessage.ButtonMessageBody.Attachment.Payload.Buttons.Url = "https://still-bayou-19762.herokuapp.com/"
	buttonMessage.ButtonMessageBody.Attachment.Payload.Buttons.Title = "May 13"
	log.Print(buttonMessage)
	buttonMessageBody, err := json.Marshal(buttonMessage)
	log.Print(buttonMessageBody)
	if err != nil {
		log.Print(err)
	}
	req, err := http.NewRequest("POST", FacebookEndPoint, bytes.NewBuffer(buttonMessageBody))
	if err != nil {
		log.Print(err)
	}
	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{Timeout: time.Duration(30 * time.Second)}
	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer res.Body.Close()
	var result map[string]interface{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Print(err)
	}
	log.Print(result)
}

func sendTextMessage(senderID int64, text string) {
	recipient := new(Recipient)
	recipient.ID = senderID
	send_message := new(SendMessage)
	send_message.Recipient = *recipient
	send_message.Message.Text = text
	send_message_body, err := json.Marshal(send_message)
	if err != nil {
		log.Print(err)
	}
	req, err := http.NewRequest("POST", FacebookEndPoint, bytes.NewBuffer(send_message_body))
	if err != nil {
		log.Print(err)
	}
	fmt.Println("%T", req)
	fmt.Println("%T", err)

	values := url.Values{}
	values.Add("access_token", accessToken)
	req.URL.RawQuery = values.Encode()
	req.Header.Add("Content-Type", "application/json; charset=UTF-8")
	client := &http.Client{Timeout: time.Duration(30 * time.Second)}
	res, err := client.Do(req)
	if err != nil {
		log.Print(err)
	}
	defer res.Body.Close()
	var result map[string]interface{}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Print(err)
	}
	if err := json.Unmarshal(body, &result); err != nil {
		log.Print(err)
	}
	log.Print(result)
}
