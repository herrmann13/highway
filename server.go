package main

import (
       "fmt"
       "net/http"
       )

func main(){

     http.HandleFunc("/", func(w http.ResponseWriter,r *http.Request){

     	fmt.Fprint(w, httpDetail(r))

	})

     http.ListenAndServe(":8080", nil)
     }


/*
Retorna os detalhes da chamada HTTP em formato de string
*/
func httpDetail(request *http.Request) []string {
     details := []string{
		"- Method: "+fmt.Sprint(request.Method)+"\n",
		"- URL: "+fmt.Sprint(request.URL)+"\n",
		"- Headers: "+fmt.Sprint(request.Header)+"\n",
		"- Body: "+fmt.Sprint(request.Body)+"\n",
		"- Host: "+fmt.Sprint(request.Host)+"\n",
		"- RemoteAddr: "+fmt.Sprint(request.RemoteAddr)+"\n",
		"- Form: "+fmt.Sprint(request.Form)+"\n",
		"- Context: "+fmt.Sprint(request.Context())+"\n",
		}
     return details
}
