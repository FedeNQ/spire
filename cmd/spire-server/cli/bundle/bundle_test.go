package bundle

import (
	"bytes"
	"context"
	"crypto/x509"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/mitchellh/cli"
	"github.com/spiffe/spire/pkg/common/pemutil"
	"github.com/spiffe/spire/proto/server/datastore"
	"github.com/spiffe/spire/test/fakes/fakedatastore"
	"github.com/spiffe/spire/test/fakes/fakeregistrationclient"
	"github.com/stretchr/testify/suite"
)

const (
	cert1PEM = `-----BEGIN CERTIFICATE-----
MIIBKjCB0aADAgECAgEBMAoGCCqGSM49BAMCMAAwIhgPMDAwMTAxMDEwMDAwMDBa
GA85OTk5MTIzMTIzNTk1OVowADBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHyv
sCk5yi+yhSzNu5aquQwvm8a1Wh+qw1fiHAkhDni+wq+g3TQWxYlV51TCPH030yXs
RxvujD4hUUaIQrXk4KKjODA2MA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RAQH/BBkw
F4YVc3BpZmZlOi8vZG9tYWluMS50ZXN0MAoGCCqGSM49BAMCA0gAMEUCIA2dO09X
makw2ekuHKWC4hBhCkpr5qY4bI8YUcXfxg/1AiEA67kMyH7bQnr7OVLUrL+b9ylA
dZglS5kKnYigmwDh+/U=
-----END CERTIFICATE-----`
	cert2PEM = `-----BEGIN CERTIFICATE-----
MIIBKjCB0aADAgECAgEBMAoGCCqGSM49BAMCMAAwIhgPMDAwMTAxMDEwMDAwMDBa
GA85OTk5MTIzMTIzNTk1OVowADBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABB8V
bmlJ8YIuN9RuQ94PYanmkIRG7MkGV5mmrO6rFAv3SFd/uVlwYNkXrh0219eHUSD4
o+4RGXoiMFJKysw5GK6jODA2MA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RAQH/BBkw
F4YVc3BpZmZlOi8vZG9tYWluMi50ZXN0MAoGCCqGSM49BAMCA0gAMEUCIQDMKwYt
q+2ZoNyl4udPj7IMYIGX8yuCNRmh7m3d9tvoDgIgbS26wSwDjngGqdiHHL8fTcgg
diIqWtxAqBLFrx8zNS4=
-----END CERTIFICATE-----`
)

func TestBundleCommands(t *testing.T) {
	suite.Run(t, new(BundleSuite))
}

type BundleSuite struct {
	suite.Suite

	cert1 *x509.Certificate
	cert2 *x509.Certificate

	ds                 *fakedatastore.DataStore
	registrationClient *fakeregistrationclient.Client
	stdin              *bytes.Buffer
	stdout             *bytes.Buffer
	stderr             *bytes.Buffer

	showCmd   cli.Command
	setCmd    cli.Command
	listCmd   cli.Command
	deleteCmd cli.Command
}

func (s *BundleSuite) SetupTest() {
	cert1, err := pemutil.ParseCertificate([]byte(cert1PEM))
	s.Require().NoError(err)
	s.cert1 = cert1

	cert2, err := pemutil.ParseCertificate([]byte(cert2PEM))
	s.Require().NoError(err)
	s.cert2 = cert2

	s.stdin = new(bytes.Buffer)
	s.stdout = new(bytes.Buffer)
	s.stderr = new(bytes.Buffer)

	s.ds = fakedatastore.New()
	s.registrationClient = fakeregistrationclient.New(s.T(), "spiffe://example.test", s.ds, nil)

	testEnv := &env{
		stdin:  s.stdin,
		stdout: s.stdout,
		stderr: s.stderr,
	}
	clientMaker := func(context.Context, string) (*clients, error) {
		return &clients{
			r: s.registrationClient,
		}, nil
	}

	s.showCmd = newShowCommand(testEnv, clientMaker)
	s.setCmd = newSetCommand(testEnv, clientMaker)
	s.listCmd = newListCommand(testEnv, clientMaker)
	s.deleteCmd = newDeleteCommand(testEnv, clientMaker)
}

func (s *BundleSuite) TearDownTest() {
	// gotta close the registration client or we will leak a goroutine
	s.registrationClient.Close()
}

func (s *BundleSuite) AfterTest(suiteName, testName string) {
	s.T().Logf("SUITE: %s TEST:%s", suiteName, testName)
	s.T().Logf("STDOUT:\n%s", s.stdout.String())
	s.T().Logf("STDIN:\n%s", s.stdin.String())
	s.T().Logf("STDERR:\n%s", s.stderr.String())
}

func (s *BundleSuite) TestShowHelp() {
	s.showCmd.Help()
	s.Require().Equal(`Usage of bundle show:
  -serverAddr string
    	Address of the SPIRE server (default "localhost:8081")
`, s.stderr.String())
}

