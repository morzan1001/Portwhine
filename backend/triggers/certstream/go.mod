module github.com/portwhine/portwhine/triggers/certstream

go 1.25.0

require (
	connectrpc.com/connect v1.19.1
	github.com/CaliDog/certstream-go v0.0.0-20200713031452-eca7997412f1
	github.com/google/uuid v1.6.0
	github.com/portwhine/portwhine v0.0.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	github.com/jmoiron/jsonq v0.0.0-20150511023944-e874b168d07e // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.51.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace github.com/portwhine/portwhine => ../..
