package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"fmt"
	"io"
)

var profile []byte

func main() {
	const staticPath = "assets"
	const indexPath = "index.html"

	fileServer := http.FileServer(http.Dir(staticPath))

	log.Print("Listening on port :8080")

	http.ListenAndServe(":8080", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pprof" {
			w.Header().Add("Access-Control-Allow-Origin","*")
			if  strings.EqualFold(r.Method ,"POST"){
				var err error
				profile,err= io.ReadAll(r.Body)
				if err!= nil{
					log.Print(err)
					w.WriteHeader(http.StatusExpectationFailed)
					fmt.Fprint(w,err.Error())
					return
				}
				log.Print("ok profile has been created...")
				w.WriteHeader(http.StatusOK)
				fmt.Fprint(w,"OK profile created")
			} else if  strings.EqualFold(r.Method ,"GET"){
				if profile == nil{
					w.WriteHeader(http.StatusExpectationFailed)
					fmt.Fprint(w,"no profile available")
					return 
				}
				w.WriteHeader(http.StatusOK)
				w.Write(profile)
			}
			return
		}

		path, err := filepath.Abs(r.URL.Path)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		path = filepath.Join(staticPath, r.URL.Path)

		_, err = os.Stat(path)
		if os.IsNotExist(err) {
			// file does not exist, serve index.html
			http.ServeFile(w, r, filepath.Join(staticPath, indexPath))
			return
		} else if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fileServer.ServeHTTP(w, r)
	}))
}
