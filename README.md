# real-time-chat-api
Golang real time chat api and it's web chat interface

Before start this go application, please make sure this requirement installed
1. Golang - https://golang.org/doc/install
2. Download required go packages by running these command in root of project folder
```
go get github.com/gorilla/mux
go get github.com/gorilla/websocket
```
3. Run the application with this command in root of project folder
```
go run main.go
```
The application will run at localhost at port 8000, so make sure the port 8000 is not used

You can access this application using two ways, with browser and api.
I created the web page to show the send message api function is working in real time.
You also can join the chat with your own username and send message in the web and to see get all sent message api is working.

To use the api:
1. Get all sent message
```
curl --location --request GET 'localhost:8000/api/messages' \
```
2. Send message
```
curl --location --request POST 'http://localhost:8000/api/messages' \
--header 'Content-Type: application/json' \
--data-raw '{
	"message": "ohok molohok"
}'
```