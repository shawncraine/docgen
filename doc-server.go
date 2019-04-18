package main

import (
	"bytes"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/fsnotify/fsnotify"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Doc represnet in-memory html/markdown documentation
type Doc struct {
	Name           string
	HTML, Markdown []byte
}

var (
	tmDoc        *template.Template
	docs         = make([]Doc, 0)
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

	// build tempaltes
	buildTemplates()
	// load documentations
	loadDocumentations()

	http.HandleFunc("/login", getLogin)
	http.HandleFunc("/post-login", postLogin)
	http.HandleFunc("/logout", logout)

	http.HandleFunc("/", basicAuth(home))
	http.HandleFunc("/docs", basicAuth(viewDoc))
	http.HandleFunc("/upload-doc", basicAuth(uploadDoc))
	http.HandleFunc("/delete-doc", basicAuth(deleteDoc))
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
	username := ""
	c, err := r.Cookie(cookieName)
	if err == nil {
		username = sess[c.Value]
	}
	type app struct {
		Name     string
		Username string
		Message  string
		Docs     []Doc
		CRUD     bool
	}
	data := struct {
		Assets Assets
		Data   app
	}{
		Assets: assets,
		Data: app{
			Name:     viper.GetString("app.name"),
			Username: username,
			Docs:     docs,
			CRUD:     hasCRUDpermission(username),
		},
	}
	if f := r.URL.Query().Get("code"); f != "" {
		switch f {
		default:
			data.Data.Message = ""
		case "0":
			data.Data.Message = "No API documentation found!"
		}
	}
	buf := new(bytes.Buffer)
	if err := tmDoc.ExecuteTemplate(buf, "home", data); err != nil {
		log.Fatal(err)
	}
	w.Header().Add("Content-Type", "text/html")
	w.Write(buf.Bytes())
}

func viewDoc(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	typ := r.URL.Query().Get("type")
	if typ == "" {
		typ = "html"
	}
	bb := []byte{}
	for _, d := range docs {
		if strings.ToLower(d.Name) == strings.ToLower(name) {
			if typ == "md" {
				bb = d.Markdown
			} else {
				bb = d.HTML
			}
			break
		}
	}
	if len(bb) == 0 {
		http.Redirect(w, r, "/home?code=0", http.StatusSeeOther)
		return
	}
	w.Header().Add("Content-Type", "text/html")
	w.Write(bb)
}

func uploadDoc(w http.ResponseWriter, r *http.Request) {
	max := int64(1024 * 1024 * 600)
	r.ParseMultipartForm(max)
	fname := r.FormValue("name")
	f, h, err := r.FormFile("file")
	if err != nil || h.Size > max {
		log.Println(err)
		w.Write([]byte("422"))
		return
	}
	bb, err := ioutil.ReadAll(f)
	if err != nil {
		log.Println(err)
		w.Write([]byte("422"))
		return
	}
	if err := ioutil.WriteFile(fmt.Sprintf("%s/%s.json", viper.GetString("app.dir"), fname), bb, 0644); err != nil {
		log.Println(err)
		w.Write([]byte("422"))
		return
	}
	//load the documentations
	loadDocumentations()
	w.Write([]byte("200"))
}

func deleteDoc(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	fname := r.FormValue("name")
	if err := os.Remove(fmt.Sprintf("%s/%s.json", viper.GetString("app.dir"), fname)); err != nil {
		log.Println(err)
		w.Write([]byte("422"))
		return
	}
	//load the documentations
	loadDocumentations()
	w.Write([]byte("200"))
}

func loadConfig() {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err := viper.ReadInConfig()   // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}
}

func buildTemplates() {
	fMap := template.FuncMap{
		"html":            html,
		"css":             css,
		"js":              js,
		"snake":           snake,
		"color":           color,
		"trimQueryParams": trimQueryParams,
		"date_time":       dateTime,
		"markdown":        markdown,
	}
	tmDoc = template.Must(template.New("login").Delims("@{{", "}}@").Funcs(fMap).Parse(assets.DocServerLogin))
	template.Must(tmDoc.New("home").Funcs(fMap).Delims("@{{", "}}@").Parse(assets.DocServerHome))
}

func loadDocumentations() {
	directory := viper.GetString("app.dir")
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		os.Mkdir(directory, os.ModePerm)
	}
	const scannExt = ".json"
	docs = make([]Doc, 0)
	err := filepath.Walk(directory, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == scannExt {
			hBuf := readJSONtoHTML(path)
			mBuf := readJSONtoMarkdownHTML(path)
			docs = append(docs, Doc{
				Name:     strings.TrimSuffix(filepath.Base(path), scannExt),
				HTML:     hBuf.Bytes(),
				Markdown: mBuf.Bytes(),
			})
		}
		return nil
	})
	if err != nil {
		log.Printf("loadDocuments: %v\n", err)
	}
}
