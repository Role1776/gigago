package gigago

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_Generate(t *testing.T) {
	var testCases = []struct {
		name               string
		apiKey             string
		systemInstruction  string
		inputMessages      []Message
		mockAIStatus       int
		mockAIResponse     *CompletionResponse
		mockAIRawResponse  string
		mockOauthStatus    int
		expectedOutput     string
		expectedOauthError error
		expectedGenError   error
	}{
		{
			name:              "Success",
			apiKey:            "FakeKey",
			systemInstruction: "You are a travel guide. Answer the user's questions.",
			inputMessages: []Message{
				{Role: RoleUser, Content: "The capital of France is"},
			},
			mockAIStatus:    http.StatusOK,
			mockOauthStatus: http.StatusOK,
			mockAIResponse: &CompletionResponse{
				Choices: []Choice{
					{
						Message: ResponseMessage{
							Content: "Paris.",
						},
					},
				},
			},
			expectedOutput: "Paris.",
		},
		{
			name:            "Failure_ClientCreation_OauthError",
			apiKey:          "FailOauth",
			mockOauthStatus: http.StatusInternalServerError,
			inputMessages: []Message{
				{Role: RoleUser, Content: "The capital of France is"},
			},
			expectedOauthError: errors.New("oauth request failed with status 500"),
		},
		{
			name:            "Failure_ClientCreation_OauthUnauthorized",
			apiKey:          "FakeKey",
			mockOauthStatus: http.StatusUnauthorized,
			inputMessages: []Message{
				{Role: RoleUser, Content: "The capital of France is"},
			},
			expectedOauthError: errors.New("oauth request failed with status 401"),
		},
		{
			name: "Failure_Generate_APIError",
			inputMessages: []Message{
				{Role: RoleUser, Content: "The capital of France is"},
			},
			apiKey:           "FakeKey",
			mockAIStatus:     http.StatusInternalServerError,
			mockOauthStatus:  http.StatusOK,
			expectedGenError: errors.New("unexpected status 500"),
		},
		{
			name:             "Failure_Generate_EmptyInput",
			apiKey:           "FakeKey",
			inputMessages:    []Message{},
			mockOauthStatus:  http.StatusOK,
			expectedGenError: errors.New("empty message"),
		},
		{
			name: "Failure_Generate_InvalidJSONResponse",
			inputMessages: []Message{
				{Role: RoleUser, Content: "Give me bad JSON"},
			},
			apiKey:            "FakeKey",
			mockAIStatus:      http.StatusOK,
			mockOauthStatus:   http.StatusOK,
			mockAIRawResponse: `error, not json`,
			expectedGenError:  errors.New("invalid character 'e' looking for beginning of value"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			serverAI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(testCase.mockAIStatus)
				if testCase.mockAIRawResponse != "" {
					_, err := w.Write([]byte(testCase.mockAIRawResponse))
					if err != nil {
						t.Fatalf("Failed to write raw response: %v", err)
					}
					return
				}
				if testCase.mockAIResponse != nil {
					if err := json.NewEncoder(w).Encode(testCase.mockAIResponse); err != nil {
						t.Fatalf("Failed to encode response in mock server: %v", err)
					}
				}
			}))
			defer serverAI.Close()

			serverOauth := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(testCase.mockOauthStatus)
				if err := json.NewEncoder(w).Encode(struct{}{}); err != nil {
					t.Fatalf("Failed to encode response in mock server: %v", err)
				}
			}))
			defer serverOauth.Close()

			client, err := NewClient(context.Background(), testCase.apiKey, WithCustomURLAI(serverAI.URL), WithCustomURLOauth(serverOauth.URL))
			if testCase.expectedOauthError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedOauthError.Error())
				return
			}
			require.NoError(t, err)
			defer client.Close()

			model := client.GenerativeModel("GigaChat")
			model.SystemInstruction = testCase.systemInstruction
			model.Temperature = 0.7

			resp, err := model.Generate(context.Background(), testCase.inputMessages)

			if testCase.expectedGenError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedGenError.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, resp)
				require.NotEmpty(t, resp.Choices)
				assert.Equal(t, testCase.expectedOutput, resp.Choices[0].Message.Content)
			}
		})
	}
}

