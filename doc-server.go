package main

import (
	"bytes"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"

	"github.com/fsnotify/fsnotify"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Doc represnet in-memory html/markdown documentation
type Doc struct {
	HTML, Markdown string
}

var (
	docs         = make(map[string]Doc, 0)
	dir          string
	docServerCMD = &cobra.Command{
		Use:   "doc-server",
		Short: "Doc run a full documentaion server",
		Long:  `Doc run a full documentaion server`,
		Run:   docServer,
		PreRun: func(cmd *cobra.Command, args []string) {
			loadConfig()
		},
	}
)

func init() {
	docServerCMD.PersistentFlags().IntVarP(&port, "port", "p", 9000, "port number to listen")
	docServerCMD.PersistentFlags().StringVarP(&dir, "dir", "d", "", "directory path where the uploaded files will reside")
	docServerCMD.PersistentFlags().BoolVarP(&isMarkdown, "md", "m", false, "display markdown format in preview")
}

func docServer(cmd *cobra.Command, args []string) {
	// watch configs change
	go func() {
		viper.WatchConfig()
		viper.OnConfigChange(func(e fsnotify.Event) {
			fmt.Println("Loading new configs...")
			loadConfig()
			fmt.Printf("Config [%s] loaded!\n", e.Name)
		})
	}()

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
	tm := template.New("main")
	tm.Delims("@{{", "}}@")
	tm.Funcs(template.FuncMap{
		"html":            html,
		"css":             css,
		"js":              js,
		"snake":           snake,
		"color":           color,
		"trimQueryParams": trimQueryParams,
		"date_time":       dateTime,
		"markdown":        markdown,
	})
	t, err := tm.Parse(assets.DocServerHome)
	if err != nil {
		log.Fatal(err)
	}
	type app struct {
		Name string
	}
	data := struct {
		Assets Assets
		Data   app
	}{
		Assets: assets,
		Data: app{
			Name: viper.GetString("app.name"),
		},
	}
	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func loadConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}
