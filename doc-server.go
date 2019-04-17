package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/spf13/cobra"
)

var (
	dir          string
	docServerCMD = &cobra.Command{
		Use:   "doc-server",
		Short: "Doc run a full documentaion server",
		Long:  `Doc run a full documentaion server`,
		Run:   docServer,
	}
)

func init() {
	docServerCMD.PersistentFlags().IntVarP(&port, "port", "p", 9000, "port number to listen")
	docServerCMD.PersistentFlags().StringVarP(&dir, "dir", "d", "", "directory path where the uploaded files will reside")
	docServerCMD.PersistentFlags().BoolVarP(&isMarkdown, "md", "m", false, "display markdown format in preview")
}

func docServer(cmd *cobra.Command, args []string) {
	http.HandleFunc("/", home)
	log.Println("Listening on port: ", port)
	log.Printf("Web Server is available at http://localhost:%s/\n", strconv.Itoa(port))
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatalf("Failed to run http server: %v", err)
	}
}

// TODO:
// 1. Run http doc server
// 2. User can login
// 3. User can upload postman collection
// 4. User can list the available documentations
// 5. User can see individual documentaion (both->html/markdown)

func home(w http.ResponseWriter, r *http.Request) {
	buf := readJSONtoMarkdownHTML("./example.json")
	if err := ioutil.WriteFile("./files/example-md.html", buf.Bytes(), 0644); err != nil {
		panic(err)
	}
	bb, err := ioutil.ReadFile("./files/example-md.html")
	if err != nil {
		panic(err)
	}
	w.Header().Add("Content-Type", "text/html")
	w.Write(bb)
}
