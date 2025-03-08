package twitch

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

type MessageType string

const (
	Welcome      MessageType = "session_welcome"
	KeepAlive    MessageType = "session_keepalive"
	Notification MessageType = "notification"
	Reconnect    MessageType = "session_reconnect"
	Revocation   MessageType = "revocation"
)

type WebsocketMessage struct {
	Metadata struct {
		MessageID           string `json:"message_id"`
		MessageType         string `json:"message_type"`
		MessageTimestamp    string `json:"message_timestamp"`
		SubscriptionType    string `json:"subscription_type"`
		SubscriptionVersion string `json:"subscription_version"`
	}
	Payload struct {
		Session      Session      `json:"session"`
		Subscription Subscription `json:"subscription"`
		Event        Event        `json:"event"`
	} `json:"payload"`
}

type Session struct {
	ID                      string `json:"id"`
	Status                  string `json:"status"`
	ConnectedAt             string `json:"connected_at"`
	KeepAliveTimeoutSeconds int    `json:"keepalive_timeout_seconds"`
	ReconnectURL            string `json:"reconnect_url"`
}

type Subscription struct {
	ID        string      `json:"id"`
	Status    string      `json:"status"`
	Type      string      `json:"type"`
	Version   string      `json:"version"`
	Cost      int         `json:"cost"`
	Condition interface{} `json:"condition"`
	Transport struct {
		Method    string `json:"method"`
		SessionID string `json:"session_id"`
	} `json:"transport"`
	CreatedAt string `json:"created_at"`
}

type Event struct {
}

type ClientCredentials struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	TokenType   string `json:"token_type"`
}

type TwitchService struct {
	ws *websocket.Conn
	id string
}

func New() *TwitchService {
	return &TwitchService{
		ws: nil,
		id: "",
	}
}

func (t *TwitchService) SubscribeChannel() {

}

func (t *TwitchService) ensureConnect(ch chan string) {
	if t.ws != nil {
		ch <- t.id
		return
	}
	ws, _, err := websocket.DefaultDialer.Dial("wss://eventsub.wss.twitch.tv/ws", nil)
	if err != nil {
		log.Println(err)
	}
	t.ws = ws

	ws.SetCloseHandler(func(code int, text string) error {
		log.Printf("Close: %d %s", code, text)
		t.ws = nil
		return nil
	})

	go func() {
		for {
			_, message, err := ws.ReadMessage()
			if err != nil {
				log.Println(err)
				return
			}

			var msg WebsocketMessage
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println(err)
				return
			}

			switch MessageType(msg.Metadata.MessageType) {
			case Welcome:
				log.Println("Welcome")
				t.id = msg.Payload.Session.ID
				ch <- t.id
				close(ch)
			case KeepAlive:
				log.Println("KeepAlive")
			case Notification:
				log.Println("Notification")
			case Reconnect:
				log.Println("Reconnect")
			case Revocation:
				log.Println("Revocation")
			default:
				log.Println("Unknown message type")
			}
		}
	}()

	t.ws = ws
}
