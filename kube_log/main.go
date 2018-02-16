package main

import (
	"bufio"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/websocket"

	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/pkg/api/v1"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	addr     = "127.0.0.1:8080"
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	kubeConfigPath = "/home/dipta/.kube/config"
	namespace      = "dipta"
	podName        = "grpc-example-dep-1858803462-jcf09"
)

func writer(ws *websocket.Conn) {
	defer func() {
		ws.Close()
	}()

	rd, err := getLogReader()
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte(err.Error()))
		return
	}

	scanner := bufio.NewScanner(rd)

	for scanner.Scan() {
		if err := ws.WriteMessage(websocket.TextMessage, []byte(scanner.Text())); err != nil {
			return
		}
	}
}

func serveWs(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		if _, ok := err.(websocket.HandshakeError); !ok {
			log.Println(err)
		}
		return
	}

	go writer(ws)
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "Not found", 404)
		return
	}
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", 405)
		return
	}
	http.ServeFile(w, r, "home.html")
}

func getLogReader() (io.ReadCloser, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	clientSet, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	_, err = clientSet.CoreV1().Pods(namespace).Get(podName, meta_v1.GetOptions{})
	if err != nil {
		return nil, err
	}

	req := clientSet.CoreV1().Pods(namespace).GetLogs(
		podName,
		&v1.PodLogOptions{
			Follow: true,
		},
	)

	out, err := req.Stream()
	if err != nil {
		return nil, err
	}

	return out, nil
}

func main() {
	http.HandleFunc("/", serveHome)
	http.HandleFunc("/ws", serveWs)
	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}
