package orm

// User - table
type User struct {
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

// Group - table
type Group struct {
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

// Session - table
type Session struct {
	ID     string
	Expire string
	UserID int
}

// Post - table
type Post struct {
	ID                 int    `json:"id"`
	Type               string `json:"type"`
	Title              string `json:"title"`
	Body               string `json:"body"`
	UnixDate           int    `json:"datetime"`
	PostType           string `json:"postType"`
	AllowedUsers       string `json:"allowedUsers"`
	IsHaveClippedFiles string `json:"isHaveClippedFiles"`
	UserID             int    `json:"userID"`
	GroupID            int    `json:"groupID"`
}

// Message - table
type Message struct {
	ID              int    `json:"id"`
	UnixDate        int    `json:"datetime"`
	MessageType     string `json:"messageType"`
	Body            string `json:"body"`
	SenderUserID    int    `json:"senderUserID"`
	ReceiverUserID  int    `json:"receiverUserID"`
	ReceiverGroupID int    `json:"receiverGroupID"`
}

// Chat - table
type Chat struct {
	ID              int    `json:"id"`
	ChatType        string `json:"type"`
	Users           string `json:"users"`
	Closed          string `json:"closed"`
	SenderUserID    int    `json:"senderUserID"`
	ReceiverUserID  int    `json:"receiverUserID"`
	ReceiverGroupID int    `json:"receiverGroupID"`
}

// Event - table
type Event struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Title    string `json:"title"`
	About    string `json:"description"`
	UnixDate int    `json:"datetime"`
	UserID   int    `json:"userID"`
	GroupID  int    `json:"groupID"`
}

// EventAnswer - table
type EventAnswer struct {
	ID      int    `json:"id"`
	Answer  string `json:"answer"`
	UserID  int    `json:"userID"`
	EventID int    `json:"eventID"`
}

// Relation - table
type Relation struct {
	ID              int    `json:"id"`
	Value           string `json:"value"`
	SenderUserID    int    `json:"senderUserID"`
	SenderGroupID   int    `json:"senderGroupID"`
	ReceiverUserID  int    `json:"receiverUserID"`
	ReceiverGroupID int    `json:"receiverGroupID"`
}

// Media - table
type Media struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	MediaType string `json:"type"`
	UnixDate  int    `json:"datetime"`
	Source    string `json:"src"`
	Preview   string `json:"preview"`
	UserID    int    `json:"userID"`
	GroupID   int    `json:"groupID"`
}

// Comment - table
type Comment struct {
	ID                 int    `json:"id"`
	Body               string `json:"body"`
	UnixDate           int    `json:"datetime"`
	IsHaveChild        string `json:"isHaveChild"`
	IsAnswer           string `json:"isAnswer"`
	IsHaveClippedFiles string `json:"isHaveClippedFiles"`
	UserID             int    `json:"userID"`
	PostID             int    `json:"postID"`
	CommentID          int    `json:"commentID"`
	MediaID            int    `json:"mediaID"`
}

// Notification - table
type Notification struct {
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

// Like - table
type Like struct {
	ID        int `json:"id"`
	UserID    int `json:"userID"`
	PostID    int `json:"postID"`
	CommentID int `json:"commentID"`
	MediaID   int `json:"mediaID"`
}

// ClippedFile - table
type ClippedFile struct {
	ID        int    `json:"id"`
	FileType  string `json:"type"`
	Source    string `json:"src"`
	Name      string `json:"filename"`
	UserID    int    `json:"userID"`
	PostID    int    `json:"postID"`
	CommentID int    `json:"commentID"`
	MessageID int    `json:"messageID"`
}
