# webserver-pwdhash
example HTTP webserver 

environment variables:

**HTTP_PORT**: port the web server listens on, default: 8080

**NUM_WORKERS**: number of workers hashing password, default: 100

**BUFFER_SIZE**: job channel buffer size, default: 1000 

**MAX_PWD_LENGTH**: max password length, default: 64


use the following command to run the webserver with the default values:
go run webserver.go

unit test:
go test ./...

