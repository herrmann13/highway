package main

import (
       "fmt"
       "net/http"
       "encoding/json"
       )

func main(){
     server(":8080")
}


/*
Inicia um servidor http
*/

func server(port string){

     http.HandleFunc("/", func(w http.ResponseWriter,r *http.Request){

     	
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintln(w, httpDetail(r))

	})
	

     http.ListenAndServe(port, nil)
}



/*
Retorna os detalhes da chamada HTTP em string
*/

func httpDetail(request *http.Request) string {

     details := map[string]any{
     	        "method":	request.Method,
		"url":		request.URL.String(),
		"headers":	request.Header,
		"body":		request.Body,
		"host":		request.Host,
		"remoteAddr":	request.RemoteAddr,
		"form":		request.Form,
		}
		
      formatedDetails, _ := json.MarshalIndent(details, "", "  ") 

      return string(formatedDetails)
}
