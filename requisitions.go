package main

import(
	"net/http"
	"fmt"
	// "io"
	"bytes"
)

func main(){

     client := &http.Client{}

     var method string
     var url string
     var body string

     fmt.Print("Method:")
     fmt.Scan(&method)

     fmt.Print("URL:")
     fmt.Scan(&url)

     fmt.Print("body:")
     fmt.Scan(&body)

     req, err := http.NewRequest(
     	  method,
     	  url,
	  bytes.NewBuffer([]byte(body)),
	  )
	  
     if err != nil{
     	panic(err)
     }
     
     req.Header.Set("Content-Type", "application/json")
     req.Header.Set("Authorization", "Bearer token")

     resp, err := client.Do(req)

     if err != nil {
     	panic(err)	
     }

     fmt.Println()

     defer resp.Body.Close()
}
