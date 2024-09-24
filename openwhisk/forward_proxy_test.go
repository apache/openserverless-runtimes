package openwhisk

import (
	"bytes"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Example_forwardInitRequest() {
	// create a client ActionProxy
	clientAP := NewActionProxy("", "", nil, nil, ProxyModeClient)

	// create a server ActionProxy
	compiler, _ := filepath.Abs("common/gobuild.py")
	log, _ := os.CreateTemp("", "log")
	serverAP := NewActionProxy("./action", compiler, log, log, ProxyModeServer)

	// start the server
	ts := httptest.NewServer(serverAP)

	// create a request body
	body := initBinary("_test/hello.zip", "@"+ts.URL)
	initBody := bytes.NewBufferString(body)
	w := httptest.NewRecorder()

	// run the client init request
	clientAP.initHandler(w, httptest.NewRequest(http.MethodPost, "/init", initBody))

	// stop the server
	runtime.Gosched()
	// wait 2 seconds before declaring a test done
	time.Sleep(2 * time.Second)
	ts.Close()
	dump(log)

	fmt.Println(w.Body.String())

	// Output:
	// {"ok":true}
}

func Example_forwardRunRequest() {
	// create a client ActionProxy
	clientAP := NewActionProxy("", "", nil, nil, ProxyModeClient)

	// create a server ActionProxy
	compiler, _ := filepath.Abs("common/gobuild.py")
	log, _ := os.CreateTemp("", "log")
	serverAP := NewActionProxy("./action", compiler, log, log, ProxyModeServer)

	// start the server
	ts := httptest.NewServer(serverAP)

	// create a request body
	body := initBinary("_test/hello.zip", "@"+ts.URL)
	initBody := bytes.NewBufferString(body)
	w := httptest.NewRecorder()

	// run the client init request
	clientAP.initHandler(w, httptest.NewRequest(http.MethodPost, "/init", initBody))
	fmt.Print(w.Body.String())

	// create a request body
	runW := httptest.NewRecorder()
	runBody := bytes.NewBufferString(`{"value": {"name": "Mike"}}`)
	clientAP.runHandler(runW, httptest.NewRequest(http.MethodPost, "/run", runBody))

	// stop the server
	runtime.Gosched()
	// wait 2 seconds before declaring a test done
	time.Sleep(2 * time.Second)
	ts.Close()
	dump(log)
	fmt.Println(runW.Body.String())

	// Output:
	// {"ok":true}
	// Main
	// Hello, Mike
	// XXX_THE_END_OF_A_WHISK_ACTIVATION_XXX
	// XXX_THE_END_OF_A_WHISK_ACTIVATION_XXX
	// {"greetings":"Hello, Mike"}
}

func TestParseMainFlag(t *testing.T) {
	tests := []struct {
		name          string
		mainAtProxy   string
		expectedMain  string
		expectedProxy string
		expectError   bool
	}{
		{
			name:          "Valid main and proxy",
			mainAtProxy:   "mainFunc@https://example.com",
			expectedMain:  "mainFunc",
			expectedProxy: "https://example.com",
			expectError:   false,
		},
		{
			name:          "Valid proxy only",
			mainAtProxy:   "https://example.com",
			expectedMain:  "",
			expectedProxy: "https://example.com",
			expectError:   false,
		},
		{
			name:        "Invalid main flag format",
			mainAtProxy: "mainFunc@https://example.com@extra",
			expectError: true,
		},
		{
			name:        "Invalid URL",
			mainAtProxy: "mainFunc@://invalid-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			proxyData, err := parseMainFlag(tt.mainAtProxy)

			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedMain, proxyData.MainFunc)
			require.Equal(t, tt.expectedProxy, proxyData.ProxyURL.String())
		})
	}
}
func TestParseMainURL(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectedURL string
		expectError bool
	}{
		{
			name:        "Valid URL with scheme",
			input:       "https://example.com",
			expectedURL: "https://example.com",
			expectError: false,
		},
		{
			name:        "Valid URL without scheme",
			input:       "example.com",
			expectedURL: "https://example.com",
			expectError: false,
		},
		{
			name:        "Invalid URL",
			input:       "://invalid-url",
			expectError: true,
		},
		{
			name:        "Empty URL",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsedURL, err := parseMainURL(tt.input)
			if tt.expectError {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expectedURL, parsedURL.String())
		})
	}
}
