module github.com/portwhine/portwhine/workers/webanalyzer

go 1.25.0

require (
	connectrpc.com/connect v1.19.1
	github.com/google/uuid v1.6.0
	github.com/portwhine/portwhine v0.0.0
	github.com/projectdiscovery/wappalyzergo v0.2.73
	google.golang.org/protobuf v1.36.11
)

require (
	golang.org/x/net v0.52.0 // indirect
	golang.org/x/text v0.35.0 // indirect
)

replace github.com/portwhine/portwhine => ../..
