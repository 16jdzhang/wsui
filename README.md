 wsui
===
use websocket to build website.<br>
wsui can help you build web application earily.

# example
```Go
// example.go
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
	go run example.go
so we can open the website by http://localhost:8080/

--------

* func (t *ui)View(body string)<br>
change the innerHTML in page body

* func (t *ui)Bind(name string, f interface{})<br>
bind a go function to a javascript function, so that the page can send argment to server, and server can return result to brower by its return value

* func (t *ui)Exec(js string)<br>
ask the browser to run the javascript code in string

-----------
my email: 16jdzhang@stu.edu.cn<br>
i won't update frequently, because i am very busy to find a job, otherwise my family will kill me.<br>
希望我能找到工作把~