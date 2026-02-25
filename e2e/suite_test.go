package e2e

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/suite"
)

type E2ETestSuite struct {
	suite.Suite
	cfg              *Config
	atomoneClientID  string
	gnoClientID      string
	senderAddress    string
	gnoSenderAddress string
	atomoneContainer string
	gnoContainer     string
}

func TestE2E(t *testing.T) {
	suite.Run(t, new(E2ETestSuite))
}

func (s *E2ETestSuite) SetupSuite() {
	cfg, err := LoadConfig()
	s.Require().NoError(err, "load config")
	s.Require().NotEmpty(cfg.TestMnemonic, "TEST_MNEMONIC must be set")
	s.cfg = cfg

	// Verify atomone is healthy
	_, err = httpGet(cfg.AtomoneREST + "/cosmos/base/tendermint/v1beta1/node_info")
	s.Require().NoError(err, "atomone REST not reachable at %s", cfg.AtomoneREST)

	// Get container IDs
	s.atomoneContainer, err = getContainerID("atomone")
	s.Require().NoError(err, "get atomone container ID")
	s.gnoContainer, err = getContainerID("gno")
	s.Require().NoError(err, "get gno container ID")

	// Verify gno is healthy via gnokey query
	_, err = gnoQuery(s.gnoContainer, "r/aib/ibc/core", "")
	s.Require().NoError(err, "gno node not reachable")

	// Get sender address from the validator key in the atomone container
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	stdout, stderr, err := dockerExec(ctx, s.atomoneContainer,
		"atomoned", "keys", "show", "validator", "-a",
		"--keyring-backend", "test", "--home", "/root/.atomone")
	s.Require().NoError(err, "get validator address: %s", stderr)
	s.senderAddress = strings.TrimSpace(stdout)
	s.Require().NotEmpty(s.senderAddress)
	s.T().Logf("Sender address: %s", s.senderAddress)

	// Recover test key in gnokey for Gno→AtomOne transfers
	err = recoverGnoKey(s.gnoContainer, "test", cfg.TestMnemonic)
	s.Require().NoError(err, "recover gnokey test key")
	s.gnoSenderAddress, err = gnoKeyAddress(s.gnoContainer, "test")
	s.Require().NoError(err, "get gnokey test address")
	s.T().Logf("Gno sender address: %s", s.gnoSenderAddress)

	// Wait for IBC clients
	s.waitForIBCClients()
}

func (s *E2ETestSuite) waitForIBCClients() {
	r := s.Require()

	// Wait for client on AtomOne
	r.Eventually(func() bool {
		id, err := queryAtomOneClientStates(s.cfg.AtomoneREST)
		if err != nil {
			return false
		}
		s.atomoneClientID = id
		return true
	}, 3*time.Minute, 2*time.Second, "IBC client on AtomOne not created in time")
	s.T().Logf("AtomOne client ID: %s", s.atomoneClientID)

	// Wait for client on Gno
	r.Eventually(func() bool {
		id, err := queryGnoClients(s.gnoContainer)
		if err != nil {
			return false
		}
		s.gnoClientID = id
		return true
	}, 3*time.Minute, 2*time.Second, "IBC client on Gno not created in time")
	s.T().Logf("Gno client ID: %s", s.gnoClientID)

	// Wait for counterparty registration
	r.Eventually(func() bool {
		_, err := queryGnoClientCounterparty(s.gnoContainer, s.gnoClientID)
		return err == nil
	}, 1*time.Minute, 2*time.Second, "counterparty not registered on Gno in time")
	s.T().Log("Counterparty registered")
}
