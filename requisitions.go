package main

import(
	"net/http"
	"fmt"
	"bytes"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main(){

     a := app.NewWithID("com.herrmann.requisitions") //instancia uma nova aplicação com um id

     /*Cria uma nova janela a partir da aplicação instanciada*/
     w := a.NewWindow("Requisição HTTP") 
     
     url := widget.NewEntry()
     

     body := widget.NewMultiLineEntry()
     body.SetText("{}")
     
     
     method := widget.NewSelect(
     	    []string{"GET","POST","PUT","PATCH","DELETE"},
	    func(value string) {
	    	       fmt.Println("Método selecionado:", value)
		       if value == "GET" || value == "DELETE" {
		       	  body.Disable()
		       } else {
		       	  body.Enable()
		       }	       
	    },
     )
     method.SetSelected("GET")
		
     authorizationContent := widget.NewMultiLineEntry()
     authorizationContent.SetMinRowsVisible(1)
     
     authorization := widget.NewSelect(
     	    []string{"No Auth","Bearer Token", "OAuth 2.0", "Basic Auth"},
	    func(value string) {
	    	       fmt.Println("Autorizacao selecionada:", value)
		       if value == "Bearer Token" || value == "Basic Auth"{
		       	  authorizationContent.Enable()
		       } else {
		          authorizationContent.Disable()
		       }
	    },
     )
     authorization.SetSelected("No Auth")


     authorizationContainer := container.NewHBox(
     			    authorization,
			    authorizationContent)



     form := widget.NewForm(
     	  methodItem,
	  urlItem,
	  authorizationItem,
	  bodyItem,
     )

     sendButton := widget.NewButton("Enviar", func() {sendRequest(method.Selected, url.Text, body.Text)} )
     
     w.SetContent(container.NewVBox(
	form,
	sendButton,
	))

     w.ShowAndRun()
     
}

func sendRequest(method string, url string, body string){

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

     client := &http.Client{}

     resp, err := client.Do(req)

     if err != nil {
     	panic(err)	
     }

     fmt.Println(resp)

     defer resp.Body.Close()
     
}
