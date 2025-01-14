## sessions-go

An in-memory concurrent-safe package for go web session management. 

This session management library primarily utilizes an in-memory store to ensure optimal performance and rapid access. While a Redis store is also supported as an alternative, the implementation of mutex locks for each session operation can lead to unnecessary resource consumption when using Redis. 

```shell
$ go get -u github.com/jiyi27/sessions
```

## usage

```go

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

// store instance globally
var store *sessions.MemoryStore

func main() {
	// You need specify the id length of a session, don't make it too big.
	store := NewStore()
	http.HandleFunc("/", handler)
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