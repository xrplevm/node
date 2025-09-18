// SPDX-License-Identifier: MIT
pragma solidity >= 0.8.24;

library BLS12381 {
    address constant G1ADD   = address(0x0b);
    address constant G1MSM   = address(0x0c);
    address constant G2ADD   = address(0x0d);
    address constant G2MSM   = address(0x0e);
    address constant PAIRING = address(0x0f);
    address constant MAP_G1  = address(0x10);
    address constant MAP_G2  = address(0x11);

    function _pcall(address precompile, bytes memory input, uint outSize)
    private view returns (bool ok, bytes memory out)
    {
        out = new bytes(outSize);
        assembly {
            ok := staticcall(gas(), precompile, add(input, 0x20), mload(input), add(out, 0x20), outSize)
        }
    }

    function mapToG1(bytes memory fp) internal view returns (bytes memory g1) {
        (bool ok, bytes memory out) = _pcall(MAP_G1, fp, 128);
        require(ok, "mapToG1 fail"); return out;
    }

    function g1Add(bytes memory a, bytes memory b) internal view returns (bytes memory sum) {
        bytes memory inbuf = bytes.concat(a, b); // 128 + 128 = 256
        (bool ok, bytes memory out) = _pcall(G1ADD, inbuf, 128);
        require(ok, "g1Add fail"); return out;
    }

    function pairing(bytes memory inbuf) internal view returns (bool) {
        (bool ok, bytes memory out) = _pcall(PAIRING, inbuf, 32);
        require(ok, "pairing fail");
        return out[31] == 0x01;
    }
}

contract PragueOpCodes {
    using BLS12381 for bytes;

    // 64 zero bytes (a valid Fp element)
    bytes constant FP_ZERO = hex"0000000000000000000000000000000000000000000000000000000000000000";

    // 128 zero bytes (G1/G2 infinity encodings)
    bytes constant G1_INFINITY = hex"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000";
    bytes constant G2_INFINITY = hex"000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"; // 256B

    function test() public {
        // BLS12381
        bytes memory P = BLS12381.mapToG1(FP_ZERO);
        bytes memory S = BLS12381.g1Add(P, G1_INFINITY);
        require(keccak256(S) == keccak256(P), "error");

        bytes memory inbuf = bytes.concat(G1_INFINITY, G2_INFINITY);
        require(BLS12381.pairing(inbuf), "error pairing");
    }

    function test_revert() public {

        //revert
        assembly{ revert(0, 0) }
    }

    function test_invalid() public {

        //revert
        assembly{ invalid() }
    }

    function test_stop() public {

        //revert
        assembly{ stop() }
    }
}
