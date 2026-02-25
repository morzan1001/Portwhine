module github.com/portwhine/portwhine/workers/screenshot

go 1.25.0

require (
	connectrpc.com/connect v1.19.1
	github.com/chromedp/chromedp v0.13.6
	github.com/google/uuid v1.6.0
	github.com/portwhine/portwhine v0.0.0
	golang.org/x/net v0.50.0
	google.golang.org/protobuf v1.36.11
)

require (
	github.com/chromedp/cdproto v0.0.0-20250403032234-65de8f5d025b // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/go-json-experiment/json v0.0.0-20250211171154-1ae217ad3535 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	golang.org/x/sys v0.41.0 // indirect
	golang.org/x/text v0.34.0 // indirect
)

replace github.com/portwhine/portwhine => ../..
