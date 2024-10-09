package host

import (
	"errors"
	"net"
	"os"
	"testing"

	"github.com/gringolito/dnsmasq-manager/pkg/model"
	"github.com/gringolito/dnsmasq-manager/tests"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var AllHosts = []model.StaticDhcpHost{
	{MacAddress: tests.ParseMAC("02:04:06:aa:bb:cc"), IPAddress: net.ParseIP("1.1.1.1"), HostName: "Foo"},
	{MacAddress: tests.ParseMAC("02:04:06:dd:ee:ff"), IPAddress: net.ParseIP("1.1.1.2"), HostName: "Bar"},
	{MacAddress: tests.ParseMAC("02:04:06:12:34:56"), IPAddress: net.ParseIP("1.1.1.3"), HostName: "Baz"},
}

var UnknownHost = model.StaticDhcpHost{MacAddress: tests.ParseMAC("02:04:06:aa:bb:ff"), IPAddress: net.ParseIP("9.9.9.9"), HostName: "Unknown"}

const (
	AllHostsFileContent = `dhcp-host=02:04:06:dd:ee:ff,1.1.1.2,Bar
dhcp-host=02:04:06:aa:bb:cc,1.1.1.1,Foo
dhcp-host=02:04:06:12:34:56,1.1.1.3,Baz`
	DeletedValidHostFileContent = `dhcp-host=02:04:06:dd:ee:ff,1.1.1.2,Bar
dhcp-host=02:04:06:12:34:56,1.1.1.3,Baz`
	AddedUnknownHostFileContent = `dhcp-host=02:04:06:dd:ee:ff,1.1.1.2,Bar
dhcp-host=02:04:06:aa:bb:cc,1.1.1.1,Foo
dhcp-host=02:04:06:12:34:56,1.1.1.3,Baz
dhcp-host=02:04:06:aa:bb:ff,9.9.9.9,Unknown`
	ValidHostFileContent    = `dhcp-host=02:04:06:aa:bb:cc,1.1.1.1,Foo`
	InvalidHostsFileContent = `dhcp-host=ab:cd:ef:gh:ij:kl,1.1.1.1,Jung`
)

func setUpStaticHostsFile(t *testing.T, content string) string {
	file, err := os.CreateTemp("", "dmm-tests-dhcp-static-leases")
	require.NoError(t, err, "Failed to create DHCP static hosts file")
	defer file.Close()

	length, err := file.Write([]byte(content))
	require.NoError(t, err, "Failed to initialize DHCP static hosts file")
	require.Equal(t, len(content), length, "DHCP static hosts file, possible content mismatch")

	return file.Name()
}

func tearDownStaticHostsFile(t *testing.T, fileName string) {
	_, err := os.Stat(fileName)
	if !errors.Is(err, os.ErrNotExist) {
		os.Remove(fileName)
	}
}

func assertFileContent(t *testing.T, expectedFileContent string, fileName string) {
	actualFileData, err := os.ReadFile(fileName)
	actualFileContent := string(actualFileData)
	require.NoError(t, err, "Failed to open test file for validation")
	assert.Equal(t, expectedFileContent, actualFileContent, "DHCP static leases file doesn't match")
}

func TestHostRepositoryFindAll(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		setup               func(tc *testcase)
		assert              func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			setup:               voidSetup,
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "FindAll() returned an unexpected error")
				assert.NotNil(t, hosts, "FindAll() unexpectedly returned nil hosts")
				assert.ElementsMatch(t, AllHosts, *hosts, "FindAll() returned unexpected hosts")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "EmptyFile",
			setupFileContent:    "",
			expectedFileContent: "",
			setup:               voidSetup,
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "FindAll() returned an unexpected error")
				assert.Empty(t, hosts, "FindAll() returned unexpected hosts")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "FindAll() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "FindAll() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			setup:            voidSetup,
			assert: func(t *testing.T, hosts *[]model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "FindAll() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "FindAll() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			hosts, err := repository.FindAll()
			test.assert(t, hosts, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositoryFind(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		argument            *model.StaticDhcpHost
		expectedHost        *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            &ValidHost,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "Find() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "Find() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "HostNotFound",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            &UnknownHost,
			expectedHost:        nil,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "Find() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "Find() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			argument:         &ValidHost,
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "Find() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "Find() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			argument:         &ValidHost,
			setup:            voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "Find() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "Find() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			host, err := repository.Find(test.argument)
			test.assert(t, host, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositoryFindByIP(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		argument            net.IP
		expectedHost        *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            ValidHost.IPAddress,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "FindByIP() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "FindByIP() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "HostNotFound",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            UnknownHost.IPAddress,
			expectedHost:        nil,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "FindByIP() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "FindByIP() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			argument:         ValidHost.IPAddress,
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "FindByIP() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "FindByIP() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			argument:         ValidHost.IPAddress,
			setup:            voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "FindByIP() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "FindByIP() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			host, err := repository.FindByIP(test.argument)
			test.assert(t, host, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositoryFindByMac(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		argument            net.HardwareAddr
		expectedHost        *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            ValidHost.MacAddress,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "FindByMac() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "FindByMac() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "HostNotFound",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            UnknownHost.MacAddress,
			expectedHost:        nil,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "FindByMac() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "FindByMac() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			argument:         tests.ParseMAC(ValidMACAddress),
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "FindByMac() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "FindByMac() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			argument:         tests.ParseMAC(ValidMACAddress),
			setup:            voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "FindByMac() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "FindByMac() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			host, err := repository.FindByMac(test.argument)
			test.assert(t, host, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositoryDelete(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		argument            *model.StaticDhcpHost
		expectedHost        *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: DeletedValidHostFileContent,
			argument:            &ValidHost,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "Delete() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "Delete() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "LastHost",
			setupFileContent:    ValidHostFileContent,
			expectedFileContent: "",
			argument:            &ValidHost,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "Delete() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "Delete() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "HostNotFound",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            &UnknownHost,
			expectedHost:        nil,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "Delete() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "Delete() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			argument:         &ValidHost,
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "Delete() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "Delete() returned an unexpected error type")
			},
		},
		{
			name:             "ReadOnlyFileError",
			setupFileContent: AllHostsFileContent,
			argument:         &ValidHost,
			setup: func(tc *testcase) {
				f, _ := os.Open(tc.fileName)
				defer f.Close()
				f.Chmod(os.FileMode(0444))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "Delete() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrPermission, "Delete() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			argument:         &ValidHost,
			setup:            voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "Delete() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "Delete() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			host, err := repository.Delete(test.argument)
			test.assert(t, host, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositoryDeleteByIP(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		argument            net.IP
		expectedHost        *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: DeletedValidHostFileContent,
			argument:            ValidHost.IPAddress,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "DeleteByIP() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "DeleteByIP() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "LastHost",
			setupFileContent:    ValidHostFileContent,
			expectedFileContent: "",
			argument:            ValidHost.IPAddress,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "DeleteByIP() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "DeleteByIP() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "HostNotFound",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            UnknownHost.IPAddress,
			expectedHost:        nil,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "DeleteByIP() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "DeleteByIP() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			argument:         net.ParseIP(ValidIPAddress),
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "DeleteByIP() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "DeleteByIP() returned an unexpected error type")
			},
		},
		{
			name:             "ReadOnlyFileError",
			setupFileContent: AllHostsFileContent,
			argument:         net.ParseIP(ValidIPAddress),
			setup: func(tc *testcase) {
				f, _ := os.Open(tc.fileName)
				defer f.Close()
				f.Chmod(os.FileMode(0444))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "DeleteByIP() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrPermission, "DeleteByIP() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			argument:         net.ParseIP(ValidIPAddress),
			setup:            voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "DeleteByIP() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "DeleteByIP() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			host, err := repository.DeleteByIP(test.argument)
			test.assert(t, host, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositoryDeleteByMac(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		argument            net.HardwareAddr
		expectedHost        *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: DeletedValidHostFileContent,
			argument:            ValidHost.MacAddress,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "DeleteByMac() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "DeleteByMac() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "LastHost",
			setupFileContent:    ValidHostFileContent,
			expectedFileContent: "",
			argument:            ValidHost.MacAddress,
			expectedHost:        &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "DeleteByMac() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "DeleteByMac() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "HostNotFound",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AllHostsFileContent,
			argument:            UnknownHost.MacAddress,
			expectedHost:        nil,
			setup:               voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.NoError(t, err, "DeleteByMac() returned an expected error")
				assert.Equal(t, tc.expectedHost, host, "DeleteByMac() returned an unexpected host")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			argument:         tests.ParseMAC(ValidMACAddress),
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "DeleteByMac() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "DeleteByMac() returned an unexpected error type")
			},
		},
		{
			name:             "ReadOnlyFileError",
			setupFileContent: AllHostsFileContent,
			argument:         tests.ParseMAC(ValidMACAddress),
			setup: func(tc *testcase) {
				f, _ := os.Open(tc.fileName)
				defer f.Close()
				f.Chmod(os.FileMode(0444))
			},
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "DeleteByMac() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrPermission, "DeleteByMac() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			argument:         tests.ParseMAC(ValidMACAddress),
			setup:            voidSetup,
			assert: func(t *testing.T, host *model.StaticDhcpHost, err error, tc *testcase) {
				assert.Error(t, err, "DeleteByMac() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "DeleteByMac() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			host, err := repository.DeleteByMac(test.argument)
			test.assert(t, host, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}

func TestHostRepositorySave(t *testing.T) {
	type testcase struct {
		name                string
		fileName            string
		setupFileContent    string
		expectedFileContent string
		host                *model.StaticDhcpHost
		setup               func(tc *testcase)
		assert              func(t *testing.T, err error, tc *testcase)
	}
	voidSetup := func(tc *testcase) {}

	var testCases = []testcase{
		{
			name:                "Success",
			setupFileContent:    AllHostsFileContent,
			expectedFileContent: AddedUnknownHostFileContent,
			host:                &UnknownHost, // Adding a host that is not present on the hosts file
			setup:               voidSetup,
			assert: func(t *testing.T, err error, tc *testcase) {
				assert.NoError(t, err, "Save() returned an expected error")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:                "EmptyFile",
			setupFileContent:    "",
			expectedFileContent: ValidHostFileContent,
			host:                &ValidHost,
			setup:               voidSetup,
			assert: func(t *testing.T, err error, tc *testcase) {
				assert.NoError(t, err, "Save() returned an expected error")
				assertFileContent(t, tc.expectedFileContent, tc.fileName)
			},
		},
		{
			name:             "FileNotFoundError",
			setupFileContent: "",
			host:             &ValidHost,
			setup: func(tc *testcase) {
				os.Remove(tc.fileName)
			},
			assert: func(t *testing.T, err error, tc *testcase) {
				assert.Error(t, err, "Save() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrNotExist, "Save() returned an unexpected error type")
			},
		},
		{
			name:             "ReadOnlyFileError",
			setupFileContent: AllHostsFileContent,
			host:             &ValidHost,
			setup: func(tc *testcase) {
				f, _ := os.Open(tc.fileName)
				defer f.Close()
				f.Chmod(os.FileMode(0444))
			},
			assert: func(t *testing.T, err error, tc *testcase) {
				assert.Error(t, err, "Save() did NOT returned an expected error")
				assert.ErrorIs(t, err, os.ErrPermission, "Save() returned an unexpected error type")
			},
		},
		{
			name:             "InvalidHostsFileError",
			setupFileContent: InvalidHostsFileContent,
			host:             &ValidHost,
			setup:            voidSetup,
			assert: func(t *testing.T, err error, tc *testcase) {
				assert.Error(t, err, "Save() did NOT returned an expected error")
				// Just to ensure that we are not getting false negatives
				assert.NotErrorIs(t, err, os.ErrNotExist, "Save() returned an unexpected error type")
				// Verify that the file content hasn't changed
				assertFileContent(t, tc.setupFileContent, tc.fileName)
			},
		},
	}

	for _, test := range testCases {
		test.fileName = setUpStaticHostsFile(t, test.setupFileContent)
		t.Run(test.name, func(t *testing.T) {
			test.setup(&test)
			repository := NewRepository(test.fileName)
			err := repository.Save(test.host)
			test.assert(t, err, &test)
		})
		tearDownStaticHostsFile(t, test.fileName)
	}
}
