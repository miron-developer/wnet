/*
	Initialize app
*/

package app

import (
	"anifor/app/dbfuncs"
	"log"
	"os"
	"path/filepath"
	"sync"
	"text/template"
)

// Application this is app struct and items
type Application struct {
	m                   sync.Mutex // mutex
	ELog                *log.Logger
	ILog                *log.Logger
	HTTPport            string
	HTTPSport           string
	CachedTemplates     map[string]*template.Template
	CurrentRequestCount int
	MaxRequestCount     int
	UsersCode           map[string]*dbfuncs.Users
	RestoreCode         map[string]string
	OnlineUsers         map[string]*WSUser // online users
	Messages            chan *WSMessage
}

// InitProg initialise
func InitProg() *Application {
	wd, _ := os.Getwd()
	logFile, _ := os.Create("logs.txt")

	elog := log.New(logFile, "\033[31m[ERROR]\033[0m\t", log.Ldate|log.Ltime|log.Lshortfile)
	info := log.New(logFile, "\033[34m[INFO]\033[0m\t", log.Ldate|log.Ltime|log.Lshortfile)
	info.Println("loggers is done!")

	cachedTmp, e := initCachedTemplates(wd)
	if e != nil {
		elog.Fatal(e)
		return nil
	}
	info.Println("cached templates is done!")

	dbfuncs.InitDB(elog)
	info.Println("db completed!")

	return &Application{
		ELog:                elog,
		ILog:                info,
		CachedTemplates:     cachedTmp,
		HTTPSport:           "4330",
		HTTPport:            "8080",
		CurrentRequestCount: 0,
		MaxRequestCount:     1200,
		UsersCode:           map[string]*dbfuncs.Users{},
		RestoreCode:         map[string]string{},
		OnlineUsers:         map[string]*WSUser{},
		Messages:            make(chan *WSMessage),
	}
}

func initCachedTemplates(dir string) (map[string]*template.Template, error) {
	cache := map[string]*template.Template{}
	pages, e := filepath.Glob(filepath.Join(dir+"/static/", "*.html"))
	if e != nil {
		return nil, e
	}

	for _, v := range pages {
		name := filepath.Base(v)
		t, e := template.ParseFiles(dir + "/static/" + name)
		if e != nil {
			return nil, e
		}
		cache[name] = t
	}
	return cache, nil
}
