package app

import (
	"net/http"
	"strconv"
	"time"
	"wnet/pkg/orm"

	"github.com/gorilla/websocket"
)

// exported consts
const (
	// WSMessage types
	WSM_USER_ONLINE_TYPE           = 1  // when some user online
	WSM_USER_OFFLINE_TYPE          = 2  // when some user offline
	WSM_ADD_USER_NOTIFICATION_TYPE = 3  // when new notification
	WSM_ADD_NAVS_NOTIFICATION_TYPE = 4  // when new notification
	WSM_CHAT_MESSAGE_TYPE          = 10 // when send chat message
	WSM_TYPING_START_TYPE          = 11 // user typing
	WSM_TYPING_FINISH_TYPE         = 12 // user finish typing
	WSM_GET_CALLED_TYPE            = 20 // when audio video call
	WSM_USER_NOT_ACCESSIBLE        = 21 // when user already in call
	WSM_CLOSE_CALL_TYPE            = 22 // when close call
	ErrorType                      = -1 // if error was

	// for ws ping pong work & connections
	WSC_WRITE_WAIT  = 10 * time.Second
	WSC_PONG_WAIT   = 60 * time.Second
	WSC_PING_PERIOD = (WSC_PONG_WAIT * 9) / 10

	// for chat rooms
	// WSR_ROOM_P2P   = 20
	// WSR_ROOM_GROUP = 21
)

// WSMessage one message from ws connection to users
type WSMessage struct {
	MsgType     int         `json:"msgType"`   // ws message type
	AddresserID string      `json:"addresser"` // who sended, if = -1, then send server
	ReceiverID  string      `json:"receiver"`  // who get, if = -2, then get all, -1 server
	Body        interface{} `json:"body"`      // message body
}

// WSUser is one ws connection user
type WSUser struct {
	Conn *websocket.Conn
	ID   int
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin:     func(*http.Request) bool { return true },
}

func (app *Application) findUserByID(finder int) *WSUser {
	for _, user := range app.OnlineUsers {
		if user.ID == finder {
			return user
		}
	}
	return nil
}

// WSWork work with channel
func (app *Application) WSWork() {
	for {
		msg := <-app.WSMessages
		if msg.ReceiverID == "all" {
			for _, v := range app.OnlineUsers {
				go v.Conn.WriteJSON(msg)
			}
		} else {
			receiverID, recErr := strconv.Atoi(msg.ReceiverID)
			addresserID, addErr := strconv.Atoi(msg.AddresserID)
			if recErr != nil || addErr != nil {
				continue
			}

			receiver := app.findUserByID(receiverID)
			addresser := app.findUserByID(addresserID)
			if receiver == nil || (msg.AddresserID == "" && addresser == nil) {
				continue
			}

			// futher acting
			go func() {
				if e := receiver.Conn.WriteJSON(msg); e != nil {
					addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
						Body: "something wrong: " + e.Error(), MsgType: ErrorType})
				}
			}()
		}
	}
}

// HandleUserMsg handle received msg from front user
func (user *WSUser) HandleUserMsg(app *Application) {
	user.Conn.SetPongHandler(
		func(string) error {
			user.Conn.SetReadDeadline(time.Now().Add(WSC_PONG_WAIT))
			return nil
		},
	)

	for {
		msg := &WSMessage{}
		if e := user.Conn.ReadJSON(msg); e != nil {
			// if user.ID > 0 {
			// 	app.WSMessages <- &WSMessage{MsgType: WSM_USER_OFFLINE_TYPE, AddresserID: "server", ReceiverID: "all", Body: user}
			// }

			app.m.Lock()
			delete(app.OnlineUsers, user.ID)
			app.m.Unlock()
			app.ELog.Println(e)
			user.Conn.Close()
			u := orm.User{ID: user.ID, Status: strconv.Itoa(int(time.Now().Unix() * 1000))}
			u.Change()
			return
		}
		msg.AddresserID = strconv.Itoa(user.ID)
		app.WSMessages <- msg
	}
}

// Pinger ping every pingPeriod
func (user *WSUser) Pinger() {
	ticker := time.NewTicker(WSC_PING_PERIOD)
	for {
		<-ticker.C
		if err := user.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			return
		}
	}
}
