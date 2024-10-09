package host

import (
	"errors"
	"fmt"
	"net"
	"testing"

	hostmock "github.com/gringolito/dnsmasq-manager/pkg/host/mock"
	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"github.com/gringolito/dnsmasq-manager/tests"
	"github.com/stretchr/testify/assert"
)

const (
	ValidMACAddress = "02:04:06:aa:bb:cc"
	ValidIPAddress  = "1.1.1.1"
)

var ValidHost = model.StaticDhcpHost{MacAddress: tests.ParseMAC(ValidMACAddress), IPAddress: net.ParseIP(ValidIPAddress), HostName: "Foo"}

func TestHostServiceInsertUpdate(t *testing.T) {
	Insert := func(service Service) error { return service.Insert(&ValidHost) }
	Update := func(service Service) error { return service.Update(&ValidHost) }

	var testCases = []struct {
		name   string
		method func(service Service) error
		on     func(mock *hostmock.RepositoryMock)
		assert func(t *testing.T, err error, mock *hostmock.RepositoryMock)
	}{
		{
			name:   "InsertSuccess",
			method: Insert,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(nil, nil)
				mock.On("Save", &ValidHost).Once().Return(nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "unexpected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "InsertDuplicatedMac",
			method: Insert,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(&ValidHost, nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				assert.Equal(t, &DuplicatedEntryError{Field: "MAC", Value: ValidHost.MacAddress.String()}, err, "error mismatch")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "InsertDuplicatedIP",
			method: Insert,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(&ValidHost, nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				assert.Equal(t, &DuplicatedEntryError{Field: "IP", Value: ValidHost.IPAddress.String()}, err, "error mismatch")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "InsertSaveError",
			method: Insert,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(nil, nil)
				mock.On("Save", &ValidHost).Once().Return(errors.New("an error"))
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "InsertFindByIPError",
			method: Insert,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "InsertFindByMacError",
			method: Insert,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateNewHost",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(nil, nil)
				mock.On("Save", &ValidHost).Once().Return(nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "unexpected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateSameMac",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(&ValidHost, nil)
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(nil, nil)
				mock.On("Save", &ValidHost).Once().Return(nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "unexpected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateSameIP",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(&ValidHost, nil)
				mock.On("Save", &ValidHost).Once().Return(nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "unexpected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateSameHost",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(&ValidHost, nil)
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(&ValidHost, nil)
				mock.On("Save", &ValidHost).Once().Return(nil)
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "unexpected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateSaveError",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(nil, nil)
				mock.On("Save", &ValidHost).Once().Return(errors.New("an error"))
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateDeleteByIPError",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, nil)
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "UpdateDeleteByMacError",
			method: Update,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "expected error not found")
				mock.AssertExpectations(t)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			repositoryMock := &hostmock.RepositoryMock{}
			test.on(repositoryMock)

			service := NewService(repositoryMock)
			err := test.method(service)
			test.assert(t, err, repositoryMock)
		})
	}
}

func TestHostServiceFetchAll(t *testing.T) {
	allHosts := []model.StaticDhcpHost{
		{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
		{MacAddress: tests.ParseMAC("02:04:06:dd:ee:ff"), IPAddress: net.ParseIP("2.2.2.2"), HostName: "Bar"},
	}

	var testCases = []struct {
		name   string
		on     func(mock *hostmock.RepositoryMock)
		assert func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock)
	}{
		{
			name: "Success",
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindAll").Once().Return(&allHosts, nil)
			},
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "FetchAll() returned an unexpected error")
				assert.NotNil(t, hosts, "FetchAll() returned unexpected nil hosts")
				assert.Equal(t, &allHosts, hosts, "FetchAll() returned unexpected hosts")
				mock.AssertExpectations(t)
			},
		},
		{
			name: "EmptyHosts",
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindAll").Once().Return(&[]model.StaticDhcpHost{}, nil)
			},
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "FetchAll() returned an unexpected error")
				assert.NotNil(t, hosts, "FetchAll() returned unexpected nil hosts")
				assert.Equal(t, &[]model.StaticDhcpHost{}, hosts, "FetchAll() returned unexpected hosts")
				mock.AssertExpectations(t)
			},
		},
		{
			name: "RepositoryError",
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindAll").Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "FetchAll() did NOT returned an expected error")
				mock.AssertExpectations(t)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			repositoryMock := &hostmock.RepositoryMock{}
			test.on(repositoryMock)

			service := NewService(repositoryMock)
			hosts, err := service.FetchAll()
			test.assert(t, hosts, err, repositoryMock)
		})
	}
}

