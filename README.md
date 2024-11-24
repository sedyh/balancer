### Balancer

This code uses some practices from:
- Clean Architecture
- Hexagonal Architecture
- Onion Architecture
- Ports and Adapters

Use this for reference:

- Primary adapter -> Controller
- Secondary adapter -> Repository
- Usecase implementation -> Service
- Entity -> Model

Taken decisions:

Use maglev hashing when choosing storage servers for your file partitions because:
- This will allow the server to get rid of the constant state due to a rare overhead when reading a file after adding a significant number of storage servers.
- This will allow for near-perfect distribution of file partitions across storage servers.
- This will make it easy to supplement the key with useful data, for example, for redundant file storage.
- This will make it easy to add replicas of parts of files for recovery if any of the nodes are unavailable through an additional identifier in the key.

Use http/2 instead of grpc to communicate with storage servers because:
- I won't get much benefit from selective compression, because most of the files will most likely already be compressed.
- The built-in load balancer does not use consistent hashing, therefore, it will not be possible to achieve the same location of partitions without constant state storage, which we would like to avoid.
- It is not a fact that the storage servers will be located next to the main service, so the close connection of the client and the server may harm in the future. However, even if they are in the same place, I will not get much benefit from the reduced overhead due to the lack of tls.

Direct the incoming stream directly to a temporary file, because:
- This will save the client from waiting for slicing and recording of each batch.
- This will allow you to rely on the actual size, rather than the content-size.
- This will allow you to check the integrity of the file through hashes without fully loading it into memory.

Cut the parts with rounding to a lesser by the power of two and put the remaining bytes into last piece:
- So that they lie flat on the disc due to the increased size of the last piece.

Include information about the number of parts in the first piece:
- In order to retrieve the number of parts to collect on downloading.

Use a basic hash equality check instead of redundancy codes because:
- Then the test task will take a lot longer, even with immutable files.

Exclude ratelimit, retry, circuitbreaker, fallbacks, hedge, etc:
- There is no time left for this, it can be added in the future