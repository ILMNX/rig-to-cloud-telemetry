module rigtelemetry/edge

go 1.22

replace rigtelemetry/shared => ../shared

require (
	github.com/caarlos0/env/v11 v11.3.1
	github.com/eclipse/paho.mqtt.golang v1.5.0
	github.com/mattn/go-sqlite3 v1.14.24
	go.bug.st/serial v1.6.2
	rigtelemetry/shared v0.0.0
)

require (
	github.com/creack/goselect v0.1.2 // indirect
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/net v0.27.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/sys v0.22.0 // indirect
)
