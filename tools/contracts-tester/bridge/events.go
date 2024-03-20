package bridge

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
)

type EventIterator interface {
	Next() bool
}

func getIteratorLength(it EventIterator) int {
	eventsFound := 0
	for it.Next() {
		eventsFound += 1
	}

	return eventsFound
}

func (suite *BridgeTestSuite) getFilterOpts() *bind.FilterOpts {
	return &bind.FilterOpts{Start: suite.initBlock, Context: suite.ctx}
}

func (suite *BridgeTestSuite) runAddClaimAttestationEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterAddClaimAttestation(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering add claim attestation events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid add claim attestation events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runAddCreateAccountAttestationEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterAddCreateAccountAttestation(suite.getFilterOpts(), nil, []common.Address{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering add create account attestation events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid add create account attestation events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runClaimEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterClaim(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering claim events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid claim events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCreateAccountEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCreateAccount(suite.getFilterOpts(), []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering create account events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid create account events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCommitEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCommit(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering commit events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid commit events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCommitWithoutAddressEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCommitWithoutAddress(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering commit without address events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid commit without address events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCreateAccountCommitEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCreateAccountCommit(suite.getFilterOpts(), nil, []common.Address{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering create account commit events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid create account commit events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCreateClaimEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCreateClaim(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering create claim events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid create claim events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCreditEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCredit(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering credit events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid credit events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) getLatestCreateClaimEvent() *BridgeCreateClaim {
	events, err := suite.bridge.FilterCreateClaim(suite.getFilterOpts(), nil, []*big.Int{}, []common.Address{})
	if err != nil {
		suite.t.Errorf("Error filtering create claim events: '%+v'", err)
		return nil
	}

	var latestEvent *BridgeCreateClaim
	for events.Next() {
		latestEvent = events.Event
	}

	return latestEvent
}

func (suite *BridgeTestSuite) runCreateBridgeEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCreateBridge(suite.getFilterOpts(), nil)
	if err != nil {
		suite.t.Errorf("Error filtering create bridge events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid create bridge events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runCreateBridgeRequestEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterCreateBridgeRequest(suite.getFilterOpts())
	if err != nil {
		suite.t.Errorf("Error filtering create bridge request events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid create bridge request events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runPausedEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterPaused(suite.getFilterOpts())
	if err != nil {
		suite.t.Errorf("Error filtering paused events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid paused events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}

func (suite *BridgeTestSuite) runUnpausedEventTest(expectedEvents int) {
	events, err := suite.bridge.FilterUnpaused(suite.getFilterOpts())
	if err != nil {
		suite.t.Errorf("Error filtering unpaused events: '%+v'", err)
	} else {
		eventsFound := getIteratorLength(events)
		if eventsFound != expectedEvents {
			suite.t.Errorf("Invalid unpaused events - expected '%+v' got '%+v'", expectedEvents, eventsFound)
		}
	}
}
