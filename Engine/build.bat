@Echo off
go build -ldflags "-X main.Version=1.0.0 -X main.BuildTime=%date%-%time%"
