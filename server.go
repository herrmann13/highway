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
	fmt.Fprintln(w, "Server Online")


     })

     http.HandleFunc("/users", func(w http.ResponseWriter,r *http.Request){
     			     
        switch r.Method {

	case "GET":

	     fmt.Fprintln(w, "Método GET em /users")

	case "POST":

	     fmt.Fprintln(w, "Método POST em /users")

	case "PUT":

             fmt.Fprintln(w, "Método PUT  em /users")

	case "DELETE":

	     fmt.Fprintln(w, "Método DELETE em /users")

	}

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
