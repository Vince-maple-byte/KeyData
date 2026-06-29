package main

import (
	"fmt"
	"log"
	"net"

	"github.com/Vince-maple-byte/KeyData/internals/memtable"
	"github.com/Vince-maple-byte/KeyData/internals/sstable"
	"github.com/Vince-maple-byte/KeyData/network"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

/*

How the structure of the timestamp would look like
Timestamp | CRC32 error checksum| Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value |  Payload

How many bytes each element in the record is:
time stamp: 8 bytes
CRC32: 4 bytes
Tombstone: 1 byte
key size: 4 bytes
payload size: 4 bytes -> The header of the record is 21 bytes long
key: (key size) bytes
payload: (payload size) bytes -> The rest of the record is 21 + keySize + payloadSize bytes long

Have it so that in the Write method after we are done adding the record into the file, make sure to add flush(), we update the index map with the key and byte offset

Make a method where we would read the file and update all of the index map key/value pairs(key and byte offset) when we open the file for the first time when the program runs

Make some unit tests to test the methods

From there we can move on to SSTable, compaction, LSM Tables, and finally bloom tables


TODO: Fix the filePath for the project so that it doesn't have to be hardcoded.
./internals/data works for the production while ../data works for unit tests
*/

func main() {
	//We need to use this interface here to avoid an import cycle between sstable and memtable
	sstable.MergeList = func() sstable.ListMerger {
		return memtable.CreateSkiplist()
	}

	filePath :=

	mem := memtable.CreateMemtable()

	mem.MemtableStartUp()

	network.NetworkMemtable = func() network.InternalMemtable {
		return mem
	}

	port := 5773

	lis, err := net.Listen("tcp", fmt.Sprintf("localhost:%d", port))

	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	var opts []grpc.ServerOption

	grpcServer := grpc.NewServer(opts...)
	//network.pb.RegisterDataServer(grpcServer, newServer())
	network.RegisterDataServer(grpcServer, &network.Server{})
	reflection.Register(grpcServer)
	//
	err = grpcServer.Serve(lis)

	if err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	fmt.Printf("logging in port: %d", port)
}
