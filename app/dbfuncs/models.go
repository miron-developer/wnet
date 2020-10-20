/*
	This file is some "graphql" models
*/

package dbfuncs

// Users - table Users
type Users struct {
	ID            int    `json:"id"`
	NickName      string `json:"nick"`
	FirstName     string `json:"firstName"`
	LastName      string `json:"lastName"`
	Age           int    `json:"age"`
	Gender        string `json:"gender"`
	Photo         string `json:"photo"`
	Carma         int    `json:"carma"`
	Role          string `json:"role"`
	Friends       string `json:"friends"`
	LastActivitie string `json:"lastActivitie"`
	Email         string `json:"email"`
	Password      string
}

// Sessions - table Sessions
type Sessions struct {
	ID     string
	Expire string
	UserID int
}

// Posts - table Posts
type Posts struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Body    string `json:"body"`
	Date    string `json:"date"`
	Carma   int    `json:"carma"`
	Changed int    `json:"changed"`
	Tags    string `json:"tags"`
	UserID  int
}

// Comments - table Comments
type Comments struct {
	ID        int    `json:"id"`
	Body      string `json:"body"`
	Date      string `json:"date"`
	Carma     int    `json:"carma"`
	HaveChild int    `json:"haveChild"`
	Changed   int    `json:"changed"`
	UserID    int
	PostID    int
	CommentID int
}

// LDGeneral - table LD for all
type LDGeneral struct {
	ID        int `json:"id"`
	Value     int `json:"value"`
	UserID    int
	WhosID    int
	TableName string
}

// Messages - table Messages
type Messages struct {
	ID         int    `json:"id"`
	Date       string `json:"date"`
	Body       string `json:"body"`
	Status     int    `json:"status"`
	SenderID   int    `json:"sender"`
	ReceiverID int    `json:"receiver"`
}
