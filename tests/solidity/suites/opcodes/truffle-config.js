module.exports = {
  networks: {
    // Development network is just left as truffle's default settings
    cosmos: {
      host: '127.0.0.1', // Localhost (default: none)
      port: 8545, // Standard Ethereum port (default: none)
      network_id: 1440002, // Any network (default: none)
      gas: 8000000, // Increased gas limit for complex contracts
      gasPrice: 0, // Set to 0 for test network
      from: "0x5A7E818D849D4926CD2E2E05B8E934D05EE87A7C", // Address derived from the private key
      accounts: [
        "0x36DC1E881F351CFE35B79E3ED27C8EE737DFD7B48A9F2D43887E25D2F87625CB"
      ]
    }
  },
  compilers: {
    solc: {
        version: '0.8.24',
        settings: {
          evmVersion: 'cancun',
          optimizer: {
            enabled: false
          },
          viaIR: false
        }
    }
  }
}
