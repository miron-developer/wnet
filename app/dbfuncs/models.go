package dbfuncs

// Users - table
type Users struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	NickName  string `json:"nickname"`
	FirstName string `json:"fName"`
	LastName  string `json:"lName"`
	Dob       string `json:"dob"`
	Age       int    `json:"age"`
	Gender    string `json:"gender"`
	Avatar    string `json:"avatar"`
	Status    string `json:"status"`
	About     string `json:"aboutMe"`
	IsPrivate string `json:"isPrivate"`
	Email     string `json:"email"`
	Password  string
}

// Groups - table
type Groups struct {
	ID           int    `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	CreationDate string `json:"cdate"`
	Age          int    `json:"age"`
	Avatar       string `json:"avatar"`
	About        string `json:"description"`
	IsPrivate    string `json:"isPrivate"`
	OwnerUserID  int    `json:"ownerUserID"`
}

// Sessions - table
type Sessions struct {
	ID     string
	Expire string
	UserID int
}

// Posts - table
type Posts struct {
	ID           int    `json:"id"`
	Type         string `json:"type"`
	Title        string `json:"title"`
	Body         string `json:"body"`
	UnixDate     int    `json:"datetime"`
	PostType     string `json:"postType"`
	AllowedUsers string `json:"allowedUsers"`
	UserID       int    `json:"userID"`
	GroupID      int    `json:"groupID"`
}

// Messages - table
type Messages struct {
	ID              int    `json:"id"`
	UnixDate        int    `json:"datetime"`
	MessageType     string `json:"messageType"`
	Body            string `json:"body"`
	SenderUserID    int    `json:"senderUserID"`
	ReceiverUserID  int    `json:"receiverUserID"`
	ReceiverGroupID int    `json:"receiverGroupID"`
}

// Events - table
type Events struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	About    string `json:"description"`
	UnixDate int    `json:"datetime"`
	UserID   int    `json:"userID"`
	GroupID  int    `json:"groupID"`
}

// EventAnswers - table
type EventAnswers struct {
	ID      int    `json:"id"`
	Answer  string `json:"answer"`
	UserID  int    `json:"userID"`
	EventID int    `json:"eventID"`
}

// Relations - table
type Relations struct {
	ID             int    `json:"id"`
	Value          string `json:"value"`
	SenderUserID   int    `json:"senderUserID"`
	ReceiverUserID int    `json:"receiverUserID"`
}

// Media - table
type Media struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	MediaType string `json:"mediaType"`
	UnixDate  int    `json:"datetime"`
	Source    string `json:"src"`
	UserID    int    `json:"userID"`
	GroupID   int    `json:"groupID"`
}

// Comments - table
type Comments struct {
	ID          int    `json:"id"`
	Body        string `json:"body"`
	UnixDate    int    `json:"datetime"`
	IsHaveChild string `json:"isHaveChild"`
	IsAnswer    string `json:"isAnswer"`
	UserID      int    `json:"userID"`
	PostID      int    `json:"postID"`
	CommentID   int    `json:"commentID"`
	MediaID     int    `json:"mediaID"`
}

// Notifications - table
type Notifications struct {
	ID               int    `json:"id"`
	UnixDate         int    `json:"datetime"`
	NotificationType string `json:"notificationType"`
	SenderUserID     int    `json:"senderUserID"`
	ReceiverUserID   int    `json:"receiverUserID"`
	PostID           int    `json:"postID"`
	CommentID        int    `json:"commentID"`
	EventID          int    `json:"eventID"`
	GroupID          int    `json:"groupID"`
	MediaID          int    `json:"mediaID"`
}

// Likes - table
type Likes struct {
	ID        int `json:"id"`
	UserID    int `json:"userID"`
	PostID    int `json:"postID"`
	CommentID int `json:"commentID"`
	MediaID   int `json:"mediaID"`
}

// ClippedFiles - table
type ClippedFiles struct {
	ID        int    `json:"id"`
	FileType  string `json:"fileType"`
	UserID    int    `json:"userID"`
	PostID    int    `json:"postID"`
	CommentID int    `json:"commentID"`
	MessageID int    `json:"messageID"`
}
