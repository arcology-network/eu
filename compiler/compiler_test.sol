// SPDX-License-Identifier: GPL-3.0
pragma solidity ^0.8.19;


contract Example {
    
    address constant public API = address(0x84); 

    uint[] public arr2 = [1, 2, 3];
    bytes private id;

    event logMsg(string message);

    constructor() {
        (bool success, bytes memory data) = address(API).call(abi.encodeWithSignature("new()"));       
        require(success, "Example.New() Failed");
        id = data;
 
        bytes memory byteArray = new bytes(75);
        for (uint  i = 0; i < 75; i ++) {
            byteArray[i] = 0x41;
        }

        require(length() == 0); 
        push(byteArray, arr2);  
        push(byteArray, arr2);          
        require(length() == 2); 

        bytes memory stored = get(1);
        require(stored.length == byteArray.length);
        for (uint  i = 0; i < byteArray.length; i ++) {
            require(stored[i] == byteArray[i]);
        }

        bytes memory elems = new bytes(5);
        for (uint  i = 0; i < elems.length; i ++) {
            elems[i] = 0xaa;
        }
        set(1, elems);
       
        stored = get(0);
        require(stored.length == byteArray.length);
        for (uint  i = 0; i < byteArray.length; i ++) {
            require(stored[i] == byteArray[i]);
        }

        stored = get(1);
        require(stored.length == elems.length); 
        for (uint  i = 0; i < elems.length; i ++) {
            require(stored[i] == elems[i]);
        }

        stored = pop();
        for (uint  i = 0; i < elems.length; i ++) {
            require(stored[i] == elems[i]);
        }
        require(length() == 1); 
    }

    function length() public returns(uint256) {  // 58 94 13 33
        (bool success, bytes memory data) = address(API).call(abi.encodeWithSignature("length(bytes) returns(uint256)", id));
        require(success, "Example.length() Failed");
        uint256 length = abi.decode(data, (uint256));
        return length;
    }

    function pop() public returns(bytes memory) { // 80 26 32 97
        (bool success, bytes memory data) = address(API).call(abi.encodeWithSignature("pop() returns(bytes)", id));
        require(success, "Example.pop() Failed");
        return abi.decode(data, (bytes)); 
    }

    function push(bytes memory elem, uint[] memory array) public { //9e c6 69 25
        (bool success, bytes memory data) = address(API).call(abi.encodeWithSignature("push(bytes,bytes)",  id, elem));
        require(success, "Example.push() Failed");
    }   


    function get(uint256 idx) public returns(bytes memory)  { // 31 fe 88 d0
        (bool success, bytes memory data) = address(API).call(abi.encodeWithSignature("get(bytes, uint256) returns(bytes)", id, idx));
        require(success, "Example.get() Failed");
        return abi.decode(data, (bytes));  
    }

    function set(uint256 idx, bytes memory elem) public { // 7a fa 62 38
        (bool success, bytes memory data) = address(API).call(abi.encodeWithSignature("set(bytes,uint256,bytes)", id, idx, elem));
        require(success, "Example.set() Failed");
    }
}
