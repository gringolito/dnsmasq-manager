package handler

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/gringolito/dnsmasq-manager/api"
	"github.com/gringolito/dnsmasq-manager/api/presenter"
	"github.com/gringolito/dnsmasq-manager/api/scope"
	"github.com/gringolito/dnsmasq-manager/config"
	"github.com/gringolito/dnsmasq-manager/pkg/host"
	hostmock "github.com/gringolito/dnsmasq-manager/pkg/host/mock"
	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"github.com/gringolito/dnsmasq-manager/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	InvalidMACAddress     = "ab:cd:ef:gh:ij:kl"
	ValidMACAddress       = "aa:bb:cc:dd:ee:ff"
	InvalidIPAddress      = "1111"
	ValidIPAddress        = "1.1.1.1"
	InvalidHostName       = "B@r"
	ValidHostJSON         = `{"HostName":"Foo", "IPAddress":"1.1.1.1", "MacAddress":"aa:bb:cc:dd:ee:ff"}`
	InvalidJSON           = `"HostName":"Foo", "IPAddress":"1.1.1.1", "MacAddress":"aa:bb:cc:dd:ee:ff"`
	MissingMACAddressJSON = `{"HostName":"Foo", "IPAddress":"1.1.1.1"}`
	MissingIPAddressJSON  = `{"HostName":"Foo", "MacAddress":"aa:bb:cc:dd:ee:ff"}`
	MissingHostNameJSON   = `{"IPAddress":"1.1.1.1", "MacAddress":"aa:bb:cc:dd:ee:ff"}`
	InvalidMACAddressJSON = `{"HostName":"Foo", "IPAddress":"1.1.1.1", "MacAddress":"ab:cd:ef:gh:ij:kl"}`
	InvalidIPAddressJSON  = `{"HostName":"Foo", "IPAddress":"1111", "MacAddress":"aa:bb:cc:dd:ee:ff"}`
	InvalidHostNameJSON   = `{"HostName":"B@r", "IPAddress":"1.1.1.1", "MacAddress":"aa:bb:cc:dd:ee:ff"}`
	AllHostsJSON          = `[
		{
			"MacAddress":"02:04:06:aa:bb:cc",
			"IPAddress":"1.1.1.1",
			"HostName":"Foo"
		},
		{
			"MacAddress":"02:04:06:dd:ee:ff",
			"IPAddress":"1.1.1.2",
			"HostName":"Bar"
		}
	]`
)

var ValidHost = model.StaticDhcpHost{MacAddress: tests.ParseMAC(ValidMACAddress), IPAddress: net.ParseIP(ValidIPAddress), HostName: "Foo"}
var AllHosts = []model.StaticDhcpHost{
	{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
	{MacAddress: tests.ParseMAC("02:04:06:dd:ee:ff"), IPAddress: net.ParseIP("1.1.1.2"), HostName: "Bar"},
}

var voidMock = func(mock *hostmock.ServiceMock) {}

func setupTest(t *testing.T, mockSetup func(mock *hostmock.ServiceMock)) *fiber.App {
	app := tests.SetupApp()
	config := tests.SetupConfig(t)
	serviceMock := &hostmock.ServiceMock{}
	router := tests.SetupRouter(app, config)
	RouteStaticHosts(router, serviceMock)
	mockSetup(serviceMock)
	return app
}

type jwtTokenConfig struct {
	SigningKey string
	Claims     jwt.MapClaims
}

func createJwtToken(t *testing.T, config *jwtTokenConfig) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, config.Claims)

	tokenString, err := token.SignedString([]byte(config.SigningKey))
	require.NoError(t, err, "Failed to sign JWT token")

	return tokenString
}

