go-quip
=======

Quip bindings for Go

# Installation

`go get` should work:

```
go get github.com/mduvall/go-quip
```

# API Support

The library currently supports the [documented endpoints](https://quip.com/api/reference).

# Usage/Examples

For usage of `go-quip` and examples using the library, refer to the [godoc](http://godoc.org/github.com/mduvall/go-quip) page.

A quick rundown:

```go
q := quip.NewClient("access_token")

// Getting recent messages...
rmp := quip.GetRecentMessagesParams{
    ThreadId: "thread_id",
}
messages := q.GetRecentMessages(&rmp)

for _, message := range messages {
    fmt.Println(message.Text)
}

// Creating a new message

nm := quip.NewMessageParams{
    ThreadId: "YLZAAAQV9Py",
    Content:  "Byah!",
}
message := q.NewMessage(&nm)
```
