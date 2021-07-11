CREATE TABLE IF NOT EXISTS Users (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
	nName TEXT UNIQUE, -- nickname
    fName TEXT NOT NULL, -- firts name
    lName TEXT NOT NULL, -- last name
    dob TEXT, -- date of birth
    age INTEGER,
    gender TEXT,
    ava TEXT NOT NULL, -- avatar
    status TEXT NOT NULL, -- online or unix timestamp
    about TEXT NOT NULL,
    isPrivate INTEGER,
    email TEXT UNIQUE,
    password TEXT NOT NULL,
	CHECK(
        type = "user" AND
        age > 13 AND
        gender IN("Male", "Female", "Default") AND
        email LIKE '%@%.%' AND
        dob IS strftime('%Y-%m-%d', dob) AND
        LENGTH(about) <= 400 AND
        isPrivate IN (0, 1)
    )
);

CREATE TABLE IF NOT EXISTS Groups (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
	title TEXT NOT NULL,
    cdate TEXT, -- creation date
    age INTEGER, -- how long exist this group
    ava TEXT NOT NULL, -- avatar
    about TEXT NOT NULL,
    isPrivate INTEGER NOT NULL,
    ownerUserID INTEGER,
    FOREIGN KEY (ownerUserID) REFERENCES Users(id) ON DELETE CASCADE,
    CHECK(
        type = "group" AND
        age >= 0 AND
        cdate IS strftime('%Y-%m-%d', cdate) AND
        LENGTH(about) <= 400 AND
        isPrivate IN (0, 1)
    )
);

CREATE TABLE IF not EXISTS Sessions (
	id TEXT PRIMARY KEY,
    expire TEXT,
    userID INTEGER UNIQUE,
	FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
	CHECK(
        expire IS strftime('%Y-%m-%d %H:%M:%S', expire)
    )
);

CREATE TABLE IF NOT EXISTS Posts (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
	title TEXT NOT NULL,
    body TEXT NOT NULL,
    datetime INTEGER NOT NULL,
    postType TEXT,
    allowedUsers TEXT,
    isHaveClippedFiles INTEGER,
    userID INTEGER,
    groupID INTEGER,
	FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (groupID) REFERENCES Groups(id) ON DELETE CASCADE,
	CHECK(
        type = "post" AND
        postType IN("public", "private", "almost_private") AND
        isHaveClippedFiles IN (0, 1)
    )
);

