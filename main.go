package main

import (
	"anifor/app"
	"os"

	"crypto/tls"
	"fmt"
	"net/http"
	"time"
)

func routes(app *app.Application) http.Handler {
	appMux := http.NewServeMux()
	appMux.HandleFunc("/", app.Hindex)

	// api routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/posts", app.Hposts)
	apiMux.HandleFunc("/post/", app.Hpost)
	apiMux.HandleFunc("/comments", app.Hcomments)
	apiMux.HandleFunc("/users", app.Husers)
	apiMux.HandleFunc("/messages", app.Hmessages)
	apiMux.HandleFunc("/online", app.HonlineUsers)
	apiMux.HandleFunc("/user/", app.Huser)
	apiMux.HandleFunc("/", app.Hindex)
	appMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// ws
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/", app.CreateWSUser)
	wsMux.HandleFunc("/change-name", app.ChangeName)
	appMux.Handle("/ws/", http.StripPrefix("/ws", wsMux))

	// sign
	signMux := http.NewServeMux()
	signMux.HandleFunc("/s/", app.HSaveUser)
	signMux.HandleFunc("/in", app.HSignIn)
	signMux.HandleFunc("/up", app.HSignUp)
	signMux.HandleFunc("/r/", app.HSavePassword)
	signMux.HandleFunc("/restore", app.HRestore)
	signMux.HandleFunc("/logout", app.HLogout)
	signMux.HandleFunc("/status", app.HcheckUserLogged)
	signMux.HandleFunc("/", app.Hindex)
	appMux.Handle("/sign/", http.StripPrefix("/sign", signMux))

	// profile
	profileMux := http.NewServeMux()
	profileMux.HandleFunc("/change-avatar", app.HChangeAvatar)
	profileMux.HandleFunc("/change-profile", app.HChangeData)
	profileMux.HandleFunc("/", app.Hindex)
	appMux.Handle("/profile/", http.StripPrefix("/profile", profileMux))

	// save
	saveMux := http.NewServeMux()
	saveMux.HandleFunc("/image", app.HSaveImage)
	saveMux.HandleFunc("/file", app.HSaveFile)
	saveMux.HandleFunc("/post", app.HSavePost)
	saveMux.HandleFunc("/message", app.HSaveMessage)
	saveMux.HandleFunc("/comment", app.HSaveComment)
	saveMux.HandleFunc("/ld", app.HSaveLikeDislike)
	saveMux.HandleFunc("/", app.Hindex)
	appMux.Handle("/save/", http.StripPrefix("/save", saveMux))

	// static files define
	static := http.FileServer(http.Dir("static"))
	appMux.Handle("/static/", http.StripPrefix("/static/", static))

	// middlewares
	muxHanlder := app.AccessLogMiddleware(appMux)
	muxHanlder = app.SecureHeaderMiddleware(muxHanlder)
	return muxHanlder
}

func main() {
	port := os.Getenv("PORT")
	app := app.InitProg()

	if port != "" {
		app.HTTPSport = port
	}

	app.ILog.Println("initialization completed!")

	// check sessions expire per minute
	go app.CheckPerMin()
	// websocket work
	go app.WSWork()

	// server
	srv := http.Server{
		Addr:         ":" + app.HTTPSport,
		ErrorLog:     app.ELog,
		Handler:      routes(app),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  120 * time.Second,
		TLSConfig: &tls.Config{
			PreferServerCipherSuites: true,
			CurvePreferences:         []tls.CurveID{tls.X25519, tls.CurveP256},
			CipherSuites: []uint16{tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
				tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256, tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305, tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305},
		},
	}

	fmt.Printf("server listening on ports %v(HTTPS) and %v(HTTP)\n", app.HTTPSport, app.HTTPport)
	app.ILog.Printf("server listening on ports %v(HTTPS) and %v(HTTP)", app.HTTPSport, app.HTTPport)

	// localhost side
	// go func() {
	// 	app.ELog.Println(http.ListenAndServe(":"+app.HTTPport, http.HandlerFunc(app.HTTPRedirect)))
	// }()
	// app.ELog.Fatal(srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem"))

	// heroku side
	app.ELog.Fatal(srv.ListenAndServe())
}
