package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"golang.org/x/oauth2"
)

var (
	tokenChan = make(chan string, 1)
	version   = "unknown"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := tokenFile
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	// run a callback server and change the oauth callback to match
	// TODO: would be better to get a random port?
	// https://stackoverflow.com/questions/43424787/how-to-use-next-available-port-in-http-listenandserve
	config.RedirectURL = "http://localhost:42871"
	callbackServer := http.Server{Addr: ":42871"}
	go startServer(callbackServer)
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf(`Go to the following link in your browser. This will kick of the OAuth workflow for %s.

%v

A callback web server is running waiting to receive the response from Google indicating you've completed the workflow.
`, os.Args[0], authURL)

	authCode := <-tokenChan
	_ = callbackServer.Shutdown(context.TODO()) // TODO handle this error?

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

const (
	errorHTML = `<h1>Google Task Mover</h1>
<p>There was an error completing the oauth Workflow: %s</p>`

	successHTML = `<h1>Google Task Mover</h1>
<p>Received the code from Google. You can close this window.</p>`
)

func startServer(srv http.Server) {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		err := r.URL.Query().Get("error")
		if len(err) != 0 {
			fmt.Fprintf(w, errorHTML, err)
			return
		}

		code := r.URL.Query().Get("code")
		tokenChan <- code
		fmt.Fprint(w, successHTML)
	})
	srv.ListenAndServe()
}
