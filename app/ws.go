package app

import (
	"net/http"
	"strconv"
	"time"
	"wnet/app/dbfuncs"

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
	// Rooms map[string]*ChatRoom
}

// ChatRoom is room between users
// type ChatRoom struct {
// 	RoomID          string
// 	Type            int // group chat or p2p
// 	Users           map[int]*WSUser
// 	Messages        chan *WSMessage
// 	LastMsgUnixTime int64
// }

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

				// if msg.MsgType == WSM_CHAT_MESSAGE_TYPE {
				// 	body := dbfuncs.MakeArrFromStruct(msg.Body)
				// 	chatMsg := &dbfuncs.Message{
				// 		UnixDate:       int(time.Now().Unix()),
				// 		Body:           body[0].(string),
				// 		SenderUserID:   addresserID,
				// 		ReceiverUserID: receiverID,
				// 		MessageType:    body[1].(string),
				// 	}
				// 	if body[2].(string) == "group" {
				// 		chatMsg.ReceiverUserID = 0
				// 		chatMsg.ReceiverGroupID = receiverID
				// 	}
				// 	if e := chatMsg.Create(); e != nil {
				// 		addresser.Conn.WriteJSON(
				// 			WSMessage{
				// 				AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
				// 				Body: "can not create message: " + e.Error(), MsgType: ErrorType,
				// 			},
				// 		)
				// 	}
				// }
			}()
		}
	}
}

// func (room *ChatRoom) Work(app *Application) {
// 	for {
// 		msg := WSMessage{}
// 		if msg.ReceiverID == "all" {
// 			for _, v := range room.Users {
// 				v.Conn.WriteJSON(msg)
// 			}
// 		} else {
// 			receiverID, recErr := strconv.Atoi(msg.ReceiverID)
// 			addresserID, addErr := strconv.Atoi(msg.AddresserID)
// 			if recErr != nil || addErr != nil {
// 				continue
// 			}

// 			receiver := app.findUserByID(receiverID)
// 			addresser := app.findUserByID(addresserID)
// 			if receiver == nil || (msg.AddresserID == "" && addresser == nil) {
// 				continue
// 			}

// 			if msg.MsgType == WSM_CHAT_MESSAGE_TYPE {
// 				body := dbfuncs.MakeArrFromStruct(msg.Body)
// 				chatMsg := &dbfuncs.Message{
// 					UnixDate:     int(time.Now().Unix()),
// 					Body:         body[0].(string),
// 					SenderUserID: addresserID,
// 					MessageType:  body[1].(string),
// 				}
// 				if room.Type == WSR_ROOM_GROUP {
// 					chatMsg.ReceiverGroupID = receiverID
// 				}
// 				if e := chatMsg.Create(); e != nil {
// 					addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
// 						Body: "something wrong: " + e.Error(), MsgType: ErrorType})
// 				}
// 			}

// 			if e := receiver.Conn.WriteJSON(msg); e != nil {
// 				addresser.Conn.WriteJSON(WSMessage{AddresserID: "server", ReceiverID: strconv.Itoa(addresser.ID),
// 					Body: "something wrong: " + e.Error(), MsgType: ErrorType})
// 			}
// 		}
// 	}
// }

// func (app *Application) CreateRoom(roomType int, roomID string, users map[int]*WSUser) {
// 	room := &ChatRoom{
// 		RoomID:   roomID, // for ex p2p id = u:1-2 (less userID first), or group id = g:1
// 		Type:     roomType,
// 		Users:    users,
// 		Messages: make(chan *WSMessage),
// 	}

// 	app.m.Lock()
// 	app.ChatRooms[room.RoomID] = room
// 	app.m.Unlock()
// 	go room.Work(app)

// 	for _, v := range users {
// 		v.Rooms[roomID] = room
// 	}
// }

// func (app *Application) CloseRoom(roomID string) {
// 	delete(app.ChatRooms, roomID)
// }

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
			u := dbfuncs.User{ID: user.ID, Status: strconv.Itoa(int(time.Now().Unix() * 1000))}
			u.Change()
			return
		}

		// roomID := dbfuncs.MakeArrFromStruct(msg.Body)[2].(string)
		// if user.Rooms != nil && roomID != "" {
		// 	user.Rooms[roomID].Messages <- msg
		// 	user.Rooms[roomID].LastMsgUnixTime = time.Now().Unix()
		// 	return
		// }
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
