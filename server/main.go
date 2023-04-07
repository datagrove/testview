package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"embed"
	_ "embed"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gliderlabs/ssh"
	"github.com/pkg/sftp"
	"github.com/spf13/cobra"
)

// testview --port 5078 --sftp localhost:5079 --http localhost:5078 --store ./TestResults

var (
	//go:embed dist/**
	res embed.FS
)

var config TestValue

func main() {
	h, _ := os.UserHomeDir()
	config = TestValue{
		Http:  "localhost:5078",
		Sftp:  "localhost:5079",
		Store: "TestResults",
		Key:   path.Join(h, ".ssh", "id_rsa"),
	}

	rootCmd := &cobra.Command{
		Use: "testview ",
		Run: func(cmd *cobra.Command, args []string) {
			log.Printf("%v", config)
			launch()
		}}

	rootCmd.PersistentFlags().StringVar(&config.Http, "http", ":5078", "http address")
	rootCmd.PersistentFlags().StringVar(&config.Sftp, "sftp", ":5079", "sftp address")
	rootCmd.PersistentFlags().StringVar(&config.Store, "store", "TestResults", "test result store")
	rootCmd.Execute()
}

// SftpHandler handler for SFTP subsystem
func SftpHandlerx(sess ssh.Session) {
	debugStream := ioutil.Discard
	serverOptions := []sftp.ServerOption{
		sftp.WithDebug(debugStream),
	}
	server, err := sftp.NewServer(
		sess,
		serverOptions...,
	)
	if err != nil {
		log.Printf("sftp server init error: %s\n", err)
		return
	}

	if err := server.Serve(); err == io.EOF {
		server.Close()
		fmt.Println("sftp client exited session.")
	} else if err != nil {
		fmt.Println("sftp server completed with error:", err)
	}
}

func launch() {
	go func() {

		ssh_server := ssh.Server{
			Addr: config.Sftp,
			PublicKeyHandler: func(ctx ssh.Context, key ssh.PublicKey) bool {
				return true
			},
			SubsystemHandlers: map[string]ssh.SubsystemHandler{
				"sftp": SftpHandlerx,
			},
		}
		kf := ssh.HostKeyFile(config.Key)
		kf(&ssh_server)
		log.Fatal(ssh_server.ListenAndServe())
	}()

	mux := http.NewServeMux()
	var staticFS = fs.FS(res)
	htmlContent, err := fs.Sub(staticFS, "dist")
	if err != nil {
		log.Fatal(err)
	}
	fs := http.FileServer(http.FS(htmlContent))
	mux.Handle("/", fs)

	mux.Handle("/TestResults/", http.StripPrefix("/TestResults/", http.FileServer(http.Dir(config.Store))))

	mux.HandleFunc("/api/runs", func(w http.ResponseWriter, r *http.Request) {
		dir := []string{}
		os.Mkdir(config.Store, 0777)
		d, e := os.ReadDir(config.Store)
		if e != nil {
			return
		}
		for _, batch := range d {
			dir = append(dir, batch.Name())
		}
		json.NewEncoder(w).Encode(dir)
	})
	mux.HandleFunc("/api/run/", func(w http.ResponseWriter, r *http.Request) {
		batch := path.Join(config.Store, r.URL.Path[8:])

		// index.json written at beginning of each test, it lets us know what files are expected
		root := []string{}
		b, e := os.ReadFile(path.Join(batch, "index.json"))
		if e != nil {
			return
		}

		json.Unmarshal(b, &root)

		testcode := map[string]string{}
		for _, feature := range root {
			testcode[feature] = "waiting"
		}

		failed := map[string]bool{}

		d, e := os.ReadDir(batch)
		if e != nil {
			log.Printf("%v", e)
		}
		// I don't need this, web can just look for the file
		for _, f := range d {
			if f.IsDir() {
				continue
			}
			//fn := path.Base(f.Name())
			p := strings.Split(f.Name(), ".")
			ext := p[len(p)-1]
			tn := f.Name()[:len(f.Name())-len(ext)-1]
			switch ext {
			case "error":
				failed[tn] = true
			case "txt": // error
				testcode[tn] = "pass"
			}
		}
		for key := range failed {
			testcode[key] = "fail"
		}

		json.NewEncoder(w).Encode(testcode)
	})

	err = http.ListenAndServe(config.Http, mux)
	log.Fatal(err)
}
func index(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
}

type TestValue struct {
	Key   string `json:"key,omitempty"`
	Http  string `json:"http,omitempty"`
	Sftp  string `json:"sftp,omitempty"`
	Store string `json:"test_root,omitempty"`
}

// Generates a private key that will be used by the SFTP server.
func generatePrivateKey(BasePath string) error {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(path.Join(BasePath, ".sftp"), 0755); err != nil {
		return err
	}

	o, err := os.OpenFile(path.Join(BasePath, ".sftp/id_rsa"), os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer o.Close()

	pkey := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(key),
	}

	if err := pem.Encode(o, pkey); err != nil {
		return err
	}

	return nil
}
