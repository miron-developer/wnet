package main

import (
	"crypto/tls"
	"os"
	"wnet/app"

	"fmt"
	"net/http"
	"time"
)

func routes(app *app.Application) http.Handler {
	appMux := http.NewServeMux()

	// api routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/news", app.HNews)
	apiMux.HandleFunc("/publications", app.HPublications)
	apiMux.HandleFunc("/notifications", app.HNotifications)
	apiMux.HandleFunc("/users", app.HUsers) // followers, following, members
	apiMux.HandleFunc("/groups", app.HGroups)
	apiMux.HandleFunc("/chats", app.HChats)
	apiMux.HandleFunc("/events", app.HEvents)
	apiMux.HandleFunc("/notification", app.HNotification)
	apiMux.HandleFunc("/gallery", app.HGallery)
	apiMux.HandleFunc("/post", app.HPost)
	apiMux.HandleFunc("/event", app.HEvent)
	apiMux.HandleFunc("/user", app.HUser)
	apiMux.HandleFunc("/group", app.HGroup)

	apiMux.HandleFunc("/posts", app.Hposts)
	apiMux.HandleFunc("/comments", app.Hcomments)
	apiMux.HandleFunc("/messages", app.Hmessages)
	apiMux.HandleFunc("/online", app.HonlineUsers)
	appMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// ws
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/", app.CreateWSUser)
	appMux.Handle("/ws/", http.StripPrefix("/ws", wsMux))

	// sign
	signMux := http.NewServeMux()
	signMux.HandleFunc("/up", app.HSignUp)
	signMux.HandleFunc("/s/", app.HSaveUser)
	signMux.HandleFunc("/in", app.HSignIn)
	signMux.HandleFunc("/status", app.HCheckUserLogged)
	signMux.HandleFunc("/re", app.HResetPassword)
	signMux.HandleFunc("/rst/", app.HSaveNewPassword)
	signMux.HandleFunc("/out", app.HLogout)
	signMux.HandleFunc("/oauth/up", app.HSignUp)
	signMux.HandleFunc("/oauth/in", app.HSignIn)
	appMux.Handle("/sign/", http.StripPrefix("/sign", signMux))

	// // profile
	// profileMux := http.NewServeMux()
	// profileMux.HandleFunc("/change-avatar", app.HChangeAvatar)
	// profileMux.HandleFunc("/change-profile", app.HChangeData)
	// profileMux.HandleFunc("/", app.Hindex)
	// appMux.Handle("/profile/", http.StripPrefix("/profile", profileMux))

	// edit
	editMux := http.NewServeMux()
	editMux.HandleFunc("/settings", app.HChangeSettings)
	editMux.HandleFunc("/settings/c", app.HConfirmSettings)
	editMux.HandleFunc("/user", app.HChangeProfile)
	editMux.HandleFunc("/group", app.HChangeProfile)
	appMux.Handle("/e/", http.StripPrefix("/e", editMux))

	// save
	saveMux := http.NewServeMux()
	saveMux.HandleFunc("/group", app.HSaveGroup)
	saveMux.HandleFunc("/post", app.HSavePost)
	saveMux.HandleFunc("/file", app.HSaveFile)
	saveMux.HandleFunc("/photo", app.HSaveMedia)
	saveMux.HandleFunc("/video", app.HSaveMedia)
	saveMux.HandleFunc("/like", app.HSaveLikeDislike)
	saveMux.HandleFunc("/event", app.HSaveEvent)
	saveMux.HandleFunc("/rlsh", app.HSaveRelation)
	saveMux.HandleFunc("/answer", app.HSaveEventAnswer)

	saveMux.HandleFunc("/message", app.HSaveMessage)
	saveMux.HandleFunc("/comment", app.HSaveComment)
	appMux.Handle("/s/", http.StripPrefix("/s", saveMux))

	// assets get
	appMux.HandleFunc("/assets/", app.HGetFile)

	// middlewares
	muxHanlder := app.AccessLogMiddleware(appMux)
	muxHanlder = app.SecureHeaderMiddleware(muxHanlder)
	return muxHanlder
}

func main() {
	port := os.Getenv("PORT")
	app := app.InitProg()

	if port != "" {
		app.Port = port
	}

	app.ILog.Println("initialization completed!")

	// check sessions expire per minute
	go app.CheckPerMin()
	// websocket work
	go app.WSWork()

	// server
	srv := http.Server{
		Addr:         ":" + app.Port,
		ErrorLog:     app.ELog,
		Handler:      routes(app),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305},
		},
	}

	fmt.Printf("server listening on ports %v(HTTPS) and %v(HTTP)\n", app.Port, "8080")
	app.ILog.Printf("server listening on ports %v(HTTPS) and %v(HTTP)", app.Port, "8080")

	// localhost side
	go func() {
		app.ELog.Println(http.ListenAndServe(":8080", http.HandlerFunc(nil)))
	}()
	app.ELog.Fatal(srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem"))

	// heroku side
	// app.ELog.Fatal(srv.ListenAndServe())
}
