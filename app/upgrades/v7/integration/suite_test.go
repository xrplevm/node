package integration

import (
	"testing"

	"github.com/stretchr/testify/suite"
)

func TestUpgradeTestSuite(t *testing.T) {
	suite.Run(t, new(UpgradeTestSuite))
}
