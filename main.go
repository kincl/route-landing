package main

import (
	"context"
	"embed"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	v1 "github.com/openshift/api/route/v1"
	routev1 "github.com/openshift/client-go/route/clientset/versioned/typed/route/v1"
)

//go:embed assets/*
var assetData embed.FS

type Homepage struct {
	routes []v1.Route
}

func (h *Homepage) start() error {
	kubeconfig := flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	flag.Parse()

	// use the current context in kubeconfig
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return err
	}

	routeV1Client, err := routev1.NewForConfig(config)
	if err != nil {
		return err
	}

	// get all routes
	routes, err := routeV1Client.Routes("").List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return err
	}
	fmt.Printf("There are %d routes in the cluster\n", len(routes.Items))
	h.routes = append(h.routes, routes.Items...)
	return nil
}

func (h *Homepage) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t := template.New("index")
	t = t.Funcs(template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }})

	index, _ := assetData.ReadFile("assets/index.html")
	t, err := t.Parse(string(index))
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
	}

	err = t.ExecuteTemplate(w, "index", h.routes)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
	}
}

func main() {
	home := new(Homepage)

	err := home.start()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err)
		os.Exit(1)
	}

	http.Handle("/", home)
	http.Handle("/assets/", http.FileServer(http.FS(assetData)))
	// http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
	// 	fmt.Fprintf(w, "Hello, %v\n", html.EscapeString(r.URL.Path))
	// })

	fmt.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
