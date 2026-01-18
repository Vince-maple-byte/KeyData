# KeyDatabase
Key-Value Database built with Go.

## What does this project do?
This is a database built that stores and retrieves values based on a key/value pair that is stored in a file. 

## Functionality
The database is going to be made up of 3 main functions: PUT, GET, and DELETE

#### PUT: Adds new key/value pairs, as well as updates them with a new value if they already exist.

#### GET: Retrieves values from the database based on the key that was provided. *After phase 2 of development you should be able to retrieve values from a range of keys.*

#### DELETE: Deletes a key/value pair from the database.

---

This project will have 4 phases of development

## Phase 1: Map based Database

In this phase, the database will function in the background like this:
- Each key/value pair is stored into the file as a record that follows this format:
    - Timestamp | CRC32 error checksum | Tombstone (It's one byte long; basically 0 and 1 to determine whether this is a deleted key or not) | Key Size | Payload(Value) Size | Key Value | Payload
        - Timestamp = 8 bytes long
        - CRC32 checksum = 4 bytes long
        - Tombstone = 1 byte long
        - Key Size = 4 bytes long
        - Payload Size = 4 bytes long
        - Key = (Key Size) bytes long
        - Payload = (Payload Size) bytes long
    - *The entire recording is stored into the file in bytes, so we are responsible for encoding and decoding it properly.*
- **Look up strategy:** To enable efficient reads, the database maintains an in-memory map that stores the byte offset of the most recent record for each key.

- **Pros:** 
    - This allows for efficient look up of key/value pairs when doing reads.
    - Efficient writes to the file since each record is append to the file. 
    - Key/value pairs are able to be detected for any potential errors that might occur when trying to store the records into the file. (CRC32 error checksum).
    - Binary encoding allows for more data to be able to be stored into the file while also containing important information such as timestamps.
- **Cons:**
    - The append only nature of this method does not allow for range lookups since the file is not sorted.
    - Even though the Map helps with finding the byte offset of each record, everytime the program starts, the entire file has to be read from top to bottom to retrieve each key byte offset causing reads to be O(N) by default.
    - The file can grow extremely large to the point that it would take a while for the map to update the byte offset of each key when the program launches
    

## Phase 2: LSM Tree Database
In this phase, we are going to address some of the problems mentioned in the previous phase by adding three new concepts: SSTables, LSM Tables, Compaction.

- **SSTables:** Read-only, immutable files that store key/value records in sorted order by key. Once written to disk, an SSTable is never modified. This immutability simplifies concurrency, enables efficient sequential disk writes, and allows the database to use stable byte offsets for indexing and fast lookups.
- **LSM Tables (Log-Structured Merge Trees):**
An LSM Tree is the overall data structure that organizes multiple SSTables across different levels. New writes are first stored in memory (e.g., in a memtable) and are periodically flushed to disk as new SSTables. Reads are performed by merging results from the memtable and multiple SSTables, always returning the most recent version of a key.
- **Compaction:** Compaction is the process of merging multiple SSTables into fewer, larger SSTables while preserving sorted order. During compaction, obsolete records, overwritten values, and tombstones are discarded, reducing storage usage and read amplification. Compaction also ensures that lower levels of the LSM Tree contain non-overlapping key ranges, which significantly improves read and range query performance.

- Pros: 
    - Scales well with large datasets.
    - Efficient range queries.
    - High write throughput because of data being first added to memory and then flushed to immutable SSTables

- Cons:
    - The potential of needing to do multiple reads on multiple SSTables accross the different levels, which is more I/O intensive
    - Not having the ability to determine whether a key exist in the without needing to search through every SSTable, causing this to be very I/O intensive.

## Phase 3: Bloom Filters
In this phase we are going to fix the issue of not being able to determine if a key exists or not by using a bloom filter. 

Bloom filters are a data structure that can give a probabilistic change that something exists in a set.  

So this just means that a query would be returned that means either it is possibly in the set or definitely not in the set.

This is not a complete solution for the LSM Tree Database issue of determining whether a key exists or not, but it reduces the chances of always needing to go through the entire database to determine if a key exists, which greatly improves the performance of the database.

## Phase 4: Concurrency/Multi threading 
This phase of the project will take a look at how do databases handle multiple requests at the same time. 

Right now I don't have a full picture on how to do this, but some of the ideas that I have are this:

- **Multi-reader / single-writer model:**
A single writer thread handles all writes, while multiple reader threads serve read requests. This aligns well with LSM design due to immutable on-disk files and reduces write-side race conditions.

- **Thread pools:** Reusing threads instead of constantly creating and destroying them to reduce overhead.

This phase will involve further research into how production databases handle concurrency and synchronization.

---
## Final thoughts
This project is designed to incrementally explore real database internals, with a focus on correctness, performance tradeoffs, and system design rather than feature completeness.

