package main

import (
	"flag"
	"net/http"
	"log"
	"fmt"
	"encoding/json"
	"os"
	"io"
	"bytes"
)

var secret string;

var sentry string;
var host string = "0.0.0.0";
var port int = 8000;

var config *Config;

type Config struct {
	Host   string
	Port   int
	Sentry string
	Secret string
}

type PushEvent struct {
	Sha        string            `json:"checkout_sha"`
	Repository *GitlabRepository `json:"repository"`
	Commits    []*GitlabCommit   `json:"commits"`
}

type GitlabRepository struct {
	Homepage string `json:"homepage"`
}

type GitlabCommit struct {
	Id        string            `json:"id"`
	Message   string            `json:"message"`
	Timestamp string            `json:"timestamp"`
	Author    *PushCommitAuthor `json:"author"`
}

type PushCommitAuthor struct {
	Name  string `json:"name"`
	Email string `json:"email"`
}

type SentryPayload struct {
	Version string                 `json:"version"`
	Ref     string                 `json:"ref"`
	Url     string                 `json:"url"`
	Commits []*SentryPayloadCommit `json:"commits"`
}

type SentryPayloadCommit struct {
	Id          string `json:"id"`
	Message     string `json:"message"`
	AuthorName  string `json:"author_name"`
	AuthorEmail string `json:"author_email"`
	Timestamp   string `json:"timestamp"`
}

func main() {
	config = configure()

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	http.HandleFunc("/push", pushHandler)
	log.Printf("Listen %s\n", addr)
	http.ListenAndServe(addr, nil)
}

func configure() *Config {
	if token := os.Getenv("GITLAB_TOKEN"); len(token) != 0 {
		secret = token
	}
	if envUrl := os.Getenv("SENTRY_URL"); len(envUrl) != 0 {
		secret = envUrl
	}
	if envHost := os.Getenv("APP_HOST"); len(envHost) != 0 {
		host = envHost
	}
	if envPort := os.Getenv("APP_PORT"); len(envPort) != 0 {
		host = envPort
	}
	config = &Config{}

	flag.StringVar(&config.Sentry, "sentry", sentry, "url for sentry releases")
	flag.StringVar(&config.Host, "host", host, "host")
	flag.IntVar(&config.Port, "port", port, "port")
	flag.StringVar(&config.Secret, "secret", secret, "gitlab token")
	flag.Parse()
	return config
}

func check(r *http.Request) bool {
	passedToken := r.Header.Get("X-Gitlab-Token")
	return len(config.Secret) == 0 || passedToken == config.Secret
}

func decodeEvent(r io.Reader) *PushEvent {
	decoder := json.NewDecoder(r)
	event := &PushEvent{}
	decoder.Decode(event)
	return event
}

func buildPayloadFromPush(event *PushEvent) *SentryPayload {
	payload := &SentryPayload{
		Version: event.Sha[0: 7],
		Ref:     event.Sha,
		Url:     fmt.Sprintf("%s/tree/%s", event.Repository.Homepage, event.Sha),
		Commits: []*SentryPayloadCommit{},
	}

	for _, inCommit := range event.Commits {
		payload.Commits = append(payload.Commits, &SentryPayloadCommit{
			Id:          inCommit.Id,
			Message:     inCommit.Message,
			AuthorName:  inCommit.Author.Name,
			AuthorEmail: inCommit.Author.Email,
			Timestamp:   inCommit.Timestamp,
		})
	}
	return payload
}

func pushHandler(w http.ResponseWriter, r *http.Request) {
	if !check(r) {
		w.Write([]byte("Invalid gitlab token"))
		return
	}
	event := decodeEvent(r.Body)

	payload := buildPayloadFromPush(event)

	sendToSentry(payload)
}

func sendToSentry(payload *SentryPayload) {

	bts, err := json.Marshal(payload)

	log.Printf("REQUEST:\n%s\n", string(bts))
	resp, err := http.Post(config.Sentry, "application/json", bytes.NewBuffer(bts))
	if err != nil {
		log.Printf("RESPONSE ERROR:\n%s\n", err)
	} else {
		buff := &bytes.Buffer{};
		buff.ReadFrom(resp.Body)
		fmt.Printf("RESPONSE:\n%s\n", buff.String())
	}
}
