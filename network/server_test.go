/*
Copyright 2019 The Knative Authors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package network

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewServer_ProtocolSupport(t *testing.T) {
	tests := []struct {
		name           string
		clientHTTP1    bool
		clientH2C      bool
		expectProtocol string
	}{
		{
			name:           "HTTP/1 client",
			clientHTTP1:    true,
			clientH2C:      false,
			expectProtocol: "HTTP/1.1",
		},
		{
			name:           "H2C client",
			clientHTTP1:    false,
			clientH2C:      true,
			expectProtocol: "HTTP/2.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("test response"))
			})

			// Create server using NewServer
			server := httptest.NewUnstartedServer(handler)
			server.Config = NewServer(":0", handler)
			server.Start()
			defer server.Close()

			// Configure client with specified protocols
			protocols := &http.Protocols{}
			protocols.SetHTTP1(tt.clientHTTP1)
			protocols.SetUnencryptedHTTP2(tt.clientH2C)

			client := &http.Client{
				Transport: &http.Transport{
					Protocols: protocols,
				},
			}

			// Make request
			resp, err := client.Get(server.URL)
			if err != nil {
				t.Fatalf("Request failed: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				t.Errorf("Expected status 200, got %d", resp.StatusCode)
			}

			if resp.Proto != tt.expectProtocol {
				t.Errorf("Expected protocol %s, got %s", tt.expectProtocol, resp.Proto)
			}
		})
	}
}
