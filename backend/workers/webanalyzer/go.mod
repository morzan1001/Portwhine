module github.com/portwhine/portwhine/workers/webanalyzer

go 1.26.0

require (
	connectrpc.com/connect v1.19.2
	github.com/google/uuid v1.6.0
	github.com/portwhine/portwhine v0.0.0
	github.com/projectdiscovery/wappalyzergo v0.2.78
	google.golang.org/protobuf v1.36.12-0.20260120151049-f2248ac996af
)

require (
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/text v0.36.0 // indirect
)

replace github.com/portwhine/portwhine => ../..
