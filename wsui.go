package wsui

import (
	"net/http"
	"github.com/gorilla/websocket"
	"encoding/json"
	"reflect"
	"fmt"
	// "log"
)

type UI interface{
	CSS(href string)
	Script(src string)
	View(body string)
	ClearPage()
	Bind(name string, f interface{})
	ClearBind()
	Exec(js string)
	Close()
}

type app struct{
	upgrader websocket.Upgrader
	onCreate func(UI)
	ws_pattern string
	*http.Server
	html string
}

type ui struct{
	conn *websocket.Conn
	func_ map[string] interface{}
	quit bool
}

func (t *app)ws_handle(w http.ResponseWriter, r *http.Request){
	if t.onCreate == nil{
		return
	}
	c, err := t.upgrader.Upgrade(w, r, nil)
	defer func(){
		if err := recover();err != nil {
			fmt.Printf("%s\n", err)
		}
		c.Close()
	}()
	if err != nil {
		panic("wsui: ws_handle 1")
	}
	ui_ := new_ui(c)
	t.onCreate(ui_)
	ui_.loop()
}

const WSJS string = `
window.onload = function(){
	_app_ = function(url){
		ws = new WebSocket(url);
		ws.onopen = function(evt) {
			console.log("OPEN");
		}
		ws.onclose = function(evt) {
			console.log("CLOSE");
			ws = null;
		}
		ws.onmessage = function(evt) {
			try{
				eval(evt.data)
			}
			catch(e){
				console.log(e)
			}
		}
		ws.onerror = function(evt) {
			console.log("ERROR: " + evt.data);
		}
		return ws
	}("%s")
}
window.onunload = function(){
	if (_app_.close){
		_app_.close()
	}
	_app_ = null
}`

func (t* app)handle(w http.ResponseWriter, r *http.Request){
	w.Write([]byte(fmt.Sprintf(t.html, "ws://"+r.Host+t.ws_pattern)))
}

func new_ui(c *websocket.Conn) *ui {
	ui_ := &ui{
		conn: c,
		func_: make(map[string] interface{}),
		quit: false,
	}
	ui_.Exec(`
_app_ = function(ws){
  var app = {}
  var app_ws = ws
  app.new_bind = function(){
    var bind = {}
    var app_return_map_ = {}
    var app_id_cnt_ = 0
    bind.func = function(call){
      return function(){
        var args = Array.prototype.slice.call(arguments);
        a_id = "id" + app_id_cnt_
        app_id_cnt_ = app_id_cnt_ + 1
        a_promise = new Promise(function(resolve, reject){
          app_return_map_[a_id] = function(res){
            resolve(res)
          }
        })
        app_ws.send(JSON.stringify({
          "call":call,
          "id"  :a_id,
          "args":args,
        }))
        return a_promise
      }
    }
    bind.return_ = function(id, res){
      app_return_map_[id](res)
      app_return_map_[id] = null
      delete app_return_map_[id]
    }
    bind.close = function(){
      app_return_map_ = null
      app_id_cnt_ = null
      bind = null
    }
    return bind
  }
  app.new_loader = function(){
    var loader = {}
    var loader_head = document.getElementsByTagName("head")[0]
    var loader_elems = []
    var loader_loaded = 0
    var loader_onload = null
    var onload_ = function(){
      loader_loaded = loader_loaded + 1
      if (loader_loaded == loader_elems.length){
        if(loader_onload){
          loader_onload()
        }
      }
    }
    loader.script = function(src){
      var elem = document.createElement("script")
      elem.src = src
      elem.onload = onload_
      loader_elems.push(elem)
    }
    loader.css = function(href){
      var elem = document.createElement("link")
      elem.rel="stylesheet"
      elem.href = href
      elem.onload = onload_
      loader_elems.push(elem)
    }
    var loader_run = function(){
      for(var i in loader_elems){
        loader_head.appendChild(loader_elems[i])
      }
    }
    loader.view = function(html){
      if (loader_elems.length == loader_loaded){
        document.body.innerHTML = html
        return
      }
      loader_onload = function(){
        document.body.innerHTML = html
      }
      loader_run()
    }
    loader.clear = function(){
      for(var i in loader_elems){
        loader_head.removeChild(loader_elems[i])
      }
      document.body.innerHTML = ""
      loader_elems = []
      loader_loaded = 0
      loader_onload = null
    }
    loader.close = function(){
      loader_head = null
      loader_elems = null
      loader_loaded = null
      loader_onload = null
      onload_ = null
      loader_run = null
      loader = null
    }
    return loader
  }
  app.bind = app.new_bind()
  app.loader = app.new_loader()
  app.close = function(){
    app.bind.close()
    app.loader.close()
    app_ws = null
    app = null
  }
  return app
}(_app_)`)
	return ui_
}