func TestNewClient(t *testing.T) {
	var testCases = []struct {
		name            string
		apiKey          string
		mockStatusCode  int
		mockRawResponse string
		mockResponse    interface{}
		expectedToken   *tokenResponse
		expectedError   error
	}{
		{
			name:           "Success",
			apiKey:         "testKey",
			mockStatusCode: http.StatusOK,
			mockResponse: &tokenResponse{
				AccessToken: "token",
				ExpiresAt:   13132454545,
			},
			expectedToken: &tokenResponse{
				AccessToken: "token",
				ExpiresAt:   13132454545,
			},
		},
		{
			name:           "Failure_ServerError",
			apiKey:         "testKey",
			mockStatusCode: http.StatusInternalServerError,
			mockResponse:   nil,
			expectedError:  errors.New("oauth request failed with status 500"),
		},
		{
			name:            "Failure_InvalidJSONResponse",
			apiKey:          "testKey",
			mockStatusCode:  http.StatusOK,
			mockResponse:    nil,
			mockRawResponse: `error, not json`,
			expectedError:   errors.New("invalid character 'e' looking for beginning of value"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(testCase.mockStatusCode)
				if testCase.mockRawResponse != "" {
					_, err := w.Write([]byte(testCase.mockRawResponse))
					if err != nil {
						t.Fatalf("Failed to write response: %v", err)
					}
				} else {
					if err := json.NewEncoder(w).Encode(testCase.mockResponse); err != nil {
						t.Fatalf("Failed to encode response: %v", err)
					}
				}

			}))
			defer server.Close()

			client, err := NewClient(t.Context(), testCase.apiKey, WithCustomClient(&http.Client{}), WithCustomURLOauth(server.URL))

			if testCase.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedError.Error())
			} else {
				require.NoError(t, err)
				require.NotNil(t, client)
				defer client.Close()
				assert.Equal(t, testCase.expectedToken, client.accessToken)
			}
		})
	}
}

func TestClient_isValid(t *testing.T) {

	c := &Client{}

	testNow := time.Date(2023, 10, 27, 10, 0, 0, 0, time.UTC)
	testNowMs := testNow.UnixNano() / int64(time.Millisecond)

	fifteenMinutesMs := int64(tokenRefreshBuffer / time.Millisecond)

	testCases := []struct {
		name      string
		expiresAt int64
		expected  bool
	}{
		{
			name:      "Token is valid (expires in 1 hour)",
			expiresAt: testNowMs + (60 * 60 * 1000),
			expected:  true,
		},
		{
			name:      "Token is almost expired (expires in 16 minutes)",
			expiresAt: testNowMs + (16 * 60 * 1000),
			expected:  true,
		},
		{
			name:      "Token is at the boundary (expires in exactly 15 minutes)",
			expiresAt: testNowMs + fifteenMinutesMs,
			expected:  false,
		},
		{
			name:      "Token needs refresh (expires in 14 minutes)",
			expiresAt: testNowMs + (14 * 60 * 1000),
			expected:  false,
		},
		{
			name:      "Token is expired (expired 1 hour ago)",
			expiresAt: testNowMs - (60 * 60 * 1000),
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {

			actual := c.isValid(tc.expiresAt, testNow)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestClient_Refresh(t *testing.T) {
	testCases := []struct {
		name            string
		apiKey          string
		response        interface{}
		mockRawResponse string
		mockStatusCode  int
		expectedOutput  *tokenResponse
		expectedError   error
	}{
		{
			name:           "success",
			apiKey:         "fakeKey",
			mockStatusCode: http.StatusOK,
			expectedError:  nil,
			response: &tokenResponse{
				AccessToken: "token",
				ExpiresAt:   13132454545,
			},
			expectedOutput: &tokenResponse{
				AccessToken: "token",
				ExpiresAt:   13132454545,
			},
		},
		{
			name:            "Failure_InvalidJSONResponse",
			apiKey:          "fakeKey",
			mockRawResponse: `error, not json`,
			mockStatusCode:  http.StatusOK,
			expectedError:   errors.New("invalid character 'e' looking for beginning of value"),
			response:        nil,
		},
		{
			name:           "Failure_ServerError",
			apiKey:         "testKey",
			mockStatusCode: http.StatusInternalServerError,
			response:       nil,
			expectedError:  errors.New("oauth request failed with status 500"),
		},
		{
			name:           "Failure_unathorized",
			apiKey:         "testKey",
			mockStatusCode: http.StatusUnauthorized,
			response:       nil,
			expectedError:  errors.New("oauth request failed with status 401"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(testCase.mockStatusCode)
				if testCase.mockRawResponse != "" {
					w.Write([]byte(testCase.mockRawResponse))
				} else {
					if err := json.NewEncoder(w).Encode(testCase.response); err != nil {
						t.Fatalf("Failed to encode response: %v", err)
					}
				}
			}))

			client := &Client{
				httpClient:   &http.Client{},
				baseURLOauth: server.URL,
				accessToken: &tokenResponse{
					AccessToken: "token",
					ExpiresAt:   13132454545,
				},
			}

			err := client.refreshToken(t.Context())
			if testCase.expectedError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.expectedError.Error())
			} else {
				require.NotNil(t, client)
				require.NoError(t, err)
				assert.Equal(t, testCase.expectedOutput, client.accessToken)
			}
		})
	}
}
