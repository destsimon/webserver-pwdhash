# webserver-pwdhash
example HTTP webserver 

the webserver expects 3 environment variables to be set:
NUM_WORKERS=1000 
BUFFER_SIZE=1000 
MAX_PWD_LENGTH=64 

run the webserver with the following command: 
NUM_WORKERS=1000 BUFFER_SIZE=1000 MAX_PWD_LENGTH=64 go run webserver.go

unit test:
go test ./...

