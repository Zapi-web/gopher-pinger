package service_test

import (
	"context"
	"testing"

	"github.com/Zapi-web/gopher-pinger/internal/domain"
	"github.com/Zapi-web/gopher-pinger/internal/service"
	"github.com/Zapi-web/gopher-pinger/internal/service/mocks"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestPingerService_GetProcess(t *testing.T) {
	testId := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	expTar := domain.Target{
		ID:  testId,
		URL: "http://example.com",
	}

	tests := []struct {
		name           string
		id             string
		expectedTarget domain.Target
		mockBehavior   func(m *mocks.StateStore)
		serviceErr     error
	}{
		{
			name:           "Success",
			id:             testId,
			expectedTarget: expTar,
			mockBehavior: func(m *mocks.StateStore) {
				m.On("Get", mock.Anything, testId).Return(expTar, nil)
			},
			serviceErr: nil,
		},
		{
			name:           "Invalid id",
			id:             "ggg",
			expectedTarget: domain.Target{},
			mockBehavior:   func(m *mocks.StateStore) {},
			serviceErr:     domain.ErrInvalidId,
		},
		{
			name:           "Not found",
			id:             testId,
			expectedTarget: domain.Target{},
			mockBehavior: func(m *mocks.StateStore) {
				m.On("Get", mock.Anything, testId).Return(domain.Target{}, domain.ErrNotFound)
			},
			serviceErr: domain.ErrNotFound,
		},
		{
			name:           "Empty input",
			id:             "",
			expectedTarget: domain.Target{},
			mockBehavior:   func(m *mocks.StateStore) {},
			serviceErr:     domain.ErrInputisEmpty,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockState := new(mocks.StateStore)

			tt.mockBehavior(mockState)

			srv := service.NewService(context.Background(), nil, mockState, nil)

			res, err := srv.GetProcess(context.Background(), tt.id)

			assert.Equal(t, tt.expectedTarget, res)
			if tt.serviceErr != nil {
				assert.ErrorIs(t, err, tt.serviceErr)
			} else {
				assert.NoError(t, err)
			}

			mockState.AssertExpectations(t)
		})
	}
}

func TestPingerService_UpdateProcess(t *testing.T) {
	testId := "01ARZ3NDEKTSV4RRFFQ69G5FAV"
	parsedULID, _ := ulid.Parse(testId)

	oldTar := domain.Target{
		ID:       testId,
		URL:      "http://example.com",
		Interval: 10,
	}
	newTar := domain.Target{
		ID:       testId,
		URL:      "http://example.com",
		Interval: 20,
	}

	type serviceMocks struct {
		state     *mocks.StateStore
		processes *mocks.ProcessStore
		metrics   *mocks.PingerMetrics
	}

	tests := []struct {
		name         string
		id           string
		interval     int
		serviceError error
		mockBehavior func(m serviceMocks)
	}{
		{
			name:         "Success",
			id:           testId,
			interval:     20,
			serviceError: nil,
			mockBehavior: func(m serviceMocks) {
				m.state.On("Get", mock.Anything, testId).Return(oldTar, nil)

				m.state.On("Delete", mock.Anything, testId).Return(nil)
				m.processes.On("Get", parsedULID).Return(context.CancelFunc(func() {}), nil)
				m.processes.On("Delete", parsedULID).Return(nil)
				m.metrics.On("DecWorker").Return()

				m.state.On("Lock", mock.Anything, testId, mock.Anything).Return(true, nil).Maybe()
				m.state.On("Unlock", mock.Anything, testId).Return(nil).Maybe()
				m.state.On("UpdateStatus", mock.Anything, testId, mock.Anything, mock.Anything).Return(nil).Maybe()

				m.processes.On("Set", parsedULID, mock.AnythingOfType("context.CancelFunc")).Return(nil)
				m.state.On("Set", mock.Anything, testId, newTar).Return(nil)
				m.metrics.On("IncWorker").Return()
			},
		},
		{
			name:         "failed getting old data (not found)",
			id:           testId,
			interval:     20,
			serviceError: domain.ErrNotFound,
			mockBehavior: func(m serviceMocks) {
				m.state.On("Get", mock.Anything, testId).Return(domain.Target{}, domain.ErrNotFound)
			},
		},
		{
			name:         "Invalid Interval",
			id:           testId,
			interval:     -5,
			serviceError: domain.ErrInvalidInterval,
			mockBehavior: func(m serviceMocks) {},
		},
		{
			name:         "Input is empty",
			id:           "",
			interval:     10,
			serviceError: domain.ErrInputisEmpty,
			mockBehavior: func(m serviceMocks) {},
		}, {
			name:         "Invalid id",
			id:           "id",
			interval:     10,
			serviceError: domain.ErrInvalidId,
			mockBehavior: func(m serviceMocks) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := serviceMocks{
				state:     new(mocks.StateStore),
				processes: new(mocks.ProcessStore),
				metrics:   new(mocks.PingerMetrics),
			}

			tt.mockBehavior(m)

			srv := service.NewService(context.Background(), m.processes, m.state, m.metrics)

			err := srv.UpdateProcess(context.Background(), tt.id, tt.interval)

			if tt.serviceError != nil {
				assert.ErrorIs(t, err, tt.serviceError)
			} else {
				assert.NoError(t, err)
			}

			m.state.AssertExpectations(t)
			m.processes.AssertExpectations(t)
			m.metrics.AssertExpectations(t)
		})
	}
}

func TestPingerService_StartMonitoring(t *testing.T) {
	type serviceMocks struct {
		state     *mocks.StateStore
		processes *mocks.ProcessStore
		metrics   *mocks.PingerMetrics
	}

	tests := []struct {
		name         string
		interval     int
		url          string
		serviceError error
		mockBehavior func(m serviceMocks)
	}{
		{
			name:         "Success",
			interval:     10,
			url:          "http://example.com",
			serviceError: nil,
			mockBehavior: func(m serviceMocks) {
				m.processes.On("Set", mock.Anything, mock.Anything).Return(nil)
				m.state.On("Set", mock.Anything, mock.Anything, mock.Anything).Return(nil)
				m.metrics.On("IncWorker").Return()
			},
		},
		{
			name:         "Invalid interval",
			interval:     -5,
			url:          "http://example.com",
			serviceError: domain.ErrInvalidInterval,
			mockBehavior: func(m serviceMocks) {},
		},
		{
			name:         "invalid URL",
			interval:     10,
			url:          "not_a_url",
			serviceError: domain.ErrInvalidURL,
			mockBehavior: func(m serviceMocks) {},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := serviceMocks{
				state:     new(mocks.StateStore),
				processes: new(mocks.ProcessStore),
				metrics:   new(mocks.PingerMetrics),
			}

			tt.mockBehavior(m)

			srv := service.NewService(context.Background(), m.processes, m.state, m.metrics)

			_, err := srv.StartMonitoring(context.Background(), tt.url, tt.interval)

			if tt.serviceError != nil {
				assert.ErrorIs(t, err, tt.serviceError)
			} else {
				assert.NoError(t, err)
			}

			m.state.AssertExpectations(t)
			m.processes.AssertExpectations(t)
			m.metrics.AssertExpectations(t)
		})
	}
}
