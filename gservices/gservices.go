package gservices

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"regexp"
	"sync"
	"time"

	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/gmail/v1"
)

type Mail struct {
	Id      string
	Date    string
	Subject string
}

func New() (client *http.Client) {
	ctx := context.Background()
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	config, err := google.ConfigFromJSON(b, gmail.GmailReadonlyScope,
		gmail.GmailComposeScope, gmail.GmailModifyScope, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client = getClient(ctx, config)
	return
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	log.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)
	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("gateway.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	log.Printf("Saving credential file to: %s\n", file)
	f, err := os.Create(file)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func GetDuty(client *http.Client, calendarId string) (string, error) {
	srv, err := calendar.New(client)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve calendar Client %v", err)
	}

	t := time.Now().Format(time.RFC3339)
	events, err := srv.Events.List(calendarId).ShowDeleted(false).
		SingleEvents(true).TimeMin(t).MaxResults(1).OrderBy("startTime").Do()
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve event. %v", err)
	}

	if len(events.Items) == 1 {
		return regexp.MustCompile(`\+\d{12}`).FindString(events.Items[0].Description), nil
	} else if len(events.Items) == 0 {
		return "", errors.New("No events found.")
	} else {
		log.Println("Found %d events - using only the first one.")
		return regexp.MustCompile(`\+\d{12}`).FindString(events.Items[0].Description), nil
	}
}

func GetMail(client *http.Client) ([]Mail, error) {
	srv, err := gmail.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"
	var pageToken string
	var wg sync.WaitGroup
	messages := []Mail{}
	for {
		req := srv.Users.Messages.List(user).LabelIds("UNREAD")
		if pageToken != "" {
			req.PageToken(pageToken)
		}
		r, err := req.Do()
		if err != nil {
			return messages, fmt.Errorf("Unable to retrieve messages. %v", err)
		}
		log.Printf("Number of messages received = %d.\n", r.ResultSizeEstimate)
		c := make(chan *gmail.Message, r.ResultSizeEstimate)
		for _, msg := range r.Messages {
			wg.Add(1)
			go func(msg *gmail.Message) {
				defer wg.Done()
				log.Println("Requesting message", msg.Id)
				messageData, err := srv.Users.Messages.Get(user, msg.Id).Do()
				if err != nil {
					log.Println("Unable to retrieve message %v: %v", msg.Id, err)
					return
					//	return messages, fmt.Errorf("Unable to retrieve message %v: %v", msg.Id, err)
				}
				log.Println("Received", messageData.Id, messageData.LabelIds)
				c <- messageData
				log.Println("Send down the channel", messageData.Id)
			}(msg)
		}
		wg.Wait()
		close(c)
		for messageData := range c {
			subject := ""
			date := ""
			for _, header := range messageData.Payload.Headers {
				if header.Name == "Subject" {
					subject = header.Value
				} else if header.Name == "Date" {
					date = header.Value
				}
			}
			messages = append(messages, Mail{
				Id:      messageData.Id,
				Date:    date,
				Subject: subject,
			})
		}
		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken
		log.Println("Moving to the next page")
	}
	return messages, nil
}

func ReadMail(client *http.Client, msg Mail) error {
	srv, err := gmail.New(client)
	if err != nil {
		return fmt.Errorf("Unable to retrieve gmail Client %v", err)
	}

	user := "me"
	log.Println("Reading message", msg.Id)
	_, err = srv.Users.Messages.Modify(user, msg.Id, &gmail.ModifyMessageRequest{
		AddLabelIds:    nil,
		RemoveLabelIds: []string{"UNREAD"},
	}).Do()
	if err != nil {
		return fmt.Errorf("Failed to read an email: %v", err)
	}
	log.Println("Read message", msg.Id)
	return nil
}