func (s *BundleSuite) TestShow() {
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://example.test",
		CaCerts:     s.cert1.Raw,
	})

	s.Require().Equal(0, s.showCmd.Run([]string{}))

	s.Require().Equal(s.stdout.String(), `-----BEGIN CERTIFICATE-----
MIIBKjCB0aADAgECAgEBMAoGCCqGSM49BAMCMAAwIhgPMDAwMTAxMDEwMDAwMDBa
GA85OTk5MTIzMTIzNTk1OVowADBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHyv
sCk5yi+yhSzNu5aquQwvm8a1Wh+qw1fiHAkhDni+wq+g3TQWxYlV51TCPH030yXs
RxvujD4hUUaIQrXk4KKjODA2MA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RAQH/BBkw
F4YVc3BpZmZlOi8vZG9tYWluMS50ZXN0MAoGCCqGSM49BAMCA0gAMEUCIA2dO09X
makw2ekuHKWC4hBhCkpr5qY4bI8YUcXfxg/1AiEA67kMyH7bQnr7OVLUrL+b9ylA
dZglS5kKnYigmwDh+/U=
-----END CERTIFICATE-----
`)
}

func (s *BundleSuite) TestSetHelp() {
	s.setCmd.Help()
	s.Require().Equal(`Usage of bundle set:
  -id string
    	SPIFFE ID of the trust domain
  -path string
    	Path to the bundle data
  -serverAddr string
    	Address of the SPIRE server (default "localhost:8081")
`, s.stderr.String())
}

func (s *BundleSuite) TestSetWithoutID() {
	rc := s.setCmd.Run([]string{})
	s.Require().Equal(1, rc)
	s.Require().Equal("id is required\n", s.stderr.String())
}

func (s *BundleSuite) TestSetWithInvalidTrustDomainID() {
	rc := s.setCmd.Run([]string{"-id", "spiffe://otherdomain.test/spire/server"})
	s.Require().Equal(1, rc)
	s.Require().Equal("\"spiffe://otherdomain.test/spire/server\" is not a valid trust domain SPIFFE ID: path is not empty\n", s.stderr.String())
}

func (s *BundleSuite) TestSetWithBadBundleData() {
	rc := s.setCmd.Run([]string{"-id", "spiffe://otherdomain.test"})
	s.Require().Equal(1, rc)
	s.Require().Equal("invalid bundle data: no PEM blocks\n", s.stderr.String())
}

func (s *BundleSuite) TestSetCreatesBundle() {
	s.stdin.WriteString(cert1PEM)
	s.assertBundleSet()
}

func (s *BundleSuite) TestSetUpdatesBundle() {
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://otherdomain.test",
		CaCerts:     []byte("BOGUSCERTS"),
	})
	s.stdin.WriteString(cert1PEM)
	s.assertBundleSet()
}

func (s *BundleSuite) TestSetCannotLoadBundleFromFile() {
	rc := s.setCmd.Run([]string{"-id", "spiffe://otherdomain.test", "-path", "/not/a/real/path/to/a/bundle"})
	s.Require().Equal(1, rc)
	s.Require().Equal("unable to load bundle data: open /not/a/real/path/to/a/bundle: no such file or directory\n", s.stderr.String())
}

func (s *BundleSuite) TestSetCreatesBundleFromFile() {
	tmpDir, err := ioutil.TempDir("", "spire-server-cli-test-")
	s.Require().NoError(err)
	defer os.RemoveAll(tmpDir)

	bundlePath := filepath.Join(tmpDir, "bundle.pem")

	s.Require().NoError(ioutil.WriteFile(bundlePath, []byte(cert1PEM), 0644))
	s.assertBundleSet("-path", bundlePath)
}

func (s *BundleSuite) TestListHelp() {
	s.listCmd.Help()
	s.Require().Equal(`Usage of bundle list:
  -id string
    	SPIFFE ID of the trust domain
  -serverAddr string
    	Address of the SPIRE server (default "localhost:8081")
`, s.stderr.String())
}