func TestStaticHostsApi(t *testing.T) {
	var testCases = []struct {
		name               string
		httpMethod         string
		route              string
		requestBody        io.Reader
		expectedStatusCode int
		expectedResponse   string
		mockSetup          func(s *hostmock.ServiceMock)
	}{
		{
			name:               "GetAllStaticHostsSuccess",
			httpMethod:         http.MethodGet,
			route:              "/api/v1/static/hosts",
			expectedStatusCode: http.StatusOK,
			expectedResponse:   AllHostsJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchAll").Once().Return(&AllHosts, nil)
			},
		},
		{
			name:               "GetAllStaticHostsServiceError",
			httpMethod:         http.MethodGet,
			route:              "/api/v1/static/hosts",
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchAll").Once().Return(nil, errors.New("an error"))
			},
		},
		{
			name:               "GetStaticHostNoQueryParameter",
			httpMethod:         http.MethodGet,
			route:              "/api/v1/static/host",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   tests.ErrorJSON(http.StatusBadRequest, InvalidRequestMessage, MissingQueryParameter),
			mockSetup:          voidMock,
		},
		{
			name:               "GetStaticHostByMACSuccess",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusOK,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(&ValidHost, nil)
			},
		},
		{
			name:               "GetStaticHostInvalidMACAddress",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", InvalidMACAddress),
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   tests.ErrorJSON(http.StatusBadRequest, InvalidMacAddressMessage, fmt.Sprintf(MalformedMacAddress, InvalidMACAddress)),
			mockSetup:          voidMock,
		},
		{
			name:               "GetStaticHostByMACNotFound",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   tests.ErrorJSON(http.StatusNotFound, StaticHostNotFoundMessage, fmt.Sprintf(NoMatchingMacAddress, ValidMACAddress)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(nil, nil)
			},
		},
		{
			name:               "GetStaticHostByMACServiceError",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(nil, errors.New("an error"))
			},
		},
		{
			name:               "GetStaticHostByIPSuccess",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?ip=%s", ValidIPAddress),
			expectedStatusCode: http.StatusOK,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByIP", net.ParseIP(ValidIPAddress)).Once().Return(&ValidHost, nil)
			},
		},
		{
			name:               "GetStaticHostByIPNotFound",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?ip=%s", ValidIPAddress),
			expectedStatusCode: http.StatusNotFound,
			expectedResponse:   tests.ErrorJSON(http.StatusNotFound, StaticHostNotFoundMessage, fmt.Sprintf(NoMatchingIPAddress, ValidIPAddress)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByIP", net.ParseIP(ValidIPAddress)).Once().Return(nil, nil)
			},
		},
		{
			name:               "GetStaticHostByIPServiceError",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?ip=%s", ValidIPAddress),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByIP", net.ParseIP(ValidIPAddress)).Once().Return(nil, errors.New("an error"))
			},
		},
		{
			name:               "PostStaticHostSuccess",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Insert", &ValidHost).Once().Return(nil)
			},
		},
		{
			name:               "PostStaticHostInvalidJSON",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ErrorJSON(http.StatusUnprocessableEntity, InvalidRequestBodyMessage, HostCouldNotBeParsed),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostMissingMACAddress",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(MissingMACAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "MacAddress", "The MacAddress field is required.", ""),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostMissingIPAddress",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(MissingIPAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "IPAddress", "The IPAddress field is required.", ""),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostMissingHostName",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(MissingHostNameJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "HostName", "The HostName field is required.", ""),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostInvalidMACAddress",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidMACAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "MacAddress", "The MacAddress field must be of type mac.", InvalidMACAddress),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostInvalidIPAddress",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidIPAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "IPAddress", "The IPAddress field must be of type ipv4.", InvalidIPAddress),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostInvalidHostName",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidHostNameJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "HostName", "The HostName field must be of type hostname.", InvalidHostName),
			mockSetup:          voidMock,
		},
		{
			name:               "PostStaticHostDuplicatedIPAddress",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusConflict,
			expectedResponse:   tests.ErrorJSON(http.StatusConflict, DuplicatedIPAddressMessage, fmt.Sprintf(IPAddressAlreadyInUse, ValidIPAddress)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Insert", &ValidHost).Once().Return(host.DuplicatedEntryError{Field: "IP", Value: ValidIPAddress})
			},
		},
		{
			name:               "PostStaticHostDuplicatedMACAddress",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusConflict,
			expectedResponse:   tests.ErrorJSON(http.StatusConflict, DuplicatedMacAddressMessage, fmt.Sprintf(MacAddressAlreadyInUse, ValidMACAddress)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Insert", &ValidHost).Once().Return(host.DuplicatedEntryError{Field: "MAC", Value: ValidMACAddress})
			},
		},
		{
			name:               "PostStaticHostServiceError",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Insert", &ValidHost).Once().Return(errors.New("an error"))
			},
		},
		{
			name:               "PutStaticHostSuccess",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Update", &ValidHost).Once().Return(nil)
			},
		},
		{
			name:               "PutStaticHostInvalidJSON",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ErrorJSON(http.StatusUnprocessableEntity, InvalidRequestBodyMessage, HostCouldNotBeParsed),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostMissingMACAddress",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(MissingMACAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "MacAddress", "The MacAddress field is required.", ""),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostMissingIPAddress",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(MissingIPAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "IPAddress", "The IPAddress field is required.", ""),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostMissingHostName",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(MissingHostNameJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "HostName", "The HostName field is required.", ""),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostInvalidMACAddress",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidMACAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "MacAddress", "The MacAddress field must be of type mac.", InvalidMACAddress),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostInvalidIPAddress",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidIPAddressJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "IPAddress", "The IPAddress field must be of type ipv4.", InvalidIPAddress),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostInvalidHostName",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(InvalidHostNameJSON),
			expectedStatusCode: http.StatusUnprocessableEntity,
			expectedResponse:   tests.ValidationErrorJSON(InvalidRequestBodyMessage, "HostName", "The HostName field must be of type hostname.", InvalidHostName),
			mockSetup:          voidMock,
		},
		{
			name:               "PutStaticHostServiceError",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Update", &ValidHost).Once().Return(errors.New("an error"))
			},
		},
		{
			name:               "DeleteStaticHostNoQueryParameter",
			httpMethod:         http.MethodDelete,
			route:              "/api/v1/static/host",
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   tests.ErrorJSON(http.StatusBadRequest, InvalidRequestMessage, MissingQueryParameter),
			mockSetup:          voidMock,
		},
		{
			name:               "DeleteStaticHostByMACSuccess",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusOK,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(&ValidHost, nil)
			},
		},
		{
			name:               "DeleteStaticHostInvalidMACAddress",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", InvalidMACAddress),
			expectedStatusCode: http.StatusBadRequest,
			expectedResponse:   tests.ErrorJSON(http.StatusBadRequest, InvalidMacAddressMessage, fmt.Sprintf(MalformedMacAddress, InvalidMACAddress)),
			mockSetup:          voidMock,
		},
		{
			name:               "DeleteStaticHostByMACNotFound",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusNoContent,
			expectedResponse:   "",
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(nil, nil)
			},
		},
		{
			name:               "DeleteStaticHostByMACServiceError",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(nil, errors.New("an error"))
			},
		},
		{
			name:               "DeleteStaticHostByIPSuccess",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?ip=%s", ValidIPAddress),
			expectedStatusCode: http.StatusOK,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByIP", net.ParseIP(ValidIPAddress)).Once().Return(&ValidHost, nil)
			},
		},
		{
			name:               "DeleteStaticHostByIPNotFound",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?ip=%s", ValidIPAddress),
			expectedStatusCode: http.StatusNoContent,
			expectedResponse:   "",
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByIP", net.ParseIP(ValidIPAddress)).Once().Return(nil, nil)
			},
		},
		{
			name:               "DeleteStaticHostByIPServiceError",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?ip=%s", ValidIPAddress),
			expectedStatusCode: http.StatusInternalServerError,
			expectedResponse:   tests.ErrorJSON(http.StatusInternalServerError, presenter.ServerErrorMessage, fmt.Sprintf(presenter.InternalServerError, tests.UUIDRegexMatch)),
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByIP", net.ParseIP(ValidIPAddress)).Once().Return(nil, errors.New("an error"))
			},
		},
	}

	for _, test := range testCases {
		description := fmt.Sprintf("%s %s %d", test.httpMethod, test.route, test.expectedStatusCode)
		t.Run(test.name, func(t *testing.T) {
			app := setupTest(t, test.mockSetup)

			request := httptest.NewRequest(test.httpMethod, test.route, test.requestBody)
			request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			response, err := app.Test(request)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatusCode, response.StatusCode, "%s: returned wrong HTTP status code", description)

			responseBody := tests.GetBody(response)
			if !tests.JSONMatches(test.expectedResponse, string(responseBody)) {
				assert.JSONEq(t, test.expectedResponse, string(responseBody), "%s: unexpected HTTP response body", description)
			}
		})
	}
}

