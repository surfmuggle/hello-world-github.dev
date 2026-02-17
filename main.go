package main

import (
		"fmt"
		"log"
		"net/http"
)

func main() {
		http.HandleFunc("/", helloHandler)
		addr := ":8080"
		fmt.Println("Starting server on", addr)
		log.Fatal(http.ListenAndServe(addr, nil))
}

func helloHandler(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, `<!doctype html>
<html>
	<head>
		<meta charset="utf-8">
		<title>Welcome</title>
	</head>
	<body>
		<h2>welcome to github.dev</h2>
		<p>hello word</p>
	</body>
</html>`)
}