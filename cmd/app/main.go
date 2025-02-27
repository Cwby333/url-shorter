package main

import (
	apprunner "github.com/Cwby333/url-shorter/internal/apprunnrer"
	_ "net/http/pprof"
)

func main() {
	app := apprunner.New()
	app.Run()
}
