package main

import (
	"crypto/tls"
	"os"
	"wnet/pkg/app"

	"fmt"
	"net/http"
	"time"
)

func main() {
	port := os.Getenv("PORT")
	app := app.InitProg()

	if port != "" {
		app.IsHeroku = true
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
		Handler:      app.SetRoutes(),
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

	fmt.Printf("server listening on port %v\n", app.Port)
	app.ILog.Printf("server listening on port %v\n", app.Port)

	if app.IsHeroku {
		app.ELog.Fatal(srv.ListenAndServe())
		return
	}
	app.ELog.Fatal(srv.ListenAndServeTLS("./tls/cert.pem", "./tls/key.pem"))
}
