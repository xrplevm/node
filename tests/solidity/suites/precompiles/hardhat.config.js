require("@nomicfoundation/hardhat-toolbox");

/** @type import('hardhat/config').HardhatUserConfig */
module.exports = {
  solidity: {
    compilers: [
      {
        version: "0.8.18",
      },
      // This version is required to compile the werc9 contract.
      {
        version: "0.4.22",
      },
    ],
  },
  networks: {
    cosmos: {
      url: "http://127.0.0.1:8545",
      chainId: 1440002,
      accounts: [
        "0x88CBEAD91AEE890D27BF06E003ADE3D4E952427E88F88D31D61D3EF5E5D54305",
        "0x3B7955D25189C99A7468192FCBC6429205C158834053EBE3F78F4512AB432DB9",
        "0xe9b1d63e8acd7fe676acb43afb390d4b0202dab61abec9cf2a561e4becb147de",
      ],
    },
  },
};
