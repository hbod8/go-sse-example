package main

import (
	"encoding/json"
	"fmt"
	htemplate "html/template"
	"net/http"
	"text/template"
	"time"
)

type ResponseData struct {
	Name string
}

type EventData struct {
	Message string
}

type Event struct {
	Type string
	Data EventData
}

func eventHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\r\n", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	sink, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	fmt.Fprint(w, ": connected\n\n")
	sink.Flush()

	fm := template.FuncMap{
		"json": func(v interface{}) string {
			a, _ := json.Marshal(v)
			return string(a)
		},
	}

	t := template.Must(template.New("Event").Funcs(fm).Parse("event: {{.Type}}\r\ndata: {{json .Data}}\r\n\r\n"))

	ctx := r.Context()

	for {
		select {
		case <-ctx.Done():
			fmt.Println("Client disconnected")
			return
		case <-time.After(3 * time.Second):
			e := Event{
				Type: "chat",
				Data: EventData{
					Message: "Goober",
				},
			}

			if err := t.Execute(w, e); err != nil {
				fmt.Println("template error:", err)
				return
			}

			fmt.Println("Sent Event")

			sink.Flush()
		}
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s %s\r\n", r.Method, r.URL.Path)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Server", "Go Web Server")

	data := ResponseData{
		Name: r.Host,
	}

	t := htemplate.Must(htemplate.ParseFiles("index.gohtml"))

	t.Execute(w, data)
}

func main() {
	fmt.Println("Hello, World!")

	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/events", eventHandler)
	http.ListenAndServe(":8080", nil)
}