func (s *BundleSuite) TestListAll() {
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://domain1.test",
		CaCerts:     s.cert1.Raw,
	})
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://domain2.test",
		CaCerts:     s.cert2.Raw,
	})

	s.Require().Equal(0, s.listCmd.Run([]string{}))

	s.Require().Equal(s.stdout.String(), `****************************************
* spiffe://domain1.test
****************************************
-----BEGIN CERTIFICATE-----
MIIBKjCB0aADAgECAgEBMAoGCCqGSM49BAMCMAAwIhgPMDAwMTAxMDEwMDAwMDBa
GA85OTk5MTIzMTIzNTk1OVowADBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABHyv
sCk5yi+yhSzNu5aquQwvm8a1Wh+qw1fiHAkhDni+wq+g3TQWxYlV51TCPH030yXs
RxvujD4hUUaIQrXk4KKjODA2MA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RAQH/BBkw
F4YVc3BpZmZlOi8vZG9tYWluMS50ZXN0MAoGCCqGSM49BAMCA0gAMEUCIA2dO09X
makw2ekuHKWC4hBhCkpr5qY4bI8YUcXfxg/1AiEA67kMyH7bQnr7OVLUrL+b9ylA
dZglS5kKnYigmwDh+/U=
-----END CERTIFICATE-----

****************************************
* spiffe://domain2.test
****************************************
-----BEGIN CERTIFICATE-----
MIIBKjCB0aADAgECAgEBMAoGCCqGSM49BAMCMAAwIhgPMDAwMTAxMDEwMDAwMDBa
GA85OTk5MTIzMTIzNTk1OVowADBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABB8V
bmlJ8YIuN9RuQ94PYanmkIRG7MkGV5mmrO6rFAv3SFd/uVlwYNkXrh0219eHUSD4
o+4RGXoiMFJKysw5GK6jODA2MA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RAQH/BBkw
F4YVc3BpZmZlOi8vZG9tYWluMi50ZXN0MAoGCCqGSM49BAMCA0gAMEUCIQDMKwYt
q+2ZoNyl4udPj7IMYIGX8yuCNRmh7m3d9tvoDgIgbS26wSwDjngGqdiHHL8fTcgg
diIqWtxAqBLFrx8zNS4=
-----END CERTIFICATE-----
`)
}

func (s *BundleSuite) TestListOne() {
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://domain1.test",
		CaCerts:     s.cert1.Raw,
	})
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://domain2.test",
		CaCerts:     s.cert2.Raw,
	})

	s.Require().Equal(0, s.listCmd.Run([]string{"-id", "spiffe://domain2.test"}))

	s.Require().Equal(s.stdout.String(), `-----BEGIN CERTIFICATE-----
MIIBKjCB0aADAgECAgEBMAoGCCqGSM49BAMCMAAwIhgPMDAwMTAxMDEwMDAwMDBa
GA85OTk5MTIzMTIzNTk1OVowADBZMBMGByqGSM49AgEGCCqGSM49AwEHA0IABB8V
bmlJ8YIuN9RuQ94PYanmkIRG7MkGV5mmrO6rFAv3SFd/uVlwYNkXrh0219eHUSD4
o+4RGXoiMFJKysw5GK6jODA2MA8GA1UdEwEB/wQFMAMBAf8wIwYDVR0RAQH/BBkw
F4YVc3BpZmZlOi8vZG9tYWluMi50ZXN0MAoGCCqGSM49BAMCA0gAMEUCIQDMKwYt
q+2ZoNyl4udPj7IMYIGX8yuCNRmh7m3d9tvoDgIgbS26wSwDjngGqdiHHL8fTcgg
diIqWtxAqBLFrx8zNS4=
-----END CERTIFICATE-----
`)
}

func (s *BundleSuite) TestDeleteHelp() {
	s.deleteCmd.Help()
	s.Require().Equal(`Usage of bundle delete:
  -id string
    	SPIFFE ID of the trust domain
  -serverAddr string
    	Address of the SPIRE server (default "localhost:8081")
`, s.stderr.String())
}

func (s *BundleSuite) TestDeleteWithoutID() {
	s.Require().Equal(1, s.deleteCmd.Run([]string{}))
	s.Require().Equal("id is required\n", s.stderr.String())
}

func (s *BundleSuite) TestDelete() {
	s.createBundle(&datastore.Bundle{
		TrustDomain: "spiffe://domain1.test",
		CaCerts:     s.cert1.Raw,
	})

	s.Require().Equal(0, s.deleteCmd.Run([]string{"-id", "spiffe://domain1.test"}))
	s.Require().Equal("bundle deleted.\n", s.stdout.String())

	_, err := s.ds.FetchBundle(context.Background(), &datastore.FetchBundleRequest{
		TrustDomain: "spiffe://domain1.test",
	})
	s.Require().EqualError(err, "no such bundle")
}

func (s *BundleSuite) assertBundleSet(extraArgs ...string) {
	rc := s.setCmd.Run(append([]string{"-id", "spiffe://otherdomain.test"}, extraArgs...))
	s.Require().Equal(0, rc)
	s.Require().Equal("bundle set.\n", s.stdout.String())

	// make sure it made it into the datastore
	resp, err := s.ds.FetchBundle(context.Background(), &datastore.FetchBundleRequest{
		TrustDomain: "spiffe://otherdomain.test",
	})
	s.Require().NoError(err)
	s.Require().NotNil(resp)
	s.Require().NotNil(resp.Bundle)
	s.Require().Equal(s.cert1.Raw, resp.Bundle.CaCerts)
}

func (s *BundleSuite) createBundle(bundle *datastore.Bundle) {
	_, err := s.ds.CreateBundle(context.Background(), &datastore.CreateBundleRequest{
		Bundle: bundle,
	})
	s.Require().NoError(err)
}
