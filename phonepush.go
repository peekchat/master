package phonepush


import (
    "html/template"
    "net/http"
    "crypto/rand"
    "fmt"
    "io"

    "appengine"
    "appengine/channel"
)

// newUUID generates a random UUID according to RFC 4122
func newUUID() (string, error) {
  uuid := make([]byte, 16)
  n, err := io.ReadFull(rand.Reader, uuid)
  if n != len(uuid) || err != nil {
    return "", err
  }
  // variant bits; see section 4.1.1
  uuid[8] = uuid[8]&^0xc0 | 0x80
  // version 4 (pseudo-random); see section 4.1.3
  uuid[6] = uuid[6]&^0xf0 | 0x40
  return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:]), nil
}

func init() {
    http.HandleFunc("/", main)
    http.HandleFunc("/send", send)
    http.HandleFunc("/_ah/channel/connected/", channelConnected)
    http.HandleFunc("/_ah/channel/disconnected/", channelDisconnected)
}

var mainTemplate = template.Must(template.ParseFiles("./webApp/www/index.html"))

func channelConnected(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    c.Infof("channelConnected: %s", r.FormValue("from"))
}

func channelDisconnected(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    c.Infof("channelDisconnected: %s", r.FormValue("from"))
}

func send(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    key := r.FormValue("key")
    message := r.FormValue("message")
    c.Infof("sending message: key=%s, '%v'", key, message)

    err := channel.SendJSON(c, key, message)
    if err != nil {
      c.Errorf("error sending message: %v", err)
    }
}

func main(w http.ResponseWriter, r *http.Request) {
    c := appengine.NewContext(r)
    key, err := newUUID()
    if err != nil {
      c.Errorf("uuid error: %v", err)
    }

    c.Infof("Creating channel with key: %v", key)
    tok, err := channel.Create(c, key)
    if err != nil {
        http.Error(w, "Couldn't create Channel", http.StatusInternalServerError)
        c.Errorf("channel.Create: %v", err)
        return
    }

    err = mainTemplate.Execute(w, map[string]string{
        "key": key,
        "token":    tok,
    })
    if err != nil {
        c.Errorf("mainTemplate: %v", err)
    }
}