func (t *ui)return_(jmsg map[string] interface{}, res interface{}){
	res_, err := json.Marshal(res)
	if err != nil {
		panic("wsui: return_")
	}
	t.Exec(fmt.Sprintf(`
	_app_.bind.return_("%s", JSON.parse(%s%s%s))
	`, jmsg["id"].(string), "`", res_, "`"))
}

func (t *ui)loop(){
	for !(t.quit){
		mt, message, err := t.conn.ReadMessage()
		if err != nil {
			panic("wsui: loop 2")
		}
		go func(){

		defer func(){
			if err := recover();err != nil {
				fmt.Printf("%s\n", err)
			}
		}()

		if mt != websocket.TextMessage {
			panic("wsui: loop 3")
		}
		// fmt.Printf("recv: %s", message)
		var jmsg map[string]interface{}
		err = json.Unmarshal(message, &jmsg)
		if err != nil {
			panic("wsui: loop 4")
		}
		call, ok := jmsg["call"].(string)
		if !ok {
			panic("wsui: loop 5")
		}
		f, ok := t.func_[call]
		if !ok {
			panic("wsui: loop 6")
		}
		v := reflect.ValueOf(f)
		if v.Kind() != reflect.Func {
			panic("wsui: loop 7")
		}
		if n := v.Type().NumOut(); n > 1 {
			panic("wsui: loop 8")
		}
		raw, ok := jmsg["args"].([]interface{})
		if !ok {
			panic("wsui: loop 9")
		}
		if len(raw) != v.Type().NumIn() {
			panic("wsui: loop 10")
		}
		args := []reflect.Value{}
		for i := range raw {
			arg := reflect.New(v.Type().In(i))
			arg_, err := json.Marshal(raw[i])
			if err != nil {
				panic("wsui: loop 11")
			}
			err = json.Unmarshal(arg_, arg.Interface())
			if err != nil {
				panic("wsui: loop 11 ?")
			}
			args = append(args, arg.Elem())
		}
		res := v.Call(args)
		switch len(res) {
		case 0:
			t.return_(jmsg, nil)
		case 1:
			t.return_(jmsg, res[0].Interface())
		default:
			panic("wsui: loop 12")
		}

		}()
	}
}

func (t *ui)CSS(href string){
	t.Exec(fmt.Sprintf(`
	_app_.loader.css("%s")
	`, href))
}

func (t *ui)Script(src string){
	t.Exec(fmt.Sprintf(`
	_app_.loader.script("%s")
	`, src))
}

func (t *ui)View(body string){
	t.Exec(fmt.Sprintf(`
	_app_.loader.view(%s %s %s)
	`, "`", body, "`"))
}

func (t *ui)ClearPage(){
	t.Exec(`
	_app_.loader.clear()
	`)
}

func (t *ui)Bind(name string, f interface{}){
	t.func_[name] = f
	t.Exec(fmt.Sprintf(`
	%s = _app_.bind.func("%s")
	`, name, name))
}

func (t *ui)ClearBind(){
	for i := range t.func_ {
		delete(t.func_, i)
		t.Exec(fmt.Sprintf(`
		%s = null
		`, i))
	}
}

func (t *ui)Exec(js string){
	defer func(){
		if err := recover();err != nil {
			fmt.Printf("%s\n", err)
			t.conn.Close()
		}
	}()
	err := t.conn.WriteMessage(websocket.TextMessage, []byte(js))
	// log.Printf("exec %s", js)
	if err != nil {
		panic("wsui: exec")
	}
}

func (t *ui)Close(){
	t.quit = true
	t.conn.Close()
}

func NewApp(addr, pattern, ws_pattern_ string, onCreate_ func(UI)) *app{
	app_ := &app{
		ws_pattern: ws_pattern_,
		onCreate: onCreate_,
	}
	mux := http.NewServeMux()
	mux.HandleFunc(pattern, app_.handle)
	mux.HandleFunc(ws_pattern_, app_.ws_handle)
	app_.Server = &http.Server{
		Addr: addr,
		Handler: mux,
	}
	app_.html = fmt.Sprintf(`
	<!DOCTYPE html>
	<html>
	<head>
	<meta charset="utf-8">
	<link id="icon" rel="icon" href="" type="image/x-icon">
	<title id="title">#</title>
	<script>
	%s
	</script>
	</head>
	<body>
	</body>
	</html>
	`, WSJS)
	return app_
}

func (t *app)Html(html string){
	t.html = html
}