func TestStaticHostsApiWithAuth(t *testing.T) {
	const (
		AuthMethod = config.AuthHS256
		AuthKey    = "FooBar77"
	)

	var testCases = []struct {
		name               string
		httpMethod         string
		route              string
		token              *jwtTokenConfig
		requestBody        io.Reader
		expectedStatusCode int
		expectedResponse   string
		mockSetup          func(s *hostmock.ServiceMock)
	}{
		{
			name:               "GetAllNoAuthentication",
			httpMethod:         http.MethodGet,
			route:              "/api/v1/static/hosts",
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.MissingOrMalformedJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "GetAllInvalidKey",
			httpMethod: http.MethodGet,
			route:      "/api/v1/static/hosts",
			token: &jwtTokenConfig{
				SigningKey: "InvalidKey",
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.InvalidOrExpiredJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "GetAllMissingScopeClaimJWT",
			httpMethod: http.MethodGet,
			route:      "/api/v1/static/hosts",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MalformedJwt),
			mockSetup:          voidMock,
		},
		{
			name:       "GetAllAuthorized",
			httpMethod: http.MethodGet,
			route:      "/api/v1/static/hosts",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpRead,
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   AllHostsJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchAll").Once().Return(&AllHosts, nil)
			},
		},
		{
			name:               "GetNoAuthentication",
			httpMethod:         http.MethodGet,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.MissingOrMalformedJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "GetInvalidKey",
			httpMethod: http.MethodGet,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: "InvalidKey",
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.InvalidOrExpiredJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "GetMissingScopeClaimJWT",
			httpMethod: http.MethodGet,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MalformedJwt),
			mockSetup:          voidMock,
		},
		{
			name:       "GetAuthorized",
			httpMethod: http.MethodGet,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpRead,
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("FetchByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(&ValidHost, nil)
			},
		},
		{
			name:               "PostNoAuthentication",
			httpMethod:         http.MethodPost,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.MissingOrMalformedJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "PostInvalidKey",
			httpMethod: http.MethodPost,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: "InvalidKey",
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.InvalidOrExpiredJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "PostMissingScopeClaimJWT",
			httpMethod: http.MethodPost,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MalformedJwt),
			mockSetup:          voidMock,
		},
		{
			name:       "PostUnauthorized",
			httpMethod: http.MethodPost,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpRead,
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MissingRole),
			mockSetup:          voidMock,
		},
		{
			name:       "PostAuthorized",
			httpMethod: http.MethodPost,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpWrite,
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Insert", &ValidHost).Once().Return(nil)
			},
		},
		{
			name:               "PutNoAuthentication",
			httpMethod:         http.MethodPut,
			route:              "/api/v1/static/host",
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.MissingOrMalformedJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "PutInvalidKey",
			httpMethod: http.MethodPut,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: "InvalidKey",
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.InvalidOrExpiredJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "PutMissingScopeClaimJWT",
			httpMethod: http.MethodPut,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MalformedJwt),
			mockSetup:          voidMock,
		},
		{
			name:       "PutUnauthorized",
			httpMethod: http.MethodPut,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpWrite,
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MissingRole),
			mockSetup:          voidMock,
		},
		{
			name:       "PutAuthorized",
			httpMethod: http.MethodPut,
			route:      "/api/v1/static/host",
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpAdmin,
				},
			},
			requestBody:        strings.NewReader(ValidHostJSON),
			expectedStatusCode: http.StatusCreated,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("Update", &ValidHost).Once().Return(nil)
			},
		},
		{
			name:               "DeleteNoAuthentication",
			httpMethod:         http.MethodDelete,
			route:              fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.MissingOrMalformedJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "DeleteInvalidKey",
			httpMethod: http.MethodDelete,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: "InvalidKey",
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			expectedStatusCode: http.StatusUnauthorized,
			expectedResponse:   tests.ErrorJSON(http.StatusUnauthorized, api.UnauthorizedMessage, api.InvalidOrExpiredJWT),
			mockSetup:          voidMock,
		},
		{
			name:       "DeleteMissingScopeClaimJWT",
			httpMethod: http.MethodDelete,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name": "unit-tests",
					"exp":  time.Now().Add(time.Minute * 30).Unix(),
				},
			},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MalformedJwt),
			mockSetup:          voidMock,
		},
		{
			name:       "DeleteUnauthorized",
			httpMethod: http.MethodDelete,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpWrite,
				},
			},
			expectedStatusCode: http.StatusForbidden,
			expectedResponse:   tests.ErrorJSON(http.StatusForbidden, api.NotAuthorizedMessage, api.MissingRole),
			mockSetup:          voidMock,
		},
		{
			name:       "DeleteAuthorized",
			httpMethod: http.MethodDelete,
			route:      fmt.Sprintf("/api/v1/static/host?mac=%s", ValidMACAddress),
			token: &jwtTokenConfig{
				SigningKey: AuthKey,
				Claims: jwt.MapClaims{
					"name":  "unit-tests",
					"exp":   time.Now().Add(time.Minute * 30).Unix(),
					"scope": scope.DhcpAdmin,
				},
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse:   ValidHostJSON,
			mockSetup: func(mock *hostmock.ServiceMock) {
				mock.On("RemoveByMac", tests.ParseMAC(ValidMACAddress)).Once().Return(&ValidHost, nil)
			},
		},
	}

	os.Setenv("DMM_AUTH_METHOD", AuthMethod)
	os.Setenv("DMM_AUTH_KEY", AuthKey)

	for _, test := range testCases {
		description := fmt.Sprintf("%s %s %d", test.httpMethod, test.route, test.expectedStatusCode)

		t.Run(test.name, func(t *testing.T) {
			app := setupTest(t, test.mockSetup)

			request := httptest.NewRequest(test.httpMethod, test.route, test.requestBody)
			request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)

			if test.token != nil {
				jwtToken := createJwtToken(t, test.token)
				request.Header.Set(fiber.HeaderAuthorization, "Bearer "+jwtToken)
			}

			response, err := app.Test(request)
			require.NoError(t, err, "app.Test() request failed")

			assert.Equal(t, test.expectedStatusCode, response.StatusCode, "%s: returned wrong HTTP status code", description)

			responseBody := tests.GetBody(response)
			if !tests.JSONMatches(test.expectedResponse, string(responseBody)) {
				assert.JSONEq(t, test.expectedResponse, string(responseBody), "%s: unexpected HTTP response body", description)
			}
		})

	}

	os.Unsetenv("DMM_AUTH_METHOD")
	os.Unsetenv("DMM_AUTH_KEY")
}
