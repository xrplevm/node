/* eslint-disable no-undef */

const OpCodes = artifacts.require('./CancunOpCodes.sol.sol')

module.exports = function (deployer) {
  deployer.deploy(OpCodes)
}
