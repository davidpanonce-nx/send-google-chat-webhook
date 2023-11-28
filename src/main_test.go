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
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

func TestgenerateMessageBody(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name            string
		ghJson          map[string]interface{}
		jobJson         map[string]interface{}
		timestamp       time.Time
		location        time.Location
		wantMessageBody map[string]interface{}
	}{
		{
			name: "test_success_workflow",
			ghJson: map[string]interface{}{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"pull_request_title": "Test Pull Request",
				"pull_request_author": "test-author",
				"pull_request_number": "123",
			},
			jobJson: map[string]interface{}{
				"status": "success",
			},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]interface{}{
				"cardsV2": map[string]interface{}{
					"cardId": "createCardMessage",
					"card": map[string]interface{}{
						"header": map[string]interface{}{
							"title":    fmt.Sprintf("Pull Request %s", "success"),
							"subtitle": fmt.Sprintf("Repository: %s", "test-repository"),
							"imageUrl": "https://github.githubassets.com/favicons/favicon.png",
						},
						"sections": []interface{}{
							map[string]interface{}{
								"widgets": []interface{}{
									map[string]interface{}{
										"decoratedText": map[string]interface{}{
											"text": fmt.Sprintf("<b>Title:</b> %s", "Test Pull Request"),
										},
									},
									map[string]interface{}{
										"decoratedText": map[string]interface{}{
											"text": fmt.Sprintf("<b>Author:</b> %s", "test-author"),
										},
									},
									map[string]interface{}{
										"buttonList": map[string]interface{}{
											"buttons": []interface{}{
												map[string]interface{}{
													"text": "Open",
													"onClick": map[string]interface{}{
														"openLink": map[string]interface{}{
															"url": fmt.Sprintf("https://github.com/%s/pull/%s",
																"test-repository", "123"),
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
			},
		},
		{
			name: "test_failed_workflow",
			ghJson: map[string]interface{}{
				"workflow":         "test-workflow",
				"ref":              "test-ref",
				"triggering_actor": "test-triggered_actor",
				"repository":       "test-repository",
				"pull_request_title": "Test Pull Request",
				"pull_request_author": "test-author",
				"pull_request_number": "123",
			},
			jobJson: map[string]interface{}{
				"status": "xxx",
			},
			timestamp: time.Date(2023, time.April, 25, 17, 44, 57, 0, time.UTC),
			wantMessageBody: map[string]interface{}{
				"cardsV2": map[string]interface{}{
					"cardId": "createCardMessage",
					"card": map[string]interface{}{
						"header": map[string]interface{}{
							"title":    fmt.Sprintf("Pull Request %s", "xxx"),
							"subtitle": fmt.Sprintf("Repository: %s", "test-repository"),
							"imageUrl": "https://github.githubassets.com/favicons/favicon-failure.png",
						},
						"sections": []interface{}{
							map[string]interface{}{
								"widgets": []interface{}{
									map[string]interface{}{
										"decoratedText": map[string]interface{}{
											"text": fmt.Sprintf("<b>Title:</b> %s", "Test Pull Request"),
										},
									},
									map[string]interface{}{
										"decoratedText": map[string]interface{}{
											"text": fmt.Sprintf("<b>Author:</b> %s", "test-author"),
										},
									},
									map[string]interface{}{
										"buttonList": map[string]interface{}{
											"buttons": []interface{}{
												map[string]interface{}{
													"text": "Open",
													"onClick": map[string]interface{}{
														"openLink": map[string]interface{}{
															"url": fmt.Sprintf("https://github.com/%s/pull/%s",
																"test-repository", "123"),
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
			},
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			gotMessageBody, err := generateMessageBody(tc.ghJson, tc.jobJson, tc.timestamp)
			if err != nil {
				t.Fatalf("failed to generate message body: %v", err)
			}

			wantMessageBodyByte, err := json.Marshal(tc.wantMessageBody)
			if err != nil {
				t.Fatalf("failed to marshal tc.wantMessageBody: %v", err)
			}

			if diff := cmp.Diff(wantMessageBodyByte, gotMessageBody); diff != "" {
				t.Errorf("messageBody got unexpected diff (-want, +got):\n%s", diff)
			}
		})
	}
}
