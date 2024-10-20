# Keystore 

Keystore is an IPFS compliant keystore built on Chestnut. It implements an IPFS keystore interface, allowing it to be used natively with many existing IPFS implementations, and tools.

We recommend using AES256-CTR for encryption based in part on this 
[helpful analysis](https://www.highgo.ca/2019/08/08/the-difference-in-five-modes-in-the-aes-encryption-algorithm/)
of database encryption approaches and trade-offs from Shawn Wang, PostgreSQL Database Core.

For a detailed example on importing and using the Keystore, please check out the [Keystore](../examples/keystore) 
example under the `examples` folder.

### IMPORTANT!

```go
package main

import  (
    "github.com/ipfs/go-ipfs/keystore"
    "github.com/libp2p/go-libp2p/core/crypto"
)
```

Please **make sure** you import
[go-ipfs](github.com/ipfs/go-ipfs) and [go-libp2p-core](https://github.com/libp2p/go-libp2p-core/), 
and are **NOT** importing [go-ipfs-keystore](github.com/ipfs/go-ipfs-keystore) and 
[go-libp2p-crypto](github.com/libp2p/go-libp2p-crypto). Those repos are **DEPRECATED**, 
out of date, archived, etc. This will save you time and sanity.
