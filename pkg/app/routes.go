package app

import "net/http"

func (app *Application) SetRoutes() http.Handler {
	appMux := http.NewServeMux()
	appMux.HandleFunc("/", app.HIndex)

	// api routes
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/", app.HIndex)
	apiMux.HandleFunc("/news", app.HNews)
	apiMux.HandleFunc("/publications", app.HPublications)
	apiMux.HandleFunc("/notifications", app.HNotifications)
	apiMux.HandleFunc("/messages", app.HMessages)
	apiMux.HandleFunc("/users", app.HUsers) // followers, following, members
	apiMux.HandleFunc("/groups", app.HGroups)
	apiMux.HandleFunc("/chats", app.HChats)
	apiMux.HandleFunc("/events", app.HEvents)
	apiMux.HandleFunc("/comments", app.HComments)
	apiMux.HandleFunc("/files", app.HClippedFiles)
	apiMux.HandleFunc("/notification", app.HNotification)
	apiMux.HandleFunc("/gallery", app.HGallery)
	apiMux.HandleFunc("/post", app.HPost)
	apiMux.HandleFunc("/media", app.HMedia)
	apiMux.HandleFunc("/event", app.HEvent)
	apiMux.HandleFunc("/user", app.HUser)
	apiMux.HandleFunc("/group", app.HGroup)
	apiMux.HandleFunc("/search", app.HSearch)
	appMux.Handle("/api/", http.StripPrefix("/api", apiMux))

	// ws
	wsMux := http.NewServeMux()
	wsMux.HandleFunc("/", app.CreateWSUser)
	appMux.Handle("/ws/", http.StripPrefix("/ws", wsMux))

	// sign
	signMux := http.NewServeMux()
	signMux.HandleFunc("/", app.HIndex)
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

	// edit
	editMux := http.NewServeMux()
	editMux.HandleFunc("/", app.HIndex)
	editMux.HandleFunc("/settings", app.HChangeSettings)
	editMux.HandleFunc("/settings/c", app.HConfirmSettings)
	editMux.HandleFunc("/user", app.HChangeProfile)
	editMux.HandleFunc("/group", app.HChangeProfile)
	appMux.Handle("/e/", http.StripPrefix("/e", editMux))

	// save
	saveMux := http.NewServeMux()
	saveMux.HandleFunc("/", app.HIndex)
	saveMux.HandleFunc("/group", app.HSaveGroup)
	saveMux.HandleFunc("/post", app.HSavePost)
	saveMux.HandleFunc("/file", app.HSaveFile)
	saveMux.HandleFunc("/photo", app.HSaveMedia)
	saveMux.HandleFunc("/video", app.HSaveMedia)
	saveMux.HandleFunc("/like", app.HSaveLikeDislike)
	saveMux.HandleFunc("/event", app.HSaveEvent)
	saveMux.HandleFunc("/rlsh", app.HSaveRelation)
	saveMux.HandleFunc("/answer", app.HSaveEventAnswer)
	saveMux.HandleFunc("/chat", app.HSaveChat)
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
