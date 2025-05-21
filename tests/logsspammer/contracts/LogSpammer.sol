// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.28;

contract LogSpammer {
    event Log(uint32 indexed foo, uint32 indexed bar, bytes data);

    constructor() {}

    function spam(uint32 n, uint32 foo, uint32 bar, bytes calldata data) public {
        for (uint32 i = 0; i < n; i++) {
            emit Log(foo, bar, data);
        }
    }
}
