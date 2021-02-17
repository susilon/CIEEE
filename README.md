# CIEEE
Example Implementation of E2EE Chat Application using Golang, Gorilla Websocket, CrypticoJS

To demonstrate how End To End Encryption chatting apps works.

Edit port number at main.go files, default is 8380<br />
to start server : go run .<br />
open serveraddress:8380 at browser,<br />
to make second client, open at another private or incognito mode browser,<br />
or using another computer.

You can see console log at the browser for activity<br />

Dockerfile available<br />
build image : docker build -t cieee-server .<br />
run container : docker run -d -p 8380:8380 --name cieee-server cieee-server<br />

[Live Demo](https://cieee.azure.susilon.com)

Requirements :<br />
Golang

Credits :<br />
Cryptico<br />
Golang Gorilla Websocket<br />
Bootstrap<br />
JQuery<br />
Font Awesome<br />
SweetAlert 2<br />

Have Fun!
