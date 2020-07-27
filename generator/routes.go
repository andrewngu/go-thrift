package endpoints

import (
  "net/http"

  "github.com/gorilla/mux"
)

type RPCClient interface {
	Call(method string, request interface{}, response interface{}) error
}

type thriftService func(RPCClient, []byte, string) (interface{}, error)

func handleError(w http.ResponseWriter, err error) {
  w.WriteHeader(http.StatusInternalServerError)
  w.Write([]byte(fmt.Printf("oops - %s", err)))
}

func JSONHandler(client RPCClient, service thriftService) http.HandlerFunc {
  return func(r http.Request, w http.ResponseWriter) {
    vars := mux.Vars(r)
    method, ok := vars["method"]

    defer r.Body.Close()å
    requestBytes, err := ioutil.ReadAll(r.Body)
    if err != nil {
      handleError(w, err)
      return
    }

    resp, err := service(client, requestBytes, method)
    if err != nil {
      handleError(w, err)
      return
    }
  }
}

func HandleServices(r *mux.Router, c RPCClient) {
  r.HandleFunc("/stsssomgr/{method}", JSONHandler(c, StSSOMgrService))
}
