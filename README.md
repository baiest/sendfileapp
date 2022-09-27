# sendfileapp

This app is a server tcp to send files, server has 3 channels to send or receive data

- To run server ```go run server/server.go```
- To run client mod receive ```go run client/client.go --channel 1 --action receive```
- To run client mod send ```go run client/client.go --channel 1 --action receive --file hello.txt```
