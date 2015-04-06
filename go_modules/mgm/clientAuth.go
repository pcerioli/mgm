package mgm

import (
  "github.com/gorilla/websocket"
  //"github.com/gorilla/sessions"
  "fmt"
  "net/http"
  "encoding/json"
  "../../go_modules/simian"
  "github.com/satori/go.uuid"
)

var upgrader = &websocket.Upgrader{ReadBufferSize: 1024, WriteBufferSize: 1024}

func (cm ClientManager) WebsocketHandler(w http.ResponseWriter, r *http.Request) {
  fmt.Println("New connection on ws")
  ws, err := upgrader.Upgrade(w, r, nil)
  if err != nil {
    fmt.Println(err)
    return
  }
  c := cm.NewClient(ws)
  c.process()
}

func (cm ClientManager) LogoutHandler(w http.ResponseWriter, r *http.Request) {
  session, _ := cm.store.Get(r, "MGM")
  delete(session.Values, "guid")
  delete(session.Values,"address")
  session.Save(r,w)
  w.Header().Set("Content-Type", "application/json")
  w.Write([]byte("{\"Success\": true}"))
}

func (cm ClientManager) ResumeHandler(w http.ResponseWriter, r *http.Request) {
  session, _ := cm.store.Get(r, "MGM")
  
  fmt.Println("Session resume attempt");
  fmt.Println(session.Values);
  fmt.Println(len(session.Values));
  
  type clientAuthResponse struct {
    Uuid string
    Success bool
  }
  
  if len(session.Values) == 0 {
    response := clientAuthResponse{"", false}
    js, err := json.Marshal(response)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/jons")
    w.Write(js)
    return
  }
  
  response := clientAuthResponse{session.Values["guid"].(string), true}
  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  w.Header().Set("Content-Type", "application/json")
  w.Write(js)

}

func (cm ClientManager) RegisterHandler(w http.ResponseWriter, r *http.Request) {
  
}

func (cm ClientManager) PasswordResetHandler(w http.ResponseWriter, r *http.Request) {
  
}

func (cm ClientManager) PasswordTokenHandler(w http.ResponseWriter, r *http.Request) {
  
}

func (cm ClientManager) LoginHandler(w http.ResponseWriter, r *http.Request) {
  decoder := json.NewDecoder(r.Body)
  
  type clientAuthRequest struct {
    Username string
    Password string
  }
  
  var t clientAuthRequest   
  err := decoder.Decode(&t)
  if err != nil {
    fmt.Println("Invalid auth request")
    return
  }
  
  type clientAuthResponse struct {
    Uuid uuid.UUID
    Message string
    Success bool
  }
  
  sim, _ := simian.Instance()
  guid,err := sim.Auth(t.Username,t.Password);
  if err != nil {
    response := clientAuthResponse{uuid.UUID{}, err.Error(), false}
    js, err := json.Marshal(response)
    if err != nil {
      http.Error(w, err.Error(), http.StatusInternalServerError)
      return
    }
    w.Header().Set("Content-Type", "application/json")
    w.Write(js)
    return
  }
  
  response := clientAuthResponse{guid, "", true}
  js, err := json.Marshal(response)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  session, _ := cm.store.Get(r, "MGM")
  session.Values["guid"] = guid.String()
  session.Values["address"] = r.RemoteAddr
  err = session.Save(r,w)
  if err != nil {
    fmt.Println(err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
    return
  }
  
  fmt.Println("Session saved, returning success")
  
  w.Header().Set("Content-Type", "application/jons")
  w.Write(js)
}