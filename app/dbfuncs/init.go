/*
	Initialize db
*/

package dbfuncs

import (
	"context"
	"database/sql"
	"log"

	// sqlite driver
	_ "github.com/mattn/go-sqlite3"
)

// ConnToDB is var to conn to db
var ConnToDB *sql.DB

func warn(eLogger *log.Logger, e error) {
	if e != nil {
		eLogger.Fatal(e)
	}
}

func initTables(eLogger *log.Logger) {
	_, e := ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS Users (
															ID INTEGER PRIMARY KEY AUTOINCREMENT, 
															nickname TEXT UNIQUE, firstName TEXT, lastName TEXT, age INTEGER, gender TEXT, photo TEXT,
															carma INTEGER, role INTEGER, friends TEXT, lastActivitie TEXT, email TEXT UNIQUE, password TEXT,
															CHECK(age > 0 AND role >= 0 AND gender IN("Мужской", "Женский") AND email LIKE '%@%.%' AND lastActivitie IS strftime('%Y-%m-%d %H:%M:%S', lastActivitie))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF not EXISTS Sessions (
															ID TEXT PRIMARY KEY, expire TEXT, userID INTEGER UNIQUE,
															FOREIGN KEY (userID) REFERENCES Users(ID) ON DELETE CASCADE,
															CHECK(expire IS strftime('%Y-%m-%d %H:%M:%S', expire))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS Posts (
															ID INTEGER PRIMARY KEY AUTOINCREMENT,
															title TEXT, body TEXT, date TEXT, carma INTEGER,
															changed INTEGER, tags TEXT, userID INTEGER,
															FOREIGN KEY (userID) REFERENCES Users(ID) ON DELETE CASCADE,
															CHECK(date IS strftime('%Y-%m-%d %H:%M:%S', date) AND changed IN (0, 1))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS Comments (
															ID INTEGER PRIMARY KEY AUTOINCREMENT,
															body TEXT, date TEXT, carma INTEGER, haveChild INTEGER, changed INTEGER,
															userID INTEGER, postID INTEGER, commentID INTEGER,
															FOREIGN KEY (userID) REFERENCES Users(ID) ON DELETE CASCADE,
															FOREIGN KEY (postID) REFERENCES Posts(ID) ON DELETE CASCADE,
															FOREIGN KEY (commentID) REFERENCES Comments(ID) ON DELETE CASCADE,
															CHECK(date IS strftime('%Y-%m-%d %H:%M:%S', date) AND haveChild IN (0, 1) AND changed IN (0, 1))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS LDPosts
															(ID INTEGER PRIMARY KEY AUTOINCREMENT,
															value INTEGER, userID INTEGER, postID INTEGER,
															FOREIGN KEY (userID) REFERENCES Users(ID) ON DELETE CASCADE, 
															FOREIGN KEY (postID) REFERENCES Posts(ID) ON DELETE CASCADE,
															CHECK(value IN (-1, 0, 1))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS LDComments
															(ID INTEGER PRIMARY KEY AUTOINCREMENT, 
															value INTEGER, userID INTEGER, commentID INTEGER,
															FOREIGN KEY (userID) REFERENCES Users(ID) ON DELETE CASCADE, 
															FOREIGN KEY (commentID) REFERENCES Comments(ID) ON DELETE CASCADE,
															CHECK(value IN (-1, 0, 1))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS LDUsers
															(ID INTEGER PRIMARY KEY AUTOINCREMENT, 
															value INTEGER, userID INTEGER, whosID INTEGER,
															FOREIGN KEY (userID) REFERENCES Users(ID) ON DELETE CASCADE, 
															FOREIGN KEY (whosID) REFERENCES Users(ID) ON DELETE CASCADE,
															CHECK(value IN (-1, 0, 1))
														)`)
	warn(eLogger, e)

	_, e = ConnToDB.ExecContext(context.Background(), `CREATE TABLE IF NOT EXISTS Messages
															(ID INTEGER PRIMARY KEY AUTOINCREMENT,
															date TEXT, body TEXT, status INTEGER,
															senderID INTEGER, receiverID INTEGER,
															FOREIGN KEY (senderID) REFERENCES Users(ID) ON DELETE CASCADE,
															FOREIGN KEY (receiverID) REFERENCES Users(ID) ON DELETE CASCADE,
															CHECK(date IS strftime('%Y-%m-%d %H:%M:%S', date) AND status IN (1, 0))
														)`)
	warn(eLogger, e)
}

// InitDB init db, settings and tables
func InitDB(eLogger *log.Logger) {
	ConnToDB, _ = sql.Open("sqlite3", "file:db/anifor.db?_auth&_auth_user=anifor&_auth_pass=anifor&_auth_crypt=sha1")

	_, e := ConnToDB.ExecContext(context.Background(), "PRAGMA foreign_keys = ON;PRAGMA case_sensitive_like = true;PRAGMA auto_vacuum = FULL;")
	warn(eLogger, e)

	initTables(eLogger)
}
