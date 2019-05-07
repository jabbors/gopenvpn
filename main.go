package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/alecthomas/template"
	"github.com/go-chi/chi"
)

type vpnState int

func (s vpnState) String() string {
	switch s {
	case Disconnected:
		return "Disconnected"
	case Connecting:
		return "Connecting"
	case Active:
		return "Active"
	}

	return "Unknown"
}

func (s vpnState) Refresh() bool {
	switch s {
	case Disconnected:
		return false
	case Connecting:
		return true
	case Active:
		return false
	}

	return false
}

const (
	Disconnected vpnState = iota
	Connecting
	Active
	Done
)

var status vpnState

type vpnServer struct {
	Name   string
	Config string
}
type htmlPage struct {
	Refresh bool
	State   string
	Servers []vpnServer
}

type process struct {
	cmd      *exec.Cmd
	messages chan string
	state    chan<- vpnState
	done     chan<- error
	quit     chan error
}

func NewProcess(state chan<- vpnState, done chan<- error) *process {
	p := process{}
	p.state = state
	p.done = done
	p.quit = make(chan error, 1)
	return &p
}

func (p *process) Start(config string) error {
	p.cmd = exec.Command("openvpn", "--config", config)
	stdout, err := p.cmd.StdoutPipe()
	if err != nil {
		return err
	}

	p.messages = make(chan string, 10)
	// Spawn a go routine consuming messages
	go func(m <-chan string, s chan<- vpnState, q <-chan error) {
		for {
			select {
			case message := <-m:
				if strings.Contains(message, "Initialization Sequence Completed") {
					s <- Active
				}
			case <-q:
				return
			}
		}
	}(p.messages, p.state, p.quit)

	// Spawn a go routine parsing cmd output and sending it to a channel
	go func(rc io.ReadCloser, m chan<- string, q <-chan error) {
		var message []byte
		for {
			select {
			case <-q:
				return
			default:
				b := make([]byte, 80)
				n, err := rc.Read(b)
				if err != nil {
					if strings.Contains(err.Error(), os.ErrClosed.Error()) {
						return
					}
					log.Println("error reading cmd output:", err)
				}
				parts := bytes.Split(b[:n], []byte{10})
				splits := len(parts)
				for i, part := range parts {
					message = append(message, part...)
					if i != splits-1 {
						m <- string(message)
						message = message[:0]
					}
				}
			}
		}
	}(stdout, p.messages, p.quit)

	// Start the cmd
	err = p.cmd.Start()
	if err != nil {
		close(p.messages)
		close(p.quit)
		return err
	}
	p.state <- Connecting

	// Spawn a process waiting for the command the finnish
	go func() {
		p.done <- p.cmd.Wait()
	}()
	return nil
}

func (p *process) Stop() error {
	p.quit <- nil
	close(p.messages)
	return p.cmd.Process.Kill()
}

func serverConfigs(confDir string) ([]vpnServer, error) {
	tmpFiles, err := filepath.Glob(confDir + "/*.ovpn")
	if err != nil {
		return nil, err
	}

	servers := make([]vpnServer, len(tmpFiles))
	for i, tmpFile := range tmpFiles {
		servers[i] = vpnServer{Name: strings.TrimSuffix(path.Base(tmpFile), ".ovpn"), Config: tmpFile}
	}

	return servers, nil
}

func main() {
	var host string
	var port int
	var configDir string
	flag.StringVar(&host, "host", "127.0.0.1", "Host to bind to")
	flag.IntVar(&port, "port", 8080, "Port to bind to")
	flag.StringVar(&configDir, "config-dir", "/etc/openvpn", "Directoy with configurations")
	flag.Parse()

	state := make(chan vpnState, 1)
	done := make(chan error, 1)

	// Spawn a go routine updating internal state
	go func(s <-chan vpnState, d <-chan error) {
		for {
			select {
			case vpnState := <-s:
				status = vpnState
			case <-d:
				status = Disconnected
			}
		}
	}(state, done)

	var p *process

	r := chi.NewRouter()

	r.Get("/state", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/state")
		w.Header().Set("Content-Type", "text/plain")
		w.Write([]byte("VPN: "))
		switch status {
		case Disconnected:
			w.Write([]byte("Disconnected\n"))
		case Connecting:
			w.Write([]byte("Connecting\n"))
		case Active:
			w.Write([]byte("Active\n"))
		}
		w.Write([]byte("Number of goroutines: "))
		w.Write([]byte(fmt.Sprintf("%d\n", runtime.NumGoroutine())))
		w.Write([]byte("Process: "))
		w.Write([]byte(fmt.Sprintf("%v\n", p)))
	})
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		log.Println("/")
		tmpl, err := template.ParseFiles("./index.html")
		if err != nil {
			log.Println("unable to read index file:", err)
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("internal server error"))
			return
		}
		var servers []vpnServer
		if status == Disconnected {
			servers, err = serverConfigs(configDir)
			if err != nil {
				log.Println("unable to get server configs:", err)
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("internal server error"))
				return
			}
		}
		data := htmlPage{Refresh: status.Refresh(), State: status.String(), Servers: servers}
		tmpl.Execute(w, data)
	})
	r.Get("/start", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.String())
		if status == Disconnected {
			configs, ok := r.URL.Query()["config"]
			if !ok || len(configs[0]) < 1 {
				log.Println("url param 'config' is missing")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("bad request"))
				return
			}
			config := configs[0]

			proc := NewProcess(state, done)
			err := proc.Start(config)
			if err != nil {
				log.Println("unable to start process:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			p = proc
		}
		http.Redirect(w, r, "/", 307)
	})
	r.Get("/stop", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.String())
		if status != Disconnected {
			if p == nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			err := p.Stop()
			if err != nil {
				log.Println("unable to start process:", err)
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			p = nil
			status = Disconnected
		}
		http.Redirect(w, r, "/", 307)
	})
	r.Get("/reset", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL.String())
		if p != nil {
			p.Stop()
			p = nil
		}
		status = Disconnected
		http.Redirect(w, r, "/", 307)
	})

	// Fire up HTTP handler
	srv := &http.Server{Addr: fmt.Sprintf("%s:%d", host, port), Handler: r}
	log.Printf("Listening for requests on %s:%d", host, port)
	err := srv.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			return
		}
		log.Fatal(err)
	}
}
