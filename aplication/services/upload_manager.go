package services

import (
	"context"
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/option"
)

type VideoUpload struct {
	Paths        []string
	VideoPath    string
	OutputBucket string
	Errors       []string
}

//NewVideoUpload - VideoUpload instance
func NewVideoUpload() *VideoUpload {
	return &VideoUpload{}
}

//UploadObject - Faz upload de um objeto para o GSC
func (vu *VideoUpload) UploadObject(objectPath string, client *storage.Client, ctx context.Context) error {
	//caminho/x/b/arquivo.mp4
	// split: caminho/x/b
	// [0] caminho/x/b
	// [1] arquivo.mp4
	path := strings.Split(objectPath, os.Getenv("localStoragePath")+"/")
	f, err := os.Open(objectPath)
	if err != nil {
		return err
	}

	defer f.Close()

	wc := client.Bucket(vu.OutputBucket).Object(path[1]).NewWriter(ctx)
	wc.ACL = []storage.ACLRule{{Entity: storage.AllUsers, Role: storage.RoleReader}}

	if _, err = io.Copy(wc, f); err != nil {
		return err
	}

	if err := wc.Close(); err != nil {
		return err
	}

	return nil
}

//loadPaths - Busca caminhos dos arquivos a serem feitos upload
func (vu *VideoUpload) loadPaths() error {

	err := filepath.Walk(vu.VideoPath, func(path string, info os.FileInfo, err error) error {

		if !info.IsDir() {
			vu.Paths = append(vu.Paths, path)
		}
		return nil
	})

	if err != nil {
		return err
	}
	return nil
}

//ProcessUpload -> Process de upload workers
func (vu *VideoUpload) ProcessUpload(concurrency int, doneUpload chan string) error {

	in := make(chan int, runtime.NumCPU()) // qual o arquivo baseado na posicao do slice Paths
	returnChannel := make(chan string)

	err := vu.loadPaths()
	if err != nil {
		return err
	}

	uploadClient, ctx, err := getClientUpload()
	if err != nil {
		return err
	}

	for process := 0; process < concurrency; process++ {
		go vu.uploadWorker(in, returnChannel, uploadClient, ctx)
	}

	go func() {
		for x := 0; x < len(vu.Paths); x++ {
			in <- x
		}
		close(in)
	}()

	for r := range returnChannel {
		if r != "" {
			doneUpload <- r
			break
		}
	}

	return nil

}

func (vu *VideoUpload) uploadWorker(in chan int, returnChan chan string, uploadClient *storage.Client, ctx context.Context) {

	for x := range in {
		err := vu.UploadObject(vu.Paths[x], uploadClient, ctx)

		if err != nil {
			vu.Errors = append(vu.Errors, vu.Paths[x])
			log.Printf("error during the upload: %v. Error: %v", vu.Paths[x], err)
			returnChan <- err.Error()
		}

		returnChan <- ""
	}

	returnChan <- "upload completed"
}

func getClientUpload() (*storage.Client, context.Context, error) {
	ctx := context.Background()
	fakegcs := os.Getenv("FAKEGCS") == "1"

	client, err := storage.NewClient(ctx)

	if fakegcs {
		transCfg := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // ignore expired SSL certificates
		}
		httpClient := &http.Client{Transport: transCfg}
		client, err = storage.NewClient(ctx, option.WithEndpoint("https://storage.gcs.io:4443/storage/v1/"), option.WithHTTPClient(httpClient))
	}

	if err != nil {
		return nil, nil, err
	}
	return client, ctx, nil
}