func TestHostServiceFetchRemove(t *testing.T) {
	FetchByMac := func(service Service) (*model.StaticDhcpHost, error) { return service.FetchByMac(ValidHost.MacAddress) }
	FetchByIP := func(service Service) (*model.StaticDhcpHost, error) { return service.FetchByIP(ValidHost.IPAddress) }
	RemoveByMac := func(service Service) (*model.StaticDhcpHost, error) { return service.RemoveByMac(ValidHost.MacAddress) }
	RemoveByIP := func(service Service) (*model.StaticDhcpHost, error) { return service.RemoveByIP(ValidHost.IPAddress) }

	var testCases = []struct {
		name   string
		method func(service Service) (*model.StaticDhcpHost, error)
		on     func(mock *hostmock.RepositoryMock)
		assert func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock)
	}{
		{
			name:   "FetchByMacFound",
			method: FetchByMac,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(&ValidHost, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "FetchByMac() returned an unexpected error")
				assert.NotNil(t, host, "FetchByMac() returned an unexpected nil host")
				assert.Equal(t, &ValidHost, host, "FetchByMac() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "FetchByMacNotFound",
			method: FetchByMac,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "FetchByMac() returned an unexpected error")
				assert.Nil(t, host, "FetchByMac() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "FetchByMacError",
			method: FetchByMac,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByMac", ValidHost.MacAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "FetchByMac() did NOT returned an expected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "FetchByIPFound",
			method: FetchByIP,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(&ValidHost, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "FetchByIP() returned an unexpected error")
				assert.NotNil(t, host, "FetchByIP() returned an unexpected nil host")
				assert.Equal(t, &ValidHost, host, "FetchByIP() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "FetchByIPNotFound",
			method: FetchByIP,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(nil, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "FetchByIP() returned an unexpected error")
				assert.Nil(t, host, "FetchByIP() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "FetchByIPError",
			method: FetchByIP,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("FindByIP", ValidHost.IPAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "FetchByIP() did NOT returned an expected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "RemoveByMacFound",
			method: RemoveByMac,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(&ValidHost, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "RemoveByMac() returned an unexpected error")
				assert.NotNil(t, host, "RemoveByMac() returned an unexpected nil host")
				assert.Equal(t, &ValidHost, host, "RemoveByMac() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "RemoveByMacNotFound",
			method: RemoveByMac,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "RemoveByMac() returned an unexpected error")
				assert.Nil(t, host, "RemoveByMac() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "RemoveByMacError",
			method: RemoveByMac,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByMac", ValidHost.MacAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "RemoveByMac() did NOT returned an expected error")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "RemoveByIPFound",
			method: RemoveByIP,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(&ValidHost, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "RemoveByIP() returned an unexpected error")
				assert.NotNil(t, host, "RemoveByIP() returned an unexpected nil host")
				assert.Equal(t, &ValidHost, host, "RemoveByIP() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "RemoveByIPNotFound",
			method: RemoveByIP,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(nil, nil)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.NoError(t, err, "RemoveByIP() returned an unexpected error")
				assert.Nil(t, host, "RemoveByIP() returned an unexpected host")
				mock.AssertExpectations(t)
			},
		},
		{
			name:   "RemoveByIPError",
			method: RemoveByIP,
			on: func(mock *hostmock.RepositoryMock) {
				mock.On("DeleteByIP", ValidHost.IPAddress).Once().Return(nil, errors.New("an error"))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, mock *hostmock.RepositoryMock) {
				assert.Error(t, err, "RemoveByIP() did NOT returned an expected error")
				mock.AssertExpectations(t)
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			repositoryMock := &hostmock.RepositoryMock{}
			test.on(repositoryMock)

			service := NewService(repositoryMock)
			host, err := test.method(service)
			test.assert(t, host, err, repositoryMock)
		})
	}
}

func TestHostServiceErrors(t *testing.T) {

	var testCases = []struct {
		name            string
		field           string
		value           string
		expectedMessage string
	}{
		{
			name:  "DuplicatedIP",
			field: "IP",
			value: "1.1.1.1",
		},
		{
			name:  "DuplicatedMAC",
			field: "MAC",
			value: "aa:bb:cc:dd:ee:ff",
		},
	}

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			err := &DuplicatedEntryError{Field: test.field, Value: test.value}
			expectedMessage := fmt.Sprintf(duplicatedEntryErrorMessage, test.field, test.value)
			assert.ErrorContains(t, err, expectedMessage)
		})
	}
}
