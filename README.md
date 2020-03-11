 wsui
===
use websocket to build website

# example
```Go
package main

import (
	"github.com/16jdzhang/wsui"
)

func main(){
	app := wsui.NewApp("localhost:8080", "/", "/ws", page1)
	app.ListenAndServe()
}

func page1(ui wsui.UI){
	ui.View(`
	<h1>this is page1</h1>
	<br>
	<button onclick="topage2()">click me, go to page2</button>
	`)
	ui.Bind("topage2", func(){
		page2(ui)
	})
}

func page2(ui wsui.UI){
	ui.View(`
	<h1>this is page2</h1>
	<br>hello, what is your name?
	<input type="text" id="text1">
	<br>
	<button onclick="hello(text1.value).then(function(res){helloh1.innerHTML = res;})">click me, </button>
	<br>
	<h1 id="helloh1"></h1>
	`)
	ui.Bind("hello", func(name string) string{
		return "hello, "+name+" !"
	})
}
```