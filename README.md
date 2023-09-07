## sessions-go

An in-memory concurrent-safe package for go web sessions management. 

We use a session-id to identify the session stored in the server, the session-id is transmitted as cookie between server and client.

The actuall data are stored in server. Learn more about security: [Better security - Session ID in cookies vs. Encrypted cookie](https://security.stackexchange.com/questions/174334/better-security-session-id-in-cookies-vs-encrypted-cookie)

## Feature
- Get all expired sessions, which enables do some stuff before session removed from store. 

```go
// If you want get expired in future, pass true to second argument which is optional.
// Otherwise just: store := NewMemoryStore(32)
store := NewMemoryStore(32, true)
session, _ := store.Get(r, "session-id")
session.InsertValue("name", "Coco")
session.Save(rsp)
...
// If you telled store you want get all session, 
// you should make another goroutine to keep listen the channel, 
// otherwise your expired session won't be removed from store.
go func() {
    expiredSessions := make(chan []*Session)
    errSession := make(chan error)
    go store.GetExpiredSessions(expiredSessions, errSession)
    select {
    case sessions := <-expiredSessions:
        for _, session := range sessions {
			// the code below will execute before expired session being removed.
			// do someting you want.
            fmt.Println(session.values)
        }
    case err := <-errSession:
        // error handling
    }
}()
```

## Usage

```go
func doNothing(_ http.ResponseWriter, _ *http.Request) {}

func handler(w http.ResponseWriter, r *http.Request) {
	// Get will return a cached session if exists, if not return a new one.
	// For simplicity, ignore error here
	session, _ := store.Get(r, "session-id")
	// If it's a new session, set some info into it.
	if session.IsNew() {
		session.InsertValue("name", "Coco")
		session.InsertValue("age", 18)
		// Set max age for session in seconds.
		session.SetMaxAge(30)
		// Save session into response to client.
		// You should save session after you make change on a session.
		session.Save(w)
		return
	}
	// If not a new session.
	name, _ := session.GetValueByKey("name")
	_, _ = fmt.Fprint(w, fmt.Sprintf("hello %v\n", name))
}

// You just need only one store instance on global.
var store *sessions.MemoryStore

func main() {
	// You need specify the id length of a session, don't make it too big.
	store = sessions.NewMemoryStore(32)
	http.HandleFunc("/", handler)
	http.HandleFunc("/favicon.ico", doNothing)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
```

```shell
$ curl localhost:8080/ -v 
*   Trying 127.0.0.1:8080...
...
< HTTP/1.1 200 OK
< Set-Cookie: session-id=F7j4Ftn5jgdKYyAxfVdR6lc=E=IdJKdx; Path=/; Expires=Tue, 05 Sep 2023 14:36:53 GMT; Max-Age=30
...

# Send get request with cookie
$ curl localhost:8080/ --cookie "session-id=F7j4Ftn5jgdKYyAxfVdR6lc=E=IdJKdx"
hello Coco
```