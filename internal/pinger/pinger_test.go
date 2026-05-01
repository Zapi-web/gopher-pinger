package pinger_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Zapi-web/gopher-pinger/internal/pinger"
	"github.com/Zapi-web/gopher-pinger/internal/pinger/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPinger_Logic(t *testing.T) {
	testId := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	testInterval := 10 * time.Millisecond

	tests := []struct {
		name           string
		interval       time.Duration
		expectedStatus int
		serviceError   error
		overrideURL    string
		serverBehavior func(http.ResponseWriter, *http.Request)
		isLocked       bool
	}{
		{
			name:           "Success",
			interval:       testInterval,
			expectedStatus: http.StatusOK,
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			},
			isLocked: false,
		},
		{
			name:           "Not valid url",
			interval:       testInterval,
			expectedStatus: -1,
			overrideURL:    "http://127.0.0.1:0",
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {},
			isLocked:       false,
		},
		{
			name:           "Url is locked",
			interval:       testInterval,
			serverBehavior: func(w http.ResponseWriter, r *http.Request) {},
			isLocked:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLocker := new(mocks.Locker)
			mockLocker.On("Lock", mock.Anything, testId, mock.Anything).Return(!tt.isLocked, nil)
			mockChan := make(chan pinger.CheckResult, 1)

			server := httptest.NewServer(http.HandlerFunc(tt.serverBehavior))
			defer server.Close()

			targetURL := server.URL
			if tt.overrideURL != "" {
				targetURL = tt.overrideURL
			}

			cancel := pinger.Start(context.Background(), mockLocker, testId, targetURL, tt.interval, mockChan)

			select {
			case res := <-mockChan:
				if tt.isLocked {
					t.Error("URL is locked, pinger is not supposed to ping")
				}

				assert.Equal(t, tt.expectedStatus, res.Status)
			case <-time.After(100 * time.Millisecond):
				if !tt.isLocked {
					t.Errorf("time-out: result has not been achieved")
				}
			}

			cancel()
			mockLocker.AssertExpectations(t)
		})
	}
}
