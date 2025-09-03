pragma solidity >=0.8.24;

contract CancunOpCodes {

   // Add a todo to the list
   function test() public {

       // blobbasefee
       assembly { pop(blobbasefee()) }
       
       // mcopy test
       assembly {
           // Allocate memory for source and destination
           let src := mload(0x40)
           let dest := add(src, 0x40)
           // Store a value at src
           mstore(src, 0x123456789abcdef0)
           // mcopy(dest, src, 32)
           mcopy(dest, src, 32)
           // Check that the value was copied
           if iszero(eq(mload(dest), 0x123456789abcdef0)) {
               revert(0, 0)
           }
       }

       // tstore and tload test (transient storage)
       assembly {
           let key := 0x42
           let val := 0xdeadbeef
           tstore(key, val)
           let loaded := tload(key)
           if iszero(eq(loaded, val)) {
               revert(0, 0)
           }
       }
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