CREATE TABLE IF NOT EXISTS Messages (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	datetime INTEGER NOT NULL,
    type TEXT,
    body TEXT NOT NULL,
	senderUserID INTEGER,
    receiverUserID INTEGER,
    receiverGroupID INTEGER,
	FOREIGN KEY (senderUserID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (receiverUserID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (receiverGroupID) REFERENCES Groups(id) ON DELETE CASCADE,
	CHECK(
        type IN("audio", "video", "text", "file", "photo", "image")
    )
);

-- opened chats
CREATE TABLE IF NOT EXISTS Chats (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    users TEXT NOT NULL,
    closed TEXT,
    senderUserID INTEGER,
    receiverUserID INTEGER,
    receiverGroupID INTEGER,
    FOREIGN KEY (senderUserID) REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY (receiverUserID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (receiverGroupID) REFERENCES Groups(id) ON DELETE CASCADE
    CHECK(
        type IN("user", "group")
    )
);

CREATE TABLE IF NOT EXISTS Events (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    title TEXT NOT NULL,
    about TEXT NOT NULL,
    datetime INTEGER NOT NULL,
    userID INTEGER,
    groupID INTEGER,
    FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (groupID) REFERENCES Groups(id) ON DELETE CASCADE,
    CHECK(
        type = "event" AND
        LENGTH(about) <= 400
    )
);

CREATE TABLE IF NOT EXISTS EventAnswers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    answer INTEGER,
    userID INTEGER,
    eventID INTEGER,
    FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (eventID) REFERENCES Events(id) ON DELETE CASCADE,
    CHECK(
        answer IN(0,1,2)
    )
);

CREATE TABLE IF NOT EXISTS Relations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    value INTEGER, -- if send req = -1; if follow one = 1(sender follow receiver); if both follow = 0.
    senderUserID INTEGER,
    senderGroupID INTEGER,
    receiverUserID INTEGER,
    receiverGroupID INTEGER,
    FOREIGN KEY (senderUserID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (senderGroupID) REFERENCES Groups(id) ON DELETE CASCADE,
	FOREIGN KEY (receiverUserID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (receiverGroupID) REFERENCES Groups(id) ON DELETE CASCADE,
    CHECK(
        value IN(0, 1, -1) 
    )
);

-- photo & video
CREATE TABLE IF NOT EXISTS Media (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    type TEXT,
    datetime INTEGER NOT NULL,
    src TEXT NOT NULL,
    preview TEXT, -- for video poster
    userID INTEGER, -- if media uploaded as gallery to user
    groupID INTEGER, -- if media uploaded as gallery to group
    FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (groupID) REFERENCES Groups(id) ON DELETE CASCADE,
    CHECK(
        type IN("video", "photo")
    )
);

CREATE TABLE IF NOT EXISTS Comments (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	body TEXT NOT NULL,
    datetime INTEGER NOT NULL,
    isHaveChild INTEGER, -- is have answers
    isAnswer INTEGER, -- is it comment answer
    isHaveClippedFiles INTEGER, -- is have clipped files
    userID INTEGER, -- who write comment
    postID INTEGER, -- if comment to post, id > 0
    commentID INTEGER, -- if comment to comment(answer), id > 0
    mediaID INTEGER,
	FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (postID) REFERENCES Posts(id) ON DELETE CASCADE,
	FOREIGN KEY (commentID) REFERENCES Comments(id) ON DELETE CASCADE,
	FOREIGN KEY (mediaID) REFERENCES Media(id) ON DELETE CASCADE,
	CHECK(
        isHaveChild IN (0, 1) AND
        isAnswer IN (0, 1) AND
        isHaveClippedFiles IN (0, 1)
    )
);

/**
 *  (type = 1)  Create post
 *  (type = 2)  Invite to event
 *  (type = 3)  Create group
 *  (type = 4)  Invite to be member in the group
 *  (type = 5)  Request to follow to user
 *  (type = 6)  Request to follow to group
 *  (type = 10) Liked post
 *  (type = 11) Liked comment
 *  (type = 12) Liked photo
 *  (type = 13) Liked video
 *  (type = 20) Comment post
 *  (type = 21) Comment comment
 *  (type = 22) Comment photo
 *  (type = 23) Comment video
 */

CREATE TABLE IF NOT EXISTS Notifications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    datetime INTEGER NOT NULL,
    type INTEGER,
    senderUserID INTEGER,
    receiverUserID INTEGER, -- if to all followers = null
    relationID INTEGER,
    postID INTEGER,
    commentID INTEGER,
    eventID INTEGER,
    groupID INTEGER,
    mediaID INTEGER,
    FOREIGN KEY (senderUserID) REFERENCES Users(id) ON DELETE CASCADE,
	FOREIGN KEY (receiverUserID) REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY (relationID) REFERENCES Relations(id) ON DELETE CASCADE,
    FOREIGN KEY (postID) REFERENCES Posts(id) ON DELETE CASCADE,
	FOREIGN KEY (commentID) REFERENCES Comments(id) ON DELETE CASCADE,
    FOREIGN KEY (eventID) REFERENCES Events(id) ON DELETE CASCADE,
    FOREIGN KEY (groupID) REFERENCES Groups(id) ON DELETE CASCADE,
	FOREIGN KEY (mediaID) REFERENCES Media(id) ON DELETE CASCADE,
    CHECK(
        type IN(1, 2, 3, 4, 5, 6, 10, 11, 12, 13, 20, 21, 22, 23)
    )
);

CREATE TABLE IF NOT EXISTS Likes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    userID INTEGER,
    postID INTEGER,
    commentID INTEGER,
    mediaID INTEGER,
    FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY (postID) REFERENCES Posts(id) ON DELETE CASCADE,
	FOREIGN KEY (commentID) REFERENCES Comments(id) ON DELETE CASCADE,
	FOREIGN KEY (mediaID) REFERENCES Media(id) ON DELETE CASCADE
);

-- clipped files
CREATE TABLE IF NOT EXISTS Files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT,
    src TEXT NOT NULL,
    name TEXT NOT NULL,
    userID INTEGER,
    postID INTEGER, -- if clipped to post, id > 0
    commentID INTEGER, -- if clipped to comment, id > 0
    messageID INTEGER, -- if clipped to message, id > 0
    FOREIGN KEY (userID) REFERENCES Users(id) ON DELETE CASCADE,
    FOREIGN KEY (postID) REFERENCES Posts(id) ON DELETE CASCADE,
	FOREIGN KEY (commentID) REFERENCES Comments(id) ON DELETE CASCADE,
	FOREIGN KEY (messageID) REFERENCES Messages(id) ON DELETE CASCADE,
    CHECK(
        type IN("audio", "video", "image", "file")
    )
);