package main

import (
       "fmt"
       "net/http"
       "encoding/json"
       "strings"
       )

func main(){
     port := ":8080"
     mux := http.NewServeMux()

     mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request){

         fmt.Fprintln(w, "Servidor Online")
     
     })

     mux.HandleFunc("/users", userHandler)

     formatedPort := strings.ReplaceAll(port, ":","")
     
     fmt.Println("Rodando servidor na porta", formatedPort)
     
     http.ListenAndServe(port, mux)
     

}



func userHandler(w http.ResponseWriter, r *http.Request){

     switch r.Method {

     	    case http.MethodGet:

	    	 id := r.URL.Query().Get("userId")

		 if id != "" {
		    fmt.Fprintln(w, "Método GET em /users para o usuário %i", id)
		 }

		 fmt.Fprintln(w, "Método GET em /users")
	    	 
	    case http.MethodPost:

	    	 fmt.Fprintln(w, "Método POST em /users")

	    case http.MethodPut:

	    	 fmt.Fprintln(w, "Método PUT em /users")

	    case http.MethodDelete:

	    	 fmt.Fprintln(w, "Método DELETE em /users")

     }     
}

/*-------------------------------------------------------------------------*/



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
