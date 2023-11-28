// Copyright 2023 The Authors (see AUTHORS file)
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/abcxyz/pkg/cli"
)

const (
	githubContextEnv = "GITHUB_CONTEXT"
	jobContextEnv    = "JOB_CONTEXT"
)

var rootCmd = func() cli.Command {
	return &cli.RootCommand{
		Name: "send-google-chat-webhook",
		Commands: map[string]cli.CommandFactory{
			"chat": func() cli.Command {
				return &cli.RootCommand{
					Name:        "workflownotification",
					Description: "notification for workflow",
					Commands: map[string]cli.CommandFactory{
						"workflownotification": func() cli.Command {
							return &WorkflowNotificationCommand{}
						},
					},
				}
			},
		},
	}
}

type WorkflowNotificationCommand struct {
	cli.BaseCommand
	flagWebhookUrl string
}

func (c *WorkflowNotificationCommand) Desc() string {
	return "Send a message to a Google Chat space"
}

func (c *WorkflowNotificationCommand) Help() string {
	return `
Usage: {{ COMMAND }} [options]

  The chat command sends messages to Google Chat spaces.
`
}

func (c *WorkflowNotificationCommand) Flags() *cli.FlagSet {
	set := c.NewFlagSet()

	f := set.NewSection("COMMAND OPTIONS")

	f.StringVar(&cli.StringVar{
		Name:    "webhook-url",
		Example: "https://chat.googleapis.com/v1/spaces/<SPACE_ID>/messages?key=<KEY>&token=<TOKEN>",
		Target:  &c.flagWebhookUrl,
		Usage:   `Webhook URL from google chat`,
	})

	return set
}

func (c *WorkflowNotificationCommand) Run(ctx context.Context, args []string) error {
	f := c.Flags()
	if err := f.Parse(args); err != nil {
		return fmt.Errorf("failed to parse flags: %w", err)
	}

	args = f.Args()
	if len(args) != 0 {
		return fmt.Errorf("expected 0 arguments, got %q", args)
	}

	ghJsonStr := c.GetEnv(githubContextEnv)
	if ghJsonStr == "" {
		return fmt.Errorf("environment var %s not set", githubContextEnv)
	}
	jobJsonStr := c.GetEnv(jobContextEnv)
	if jobJsonStr == "" {
		return fmt.Errorf("environment var %s not set", jobContextEnv)
	}

	ghJson := map[string]any{}
	jobJson := map[string]any{}
	if err := json.Unmarshal([]byte(ghJsonStr), &ghJson); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", githubContextEnv, err)
	}
	if err := json.Unmarshal([]byte(jobJsonStr), &jobJson); err != nil {
		return fmt.Errorf("failed unmarshaling %s: %w", jobContextEnv, err)
	}

	b, err := generateMessageBody(ghJson, jobJson, time.Now())
	if err != nil {
		return fmt.Errorf("failed to generate message body: %w", err)
	}

	url := c.flagWebhookUrl

	request, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(b))
	if err != nil {
		return fmt.Errorf("creating http request failed: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("sending http request failed: %w", err)
	}
	defer resp.Body.Close()

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		return fmt.Errorf("unexpected HTTP status code %d (%s)", got, http.StatusText(got))
	}

	return nil
}

func main() {
	ctx, done := signal.NotifyContext(context.Background(),
		syscall.SIGINT, syscall.SIGTERM)
	defer done()

	if err := realMain(ctx); err != nil {
		done()
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func realMain(ctx context.Context) error {
	return rootCmd().Run(ctx, os.Args[1:]) //nolint:wrapcheck // Want passthrough
}

func generateMessageBody(ghJson map[string]interface{}, jobJson map[string]interface{}, timestamp time.Time) ([]byte, error) {
    title := fmt.Sprintf("Pull Request %v", jobJson["status"])

    subtitle := fmt.Sprintf("Repository: %v", ghJson["repository"])

    imageUrl := "https://github.githubassets.com/favicons/favicon.png"

    decoratedText1 := fmt.Sprintf("<b>Title:</b> %v", ghJson["pull_request_title"])
    decoratedText2 := fmt.Sprintf("<b>Author:</b> %v", ghJson["pull_request_author"])

    openLinkURL := fmt.Sprintf("https://github.com/%v/pull/%v", ghJson["repository"], ghJson["pull_request_number"])

    jsonData := map[string]interface{}{
        "cardsV2": map[string]interface{}{
            "cardId": "createCardMessage",
            "card": map[string]interface{}{
                "header": map[string]interface{}{
                    "title":    title,
                    "subtitle": subtitle,
                    "imageUrl": imageUrl,
                },
                "sections": []interface{}{
                    map[string]interface{}{
                        "widgets": []interface{}{
                            map[string]interface{}{
                                "decoratedText": map[string]interface{}{
                                    "text": decoratedText1,
                                },
                            },
                            map[string]interface{}{
                                "decoratedText": map[string]interface{}{
                                    "text": decoratedText2,
                                },
                            },
                            map[string]interface{}{
                                "buttonList": map[string]interface{}{
                                    "buttons": []interface{}{
                                        map[string]interface{}{
                                            "text": "Open",
                                            "onClick": map[string]interface{}{
                                                "openLink": map[string]interface{}{
                                                    "url": openLinkURL,
                                                },
                                            },
                                        },
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }

    return json.Marshal(jsonData)
}
