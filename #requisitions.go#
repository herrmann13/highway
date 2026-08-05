package main

import(
	"net/http"
	"fmt"
	"bytes"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

func main(){

     a := app.NewWithID("com.herrmann.requisitions") //instancia uma nova aplicação com um id

     /*Cria uma nova janela a partir da aplicação instanciada*/
     
     w := a.NewWindow("Requisição HTTP") 
     
     url := widget.NewEntry()
     url.SetPlaceHolder("https://localhost:3000")
     
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
     authorizationContent.SetPlaceHolder("Auth Content")
     
     authorization := widget.NewSelect(
     	    []string{"No Auth","Bearer Token", "OAuth 2.0", "Basic Auth"},
	    func(value string) {
	    	       fmt.Println("Autorizacao selecionada:", value)
		       if value == "Bearer Token" || value == "Basic Auth"{
		       	  authorizationContent.Enable()
			  
			  if value =="Basic Auth"{
			     authorizationContent.SetPlaceHolder("{user:password}")
			  }
			  
		       } else {
		          authorizationContent.Disable()
		       }
	    },
     )
     authorization.SetSelected("No Auth")

     authorizationContainer := container.NewBorder(
	nil,
	nil,
	authorization,
	nil,
	authorizationContent,
	)

     headerItems := []string{}
     headerEntry := widget.NewEntry()
     headerEntry.SetPlaceHolder("Content-Type: application/json")
     headerList := container.NewVBox()
	   
     addHeaderButton := widget.NewButton("Adicionar Header",
     		     			 func() {
					       header := headerEntry.Text
					       var headerItem *fyne.Container
					       
					       headerItem = container.NewHBox(
     						     widget.NewLabel(header),
						     widget.NewButton("Remover", func(){
						     		     headerList.Remove(headerItem)

								     for i, h := range headerItems {
								     	 if h==header{
									 headerItems = append(headerItems[:i], headerItems[i+1:]...)
									 break
									 }
								     }
								     
						                     headerList.Refresh()
						     }))

					 	headerList.Add(
							headerItem,
							)
						headerItems = append(headerItems, header)
						headerList.Refresh()
						headerEntry.SetText("")
						})
     headersContainer := container.NewVBox(
      		      headerList,
		      headerEntry,
		      addHeaderButton,
     		      )
		     

     requisitionLayout := container.NewVBox(
     		       method,
		       headersContainer,
		       url,
		       body,
		       authorizationContainer,)
		       

     sendButton := widget.NewButton("Enviar",
				     func() {sendRequest(method.Selected,
						     	 url.Text,
						     	 body.Text,
					                 )})
				
     top := fyne.CanvasObject(nil)
     bottom := sendButton
     left := fyne.CanvasObject(nil)
     right := fyne.CanvasObject(nil)
     center := requisitionLayout
     
     w.SetContent(container.NewBorder(
	top,
	bottom,
	left,
	right,
	center,
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
