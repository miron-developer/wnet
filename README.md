# Real-time project.
## AniFor - anime forum

---

### AniFor site
[site here](https://anifor.herokuapp.com/)

### Basic
This project about creating simple forum(single page application also `SPA`) with chat(`websocket` power).
Forum topic: east side culture as anime, manga, ranobe

### Functionality
AniFor's functionality:
- Sign-in\up with smtp
- Restore password with smtp
- Change user avatar
- Change user data
- Create post with own engine(post engine also `PE`)
- See posts
- Create comment
- See comments
- Set like or dislike
- See post/comment carma
- Chat between 2 user
- See user profile
- Set user carma
- Send image in chat
- Typing progress in chat
- Searching `(in next versions)`
- Filtering `(in next versions)`

### Project structure
root:
* `app`: application backend layer
  * `dbfuncs`: layer between database and application
    * `converter.go` - for convert between interface and other types
    * `functions.go` - general sql functions
    * `init.go` - initialize database
    * `models.go` - db table's models
    * `mutation.go` - somewith as `graphql`'s mutation
    * `query.go` - something as `graphql`'s query
  * `api.go` - funcs for api
  * `funstions.go` - general funcs for whole app
  * `handlers.go` - all handlers
  * `init.go` - app initalizer
  * `sessions.go` - funcs for sessions as `sessionStart` and `updateCooks`
  * `sign.go` - funcs for user signification
  * `upload.go` - file uploader
  * `ws.go` - websocket's funcs
* `db` - our db
* `static` - static files or assets:
  * `css` - css files:
  * `img` - images
    * `app` - images for application
    * `avatar` - users avatars save here
    * `post` - posts images
  * `js` - js files
    * `logical` - js files witch handle logical side
      * `post.js` - handle create comment, show comments, like and dislike
      * `sign.js` - handle user logout, sign-in\up
      * `user.js` - handle spa change on user online and offline
      * `ws.js` - websocket on client
    * `postEngine` - create post engine
      * `cursor.js` - change cursor between elements
      * `editor.js` - change element to `edit` state
      * `engine.js` - main file of engine, start engine
      * `generators.js` - create new elements
      * `trasher.js` - remove current element
    * `spa` - spa funcs
      * `api.js` - getter data from back
      * `content.js` - fill with content
      * `fetch.js` - fetcher
      * `inform.js` - inform user wher some action
      * `pages.js` - all pages
      * `router.js` - handle routing
    * `script.js` - app's main script
  * `index.html` - main page body
* `tls` - selfgenerated ssl sertificate
* `CNAME` - site link
* Dockerfile and so on.
