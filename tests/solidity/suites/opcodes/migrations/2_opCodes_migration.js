/* eslint-disable no-undef */

const CancunOpCodes = artifacts.require('./CancunOpCodes.sol')
const PragueOpCodes = artifacts.require('./PragueOpCodes.sol')

module.exports = async function (deployer) {
  await deployer.deploy(CancunOpCodes)
  await deployer.deploy(PragueOpCodes)
}
