/*
	Initialize app
*/

package app

import (
	"log"
	"os"
	"sync"
	"wnet/app/dbfuncs"
)

// Application this is app struct and items
type Application struct {
	m                   sync.Mutex
	ELog                *log.Logger
	ILog                *log.Logger
	Port                string
	CurrentRequestCount int
	MaxRequestCount     int
	IsHeroku            bool
	UsersCode           map[string]*dbfuncs.User
	RestoreCode         map[string]string
	ChangeCode          map[string]*dbfuncs.User
	OnlineUsers         map[int]*WSUser
	WSMessages          chan *WSMessage
	// ChatRooms           map[string]*ChatRoom
}

// InitProg initialise
func InitProg() *Application {
	logFile, _ := os.Create("logs.txt")

	elog := log.New(logFile, "\033[31m[ERROR]\033[0m\t", log.Ldate|log.Ltime|log.Lshortfile)
	info := log.New(logFile, "\033[34m[INFO]\033[0m\t", log.Ldate|log.Ltime|log.Lshortfile)
	info.Println("loggers is done!")

	dbfuncs.InitDB(elog)
	info.Println("db completed!")

	return &Application{
		ELog:                elog,
		ILog:                info,
		Port:                "4330",
		CurrentRequestCount: 0,
		MaxRequestCount:     1200,
		IsHeroku:            false,
		UsersCode:           map[string]*dbfuncs.User{},
		RestoreCode:         map[string]string{},
		ChangeCode:          map[string]*dbfuncs.User{},
		OnlineUsers:         map[int]*WSUser{},
		WSMessages:          make(chan *WSMessage),
		// ChatRooms:           map[string]*ChatRoom{},
	}
}
