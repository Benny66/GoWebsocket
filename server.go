// Copyright 2015 The Gorilla WebSocket Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build ignore

package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"html/template"
	"log"
	"net/http"
)

var addr = flag.String("addr", "127.0.0.1:8080", "http service address")

var upgrader = websocket.Upgrader{} // use default options

func echo(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	fmt.Println(r.URL)
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		fmt.Println(mt, message)
		var data map[string]interface{}
		if err = json.Unmarshal(message, &data); err != nil {
			fmt.Println("json err", err)
			return
		}
		if data["type"] == "heartbeat" {
			data["data"] = "ok";
			jsonStr, jsonStrErr := json.Marshal(data)
			if jsonStrErr != nil {
				log.Println("json write:", jsonStrErr)
				return
			}
			err = c.WriteMessage(mt, []byte(jsonStr))
			if err != nil {
				log.Println("write:", err)
			}
		}else{
			data["type"] = "monitorAlarm";
			data["data"] = 1212;
			jsonStr, jsonStrErr := json.Marshal(data)
			if jsonStrErr != nil {
				log.Println("json write:", jsonStrErr)
				return
			}
			err = c.WriteMessage(mt, []byte(jsonStr))
			if err != nil {
				log.Println("write:", err)
			}
			//go func() {
			//	for i := 0; i < 10; i++ {
			//		time.Sleep(1 * time.Second)
			//		err = c.WriteMessage(mt, []byte("{data:subscribe}"))
			//		if err != nil {
			//			log.Println("write:", err)
			//			break
			//		}
			//	}
			//}()
		}

	}
}

func home(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

func main() {
	flag.Parse()
	log.SetFlags(0)
	http.HandleFunc("/echo", echo)
	http.HandleFunc("/", home)

	log.Fatal(http.ListenAndServe(*addr, nil))


}

var homeTemplate = template.Must(template.New("").Parse(`
<!DOCTYPE html>
<html>
<head>
<meta charset="utf-8">
<script>  
window.addEventListener("load", function(evt) {
    var output = document.getElementById("output");
    var input = document.getElementById("input");
    var ws;
    var print = function(message) {
        var d = document.createElement("div");
        d.textContent = message;
        output.appendChild(d);
    };
    document.getElementById("open").onclick = function(evt) {
        if (ws) {
            return false;
        }
        ws = new WebSocket("{{.}}");
        ws.onopen = function(evt) {
            print("OPEN");
        }
        ws.onclose = function(evt) {
            print("CLOSE");
            ws = null;
        }
        ws.onmessage = function(evt) {
            print("RESPONSE: " + evt.data);
        }
        ws.onerror = function(evt) {
            print("ERROR: " + evt.data);
        }
        return false;
    };
    document.getElementById("send").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        print("SEND: " + input.value);
		var messageJson = '{"type":' + input.value + '}'
        ws.send(messageJson);
        return false;
    };
    document.getElementById("close").onclick = function(evt) {
        if (!ws) {
            return false;
        }
        ws.close();
        return false;
    };
});
</script>
</head>
<body>
<table>
<tr><td valign="top" width="50%">
<p>Click "Open" to create a connection to the server, 
"Send" to send a message to the server and "Close" to close the connection. 
You can change the message and send multiple times.
<p>
<form>
<button id="open">Open</button>
<button id="close">Close</button>
<p><input id="input" type="text" value="Hello world!">
<button id="send">Send</button>
</form>
</td><td valign="top" width="50%">
<div id="output"></div>
</td></tr></table>
</body>
</html>
`))