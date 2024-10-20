# 🌰 &nbsp;Chestnut

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/jrapoport/chestnut/test.yml?branch=master&style=flat-square) 
[![Go Report Card](https://goreportcard.com/badge/github.com/jrapoport/chestnut?style=flat-square&)](https://goreportcard.com/report/github.com/jrapoport/chestnut) 
[![Codecov branch](https://img.shields.io/codecov/c/github/jrapoport/chestnut/master?style=flat-square&token=7REY4BDPHW)](https://codecov.io/gh/jrapoport/chestnut)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/jrapoport/chestnut?style=flat-square) 
[![GitHub](https://img.shields.io/github/license/jrapoport/chestnut?style=flat-square)](https://github.com/jrapoport/chestnut/blob/master/LICENSE)

[![Buy Me A Coffee](https://img.shields.io/badge/buy%20me%20a%20coffee-☕-6F4E37?style=flat-square)](https://www.buymeacoffee.com/jrapoport)


Chestnut is encrypted storage for Go. The goal was an easy to use encrypted 
store with helpful features that was quick to set up, but highly flexible. 

Chestnut is written in pure go and designed **not** to have strong opinions 
about things like storage, compression, hashing, secrets, or encryption. 
Chestnut is a storage chest, and not a datastore itself. As such, Chestnut must 
be backed by a storage solution. 

Currently, Chestnut supports [BBolt](https://github.com/etcd-io/bbolt) and
[NutsDB](https://github.com/nutsdb/nutsdb) as backing storage.

## Table of Contents
- [Getting Started](#getting-started)
    * [Installing](#installing)
    * [Importing Chestnut](#importing-chestnut)
        + [Requirements](#requirements)
- [Storage](#storage)
    * [Built-in](#supported)
        + [BBolt](#bbolt)
        + [NutsDB](#nutsdb)
    * [Planned](#planned)
- [Encryption](#encryption)
    * [AES256-CTR](#aes256-ctr)
    * [Custom Encryption](#custom-encryption)
    * [Chained Encryption](#chained-encryption)
    * [Sparse Encryption](#sparse-encryption)
        + [What is "sparse" encryption?](#what-is--sparse--encryption-)
        + [Enabling Sparse Encryption](#enabling-sparse-encryption)
        + [Using Sparse Encryption](#using-sparse-encryption)
            - [Sparse Loading](#sparse-loading)
            - [Decryption](#decryption)
- [Secrets](#secrets)
    + [TextSecret](#textsecret)
    + [ManagedSecret](#managedsecret)
    + [SecureSecret](#securesecret)
- [Compression](#compression)
    * [Zstandard](#zstandard)
    * [Custom Compression](#custom-compression)
    * [Compression + Sparse Encryption](#compression---sparse-encryption)
- [Operations](#operations)
    * [Basic Operations](#basic-operations)
        + [Put](#put)
        + [Get](#get)
        + [Delete](#delete)
    * [Struct Operations](#struct-operations)
        + [Save](#save)
        + [Load](#load)
        + [Sparse](#sparse)
    * [Keyed Operations](#keyed-operations)
        + [SaveKeyed](#savekeyed)
        + [LoadKeyed](#loadkeyed)
        + [SparseKeyed](#sparsekeyed)
    * [Extra Operations](#extra-operations)
        + [Has](#has)
        + [List](#list)
        + [Export](#export)
- [Struct Field Tags](#struct-field-tags)
    * [Secure](#secure)
    * [Hash](#hash)
        + [SHA256](#sha256)
        + [Hash Prefix](#hash-prefix)
    * [Multiple Tags](#multiple-tags)
- [Disable Overwrites](#disable-overwrites)
- [Keystore](#keystore)
    * [Importing Keystore](#importing-keystore)
    * [Important Note](#important-note)
- [Logging](#logging)
    + [Logrus Logger](#logrus-logger)
    + [Zap Logger](#zap-logger)
    + [Standard Logger](#standard-logger)
    + [Storage](#storage-1)
- [Examples](#examples)
- [Known Issues](#known-issues)
- [Misc](#misc)
    * [JSON encoding](#json-encoding)

## Getting Started

### Installing

To start using Chestnut, install Go (version 1.11+) and run `go get`:

```sh
$ go get -u github.com/jrapoport/chestnut
```

### Importing Chestnut

To use Chestnut as an encrypted store, import as:

```go
import (
  "github.com/jrapoport/chestnut"
  "github.com/jrapoport/chestnut/encryptor/aes"
  "github.com/jrapoport/chestnut/encryptor/crypto"
  "github.com/jrapoport/chestnut/storage/nuts"
)

// use nutsdb for storage
store := nuts.NewStore(path)

// use AES256-CFB for encryption
opt := chestnut.WithAES(crypto.Key256, aes.CFB, mySecret)

cn := chestnut.NewChestnut(store, opt)
if err := cn.Open(); err != nil {
    return err
}

defer cn.Close()

```

#### Requirements
Chestnut has two requirements:
1) [Storage](#storage) that supports the `storage.Storage` interface 
   (with a lightweight adapter).
2) [Encryption](#encryption) which supports the `crypto.Encryptor` interface.

## Storage
Chestnut will work seamlessly with **any** storage solution (or adapter) that 
supports the`storage.Storage` interface.

### Built-in

Currently, Chestnut has built-in support for
[BBolt](https://github.com/etcd-io/bbolt) and 
[NutsDB](https://github.com/nutsdb/nutsdb).

#### BBolt

https://github.com/etcd-io/bbolt
Chestnut has built-in support for using
[BBolt](https://github.com/etcd-io/bbolt) as a backing store.

To use bbolt for a backing store you can import Chestnut's `bolt` package
and call `bolt.NewStore()`:

```go
import "github.com/jrapoport/chestnut/storage/bolt"

//use or create a bbolt backing store at path
store := bolt.NewStore(path)

// use bbolt for the storage chest
cn := chestnut.NewChestnut(store, ...)
```
 
#### NutsDB

https://github.com/nutsdb/nutsdb  
Chestnut has built-in support for using 
[NutsDB](https://github.com/nutsdb/nutsdb) as a backing store.  

To use nutsDB for a backing store you can import Chestnut's `nuts` package
and call `nuts.NewStore()`:

```go
import "github.com/jrapoport/chestnut/storage/nuts"

//use or create a nutsdb backing store at path
store := nuts.NewStore(path)

// use nutsdb for the storage chest
cn := chestnut.NewChestnut(store, ...)
```

### Planned

Other K/V stores like LevelDB.

[GORM](https://github.com/go-gorm/gorm) (probably not)
  Gorm is an ORM, so while it's not a datastore per se, it could be adapted 
  to support sparse encryption and would mean automatic support for databases 
  like mysql, sqlite, etc. However, most (if not all) of those DBs *already* 
  support built-in encryption, so w/o some compelling use-case that's not 
  already covered I don't see a lot of value-add.

## Encryption
Chestnut supports several flavors of AES out of the box:
* AES128-CFB, AES192-CFB, and AES256-CFB
* AES128-CTR, AES192-CTR, and AES256-CTR
* AES128-GCM, AES192-GCM, and AES256-GCM

You can add AES encryption to Chestnut by passing the `chestnut.WithAES()` option:
```go
opt := chestnut.WithAES(crypto.Key256, aes.CFB, mySecret)
```

### AES256-CTR
For encryption we recommend using AES256-CTR. We chose AES256-CTR based in part
on this [helpful analysis](https://www.highgo.ca/2019/08/08/the-difference-in-five-modes-in-the-aes-encryption-algorithm/)
from Shawn Wang, PostgreSQL Database Core.

### Custom Encryption
Chestnut supports drop-in custom encryption. A struct that supports the 
`crypto.Encryptor` interface can be used with the `chestnut.WithEncryptor()` 
option. 

Supporting `crypto.Encryptor` interface is straightforward and mainly consists 
of vending the following two methods:

```go
// Encrypt returns data encrypted with the secret.
Encrypt(plaintext []byte) (ciphertext []byte, err error)

// Decrypt returns data decrypted with the secret.
Decrypt(ciphertext []byte) (plaintext []byte, err error)
```

### Chained Encryption
Chestnut also supports chained encryption which allows data to be arbitrarily 
transformed by a chain of Encryptors in a FIFO order.

A chain of `crypto.Encryptor`s can be passed to Chestnut with the 
`chestnut.WithEncryptorChain` option:

```go
opt := chestnut.WithEncryptorChain(
    encryptor.NewAESEncryptor(crypto.Key128, aes.CFB, secret1),
    encryptor.NewAESEncryptor(crypto.Key192, aes.CTR, secret2),
    encryptor.NewAESEncryptor(crypto.Key256, aes.GCM, secret3),
)
```

or by using a `crypto.ChainEncryptor` with the `chestnut.WithEncryptor` option:

```go
encryptors := []crypto.Encryptor{
    encryptor.NewAESEncryptor(crypto.Key128, aes.CFB, secret1),
    encryptor.NewAESEncryptor(crypto.Key192, aes.CTR, secret2),
    encryptor.NewAESEncryptor(crypto.Key256, aes.GCM, secret3),
}
chain := crypto.NewChainEncryptor(encryptors...)
opt := chestnut.WithEncryptor(chain)
```

If you use both the `chestnut.WithEncryptor` and the
`chestnut.WithEncryptorChain` options, the `crypto.Encryptor` from
`chestnut.WithEncryptor` will be **prepended*** to the chain.

### Sparse Encryption
Chestnut supports the sparse encryption of structs.

Sparse encryption is a transparent feature of saving structs with
`Chestnut.Save()`, `Chestnut.Load()`, and `Chestnut.Sparse()`; or structs that
support the `value.Keyed` interface with `Chestnut.SaveKeyed()`,
`Chestnut.LoadKeyed()`, and `Chestnut.SparseKeyed()`.

#### What is "sparse" encryption?
With sparse encryption, only struct fields marked as `secure` will be encrypted.
The remaining "plaintext" fields are encoded and stored separately. 

This allows you to load a "sparse" copy of the struct by calling
`Chestnut.Sparse()` or `Chestnut.SparseKeyed()` (if you have a `value.Keyed`
value) and examine the plaintext fields **without** the overhead of decryption.
When a sparse struct is loaded, *the contents of struct fields marked as
`secure` are replaced by empty values*.

#### Enabling Sparse Encryption
Chestnut uses struct tags to indicate which specific struct fields should be 
encrypted. To enable sparse encryption for a struct, add the `secure` tag option
to the JSON tag of *at least one* struct field:

```go
SecretKey string `json:",secure"` // 'secure' option (bare minimum)
```
like so: 

```go
type MySparseStruct struct {
    SecretValue string `json:"secret_value,secure"` // <-- add 'secure' here
    PublicValue string `json:"public_value"`
}
```

#### Using Sparse Encryption
Structs can be sparsely encrypted by calling `Chestnut.Save()`, or if the struct 
supports the `value.Keyed` interface, `Chestnut.SaveKeyed()`. Chestnut will 
automatically detect the `secure` tag and do the rest. 

**If no `secure` fields are found, Chestnut will encrypt the entire struct.**

```go
sparseObj := &MySparseStruct{
    SecretValue: "this is a secret",
    PublicValue: "this is public",
}

err := cn.Save("my-namespace",  []byte("my-key"), sparseObj)
```

When `MySparseStruct` is saved, Chestnut will detect the `secure` struct field
and **only encrypt** those fields. Any remaining fields will be encoded as
plaintext. In the case of `MySparseStruct`this means that `SecretValue`**will
be** encrypted prior to being encoded, and `PublicValue`**will not** be
encrypted.

##### Sparse Loading 
A sparse struct can be loaded by calling `Chestnut.Sparse()`, or if the struct
supports the `value.Keyed` interface, `Chestnut.SparseKeyed()`. When these
methods are called to load a sparsely encrypted struct, a partially decoded 
struct will be returned, but the no decryption will occur. Secure fields will 
instead be replaced by empty values.

```go
sparseObj := &MySparseStruct{}

err := cn.Sparse("my-namespace",  []byte("my-key"), sparseObj)
```

Examining the struct will reveal that the `secure` fields were replaced with 
empty values, and not decrypted.

```go
*MySparseStruct{
    SecretValue: ""
    PublicValue: "this is public"
}
```

**Only sparsely encrypted structs can be sparsely loaded**  
If `Chestnut.Sparse()` or `Chestnut.SparseKeyed()` is called on a struct that 
was not sparsely encrypted, the fully decrypted struct will be returned.

##### Decryption

A sparsely encrypted struct can be fully decrypted by calling `Chestnut.Load()`,
or if the struct supports the `value.Keyed` interface,
`Chestnut.LoadKeyed()`. When any of those methods are called on a sparsely
encrypted struct, a fully decrpted copy of the struct is returned.

## Secrets

Chestnut secrets are handled through the `crypto.Secret` interface. The 
`crypto.Secret` interface is designed to provide a high degree of flexibility 
around how you store, retrieve, and manage the secrets you use for encryption. 

While Chestnut currently only comes with AES symmetric key encryption, the
`crypto.Secret` interface can easily be adapted to support other forms of 
encryption like a private key-based `crypto.Encryptor`.

Chestnut currently provides three basic immplementations of the `crypto.Secret` 
interface which should cover most cases.

#### TextSecret

`crypto.TextSecret` provides a lightweight wrapper around a plaintext `string`.

```go
textSecret := crypto.NewTextSecret("a-secret")
```

#### ManagedSecret

`crypto.ManagedSecret` provides a unique ID alongside a plaintext `string`
secret. You can use this id to securely track the secret if you use external
vaults or functionality like rollover.

```go
managedSecret := crypto.NewManagedSecret("my-secret-id", "a-secret")
```

#### SecureSecret

`crypto.SecureSecret` provides a unique id for a secret alongside an
`openSecret()` callback which returns a byte representation of the secret for
encryption and decryption on `SecureSecret.Open()`. When `crypto.SecureSecret`
calls `openSecret()` it will pass a copy of itself as a `crypto.Secret`. This
allows for remote loading of the secret based on its id, or using a secure
in-memory storage solution for the secret like
[memguarded](https://github.com/n0rad/memguarded).

```go
openSecret := func(s crypto.Secret) []byte {
	// fetch the secret 
    mySecret := getMySecretFromTheVault(s.ID())
    return mySecret
}
secureSecret := crypto.NewSecureSecret("my-secret-id", openSecret)
```

## Compression

Chestnut supports compression of the encoded data. Compression takes place
*prior to* encryption.

Compression can be enabled through the `chestnut.WithCompression` option and 
passing it a supported compression format:

```go
opt := chestnut.WithCompression(compress.Zstd)
```

Data compressed while `chestnut.WithCompression` is active with a supported
compression format will continue to be correctly decompressed when read *even
if* compression is no longer active (i.e. `chestnut.WithCompression` is no
longer being used). This is not true with custom compression. Data compressed
using custom compression cannot be decompressed if that custom compression is
disabled.

### Zstandard

Chestnut currently supports [Zstandard](https://facebook.github.io/zstd/)
compression out of the box with the `compress.Zstd` format option. To enable 
Zstandard compression, call `chestnut.WithCompression` passing `compress.Zstd`
as the compression format:

```go
opt := chestnut.WithCompression(compress.Zstd)
```

Please Note: I have no affiliation with Facebook (past or present) and just 
liked this compression format.

### Custom Compression

If you wish to supply your own compression routines you can do so easily with
the `chestnut.WithCompressors` option:

```go
opt := chestnut.WithCompressors(myCompressorFn, myDecompressorFn)
```

Your two custom compression functions, a compressor `compress.CompressorFunc`, 
and a decompressor `compress.DecompressorFunc` must have the following format:

```go
Compressor(data []byte) (compressed []byte, err error)

Decompressor(compressed []byte) (data []byte, err error)
```

### Compression + Sparse Encryption

Enabling compression will *not* affect sparse encryption. Sparsely encrypted
values compress their secure and plaintext encodings independently.

## Operations

Chestnut supports all basic CRUD operations with a few extras.

All `WRITE` operations: `Chestnut.Put()`, `Chestnut.Save()`, &
`Chestnut.SaveKeyed()`, will encrypt data prior to it being stored.

All `READ` operations: `Chestnut.Get()`, `Chestnut.Load()`, &
`Chestnut.LoadKeyed()`, will decrypt data prior to it being returned.

All `SPARSE` operations: `Chestnut.Sparse()`, & `Chestnut.SparseKeyed()`,
will **not** decrypt data prior to it being returned.

**In all cases no record of the plaintext data is kept**  
(even with DebugLevel logging enabled).

### Basic Operations

#### Put

To save an encrypted value to a namespaced key in the storage chest, use the 
`Chestnut.Put()` function:

```go
err := cn.Put("my-namespace", []byte("my-key"), []byte("plaintext"))
```

This will set the value of the `"my-key"` key to the encrypted ciphertext of 
`"plaintext"` in the `my-namespace` namespace. If a namespace does not exist,
it will be automatically created.

If the key already exists, and the storage chest was initialized with the 
`chestnut.OverwritesForbidden` option, this call will fail with ErrForbidden.

To retrieve this value, we can use the `Chestnut.Get()` function:

#### Get

To retrieve a decrypted value from a namespaced key in the storage chest, we can 
use the `Chestnut.Get()` function:

```go
plaintext, err := cn.Get("my-namespace", []byte("my-key"))
```

#### Delete

Use the `Chestnut.Delete()` function to delete a key from the store.

```go
err := cn.Delete("my-namespace", []byte("my-key"))
```

### Struct Operations

Chestnut provides several functions for working directly with structs. In
addition to handling the marshalling, encoding and encryption of structs for
you, these functions provide automatic support for the Chestnut struct field tag
options `secure` and `hash`. SEE: [Struct Field Tags](#struct-field-tags) for
more detail.

#### Save

To encrypt and save a struct to the store we can use the `Chestnut.Save()` 
function:

```go
err := cn.Save("my-namespace", []byte("my-key"), myStruct)
```

Chestnut will marshal and encrypt the encoded byte representation. If the struct
supports the `secure` struct field tag option on one of its fields, Chestnut 
will automatically sparsely encrypt the struct. Other supported struct field tag
options will also be applied. 

#### Load

To retrieve the fully decrypted struct, we can use the `Chestnut.Load()` 
function:

```go
err := cn.Load("my-namespace", []byte("my-key"), &myStruct)
```

#### Sparse

`Chestnut.Sparse()` loads the struct at key and returns the sparsely decoded
result. Unlike `Chestnut.Load()`, it **does not decrypt** the encoded struct and
secure fields are replaced with empty values. To retrieve a sparse value, we
can use the `Chestnut.Sparse()` function:

```go
err := cn.Sparse("my-namespace", []byte("my-key"), &myStruct)
```

If the struct was not saved as a sparsely encoded struct this has no effect and
is equivalent to calling `Chestnut.Load()`. Structs must have been saved with
secure fields to be loaded as sparse structs by `Chestnut.Sparse()`.

When a sparse struct is returned, any fields marked as `secure` will be decoded  
as nil or empty values. For more information, please see the section on 
[sparse encryption](#sparse-encryption).

### Keyed Operations

Chestnut provides several convenience functions for working with struct values
that support the `value.Keyed` interface. Keyed values can supply their own 
namespace and keys via calls to `Keyed.Namespace()` and `Keyed.Keys()`, 
respectively. Internally these functions are equivalent to calling 
`Chestnut.Save()`, `Chestnut.Load()`, and `Chestnut.Sparse()` with an explicit 
namespace and key.

#### SaveKeyed

To encrypt and store a struct that implements the `value.Keyed` interface to
the store we can use the `Chestnut.SaveKeyed()` function:

```go
err := cn.SaveKeyed(myKeyedStruct)
```

To save a keyed struct with `Chestnut.SaveKeyed()`, the struct must be 
initialized with namespace and key you want to save it to *prior to* calling 
`Chestnut.LoadKeyed()` in order to satisfy the `value.Keyed` interface:

```go
ko := MyKeyedValue{ name: "my-namespase", key: "my key"}
err := cn.SaveKeyed(&ko)
```

For more information, please see [Chestnut.Save()](#save)

#### LoadKeyed

To retrieve the fully decrypted struct that implements the `value.Keyed` 
interface, we can use the `Chestnut.LoadKeyed()` function:

```go
err := cn.LoadKeyed(&myKeyedStruct)
```

To load a keyed struct with `Chestnut.LoadKeyed()`, the struct must be
initialized with namespace and key you want to retrieve *prior to* calling
`Chestnut.LoadKeyed()` in order to satisfy the `value.Keyed` interface:

```go
ko := MyKeyedValue{ name: "my-namespase", key: "my key"}
err := cn.LoadKeyed(&ko)
```

For more information, please see [Chestnut.Load()](#load)

#### SparseKeyed

To retrieve a sparsely encrypted struct that implements the `value.Keyed`
interface, we can use the `Chestnut.SparseKeyed()` function:

```go
err := cn.SparseKeyed(&myKeyedStruct)
```

To load a sparse keyed struct with `Chestnut.SparseKeyed()`, the struct must be
initialized with namespace and key you want to retrieve *prior to* calling
`Chestnut.SparseKeyed()` in order to satisfy the `value.Keyed` interface:

```go
ko := MyKeyedValue{ name: "my-namespase", key: "my key"}
err := cn.SparseKeyed(&ko)
```

For more information, please see [Chestnut.Sparse()](#sparse)

### Extra Operations

Chestnut supports a few additional functions that you might find helpful. In the 
future more may be added assuming they can be reasonably supported by the
`storage.Storage` interface and generally make sense to do so. If there is a 
specific function you'd like to see added, please feel free to open an issue 
request to discuss.

#### Has

You can check to see if a key exists by calling `Chestnut.Has()`. If the key is
found, it will return true, otherwise false. If an error occured,
`Chestnut.Has()` will return false along with the error.

```go
has, err := cn.Has("my-namespace", []byte("my-key"))
```

#### List

To get a list of all the keys for a namespace you can call `Chestnut.List()`:

```go
keys, err := cn.List("my-namespace")
```

#### ListAll

To get a mapped list of all keys in the store organized by namespace you can call
`Chestnut.ListAll()`:

```go
keymap, err := cn.ListAll()
```

#### Export

To export the storage chest to another path you can call `Chestnut.Export()`:

```go
err := cn.Export("/a/path/someplace")
```

Chestnut cannot be exported to its current location. If you call 
`Chestnut.Export()` and pass the path to Chestnut's current location an error
will be returned.

## Struct Field Tags

Chestnut currently supports two extensions to the `` `json` `` struct field tag
as options: `secure` when added to the `` `json` `` tag, `secure` marks the
field for sparse encryption. `hash` when added to the `` `json` `` tag, `hash`
marks *a string* field for hashing.

These options will be automatically detected and applied when the struct is
saved with `Chestnut.Save()`, or `Chestnut.SaveKeyed()`.

**NOTE:** The order in which the tag options appear is unimportant.

```go
// these are equivalent

`json:"my_value,secure,omitempty"`

`json:"my_value,omitempty,secure"`
```

In the future, Chestnut will also support its own struct field tag. `` `cn` ``.

### Secure

When the `secure` option is added to a `` `json` `` struct field tag, the struct
field is marked for sparse encryption. If Chestnut detects a `secure` option on
a struct field tag, **only** those fields marked with `secure` will be encrypted.
If no `secure` fields are found, Chestnut will encrypt the entire struct.

To mark a struct field as secure, just add `secure` as an option to a `` `json`
`` struct field tag (like `omitempty`). The following are some examples of how
the `secure` option can be added to the `` `json` `` struct field tag:

```go
type MySecureStruct struct {
    ValueA     int      `json:",secure"`           // *will* be encrypted
    ValueB     struct{} `json:"value_b,secure"`    // *will* be encrypted
    ValueC     string   `json:",omitempty,secure"` // *will* be encrypted
    PlaintextA string                              // will *not* be encrypted
    PlaintextB int      `json:""`                  // will *not* be encrypted
    PlaintextC int      `json:"-"`                 // will *not* be encrypted
    privateA   int      `json:",secure"`           // will *not* be encrypted
}

```

Fields marked with `secure` are encrypted hierarchically, meaning if you have:

```go
package main

type MyStructA struct {
	ValueA string `json:"value_a,secure"`    // *will* be encrypted
}

type MyStructB struct {
	MyStructA                                // will *not* be encrypted
	ValueB    string `json:"value_b"`        // will *not* be encrypted
}

type MyStructC struct {
	MyStructA                                // will *not* be encrypted
	ValueC    string `json:"value_c"`        // will *not* be encrypted
}

type MyStructD struct {
	ValueD string    `json:"value_d,secure"` // *will* be encrypted
	Embed1 MyStructA                         // will *not* be encrypted
	Embed2 MyStructB                         // will *not* be encrypted
	Embed3 MyStructB `json:"embed_3,secure"` // *will* be encrypted
}

var myStruct = &MyStructD{
	ValueD: "foo",
	Embed1: MyStructA{
		ValueA: "bar",
	},
	Embed2: MyStructB{
		MyStructA: MyStructA{
			ValueA: "quack",
		},
		ValueB: "baz",
	},
	Embed3: MyStructB{
		MyStructA: MyStructA{
			ValueA: "foobar",
		},
		ValueB: "bonk",
	},
}
```

`myStruct` will be encrypted by Chestnut as:

```go
*MyStructD {
  ValueD: ****
  Embed1: main.MyStructA{
      ValueA: ****
  },
  Embed2: main.MyStructB{
      MyStructA: main.MyStructA{
      	ValueA: ****
      },
      ValueB: ****
  },
  Embed3: ****
}
```
where `'****'` represents an encrypted value.

Please see [Sparse Encryption](#sparse-encryption) for more information.

### Hash

When the `hash` option is added to a `` `json` `` struct field tag of a `string`
field, the `string` field is marked for hashing. If Chestnut detects a `hash` 
option on a `string` field, the string value of the field will be replaced with 
its hash.

If the `hash` option is applied to a struct field that is not type `string`, it 
is ignored.

To hash a string field of a struct, just add `hash` as an option to a `` `json`
`` struct field tag (like `omitempty`). The following are some examples of how
the `hash` option can be added to the `` `json` `` struct field tag:

```go
type MyHashStruct struct {
    ValueA     string   `json:",hash"`           // *will* be hashed
    ValueB     string   `json:"value_b,hash"`    // *will* be hashed
    ValueC     string   `json:",omitempty,hash"` // *will* be hashed
    ValueD     string   `json:",hash,omitempty"` // *will* be hashed
    ...
    Count      int      `json:"count,hash"`      // will *not* be hashed
}
```

Taking the above struct as an example:

```go
var myHashStruct = &MyHashStruct {
    ValueA: "value a",
    ValueB  "value b",
    ValueC  "value c",
    ValueD  "value d",
    ...
    Count   42,
}
```

`myHashStruct` will be encoded as:

```go
*main.MyHashStruct {
    ValueA: "sha256:BEBA1D9847D6E595D8DD6832DEE5432916C6F7AE438BC9A99C5BAFDD0E93793E"
    ValueB  "sha256:2A53D83488A34E898436908A7064276859FFF56F69D16E2F61573057EDEBFB64"
    ValueC  "sha256:80C92DE321A1CB8AEA3025890DE39A5BAA95A91DF022B10E1A71BEABB8BCC1BE"
    ValueD  "sha256:8080378350428FABDE1724D9B920D613B8920A71151F1A9AF37EB4AF43628AE4"
	...
    Count   42
}
```

#### SHA256

Chestnut currently supports SHA256 for hashing. In the future the `hash` option
may be extending to include an algorithm name e.g.:

```
`json:"some_field,hash=sha3-256"`
```

in which which case `hash` would continue to default to `sha256`.

#### Hash Prefix

Hashed struct fields will have the algorithm used to hash the value pre-pended
to the hash string:

```properties
sha256:BEBA1D9847D6E595D8DD6832DEE5432916C6F7AE438BC9A99C5BAFDD0E93793E
```

Chestnut uses the `[hash alogorithm name]:` prefix to know that it has *already* 
hashed the value, and it should hash it again when the struct is saved. 

**IMPORTANT!** Changing or removing the hash prefix will cause Chestnut to 
**rehash the value** of the struct field the next time the struct is saved.

### Multiple Tags

Chestnut supports the combining of tag options. You are free to mark a struct 
field as both `secure` and `hash`:

```go
type MyCombinedStruct struct {
    ValueA     string   `json:"value_a,secure,hash"` // will be hashed *AND* encrypted 
    ...
}
```
As with other tag options, the order in which they appear is unimportant. 

**However**, the order in which Chestnut **applies** them is **fixed**. A field
marked with both `secure` and `hash` will be **first** be hashed, and **then** 
encrypted. This order of operations cannot be changed for obvious reasons.

## Disable Overwrites

Chestnut supports the disabling of overwrites via the
`chestnut.OverwritesForbidden` option.

```go
cn := chestnut.NewChestnut(store, encryptor chestnut.OverwritesForbidden())
```
When this option is set, once a value has been saved to a namespaced key,
successive calls to save a value to the same key will fail with ErrForbidden.

The key must be explicitly deleted before a new call to save a value for the
same key will succeed.

## Keystore

Chestnut includes an implementation of IPFS compliant keystore which can be
found [here](keystore). 

### Importing Keystore

Using the Keystore is straight forward:

```go
package main

import (
	"github.com/jrapoport/chestnut"
	"github.com/jrapoport/chestnut/encryptor/aes"
	"github.com/jrapoport/chestnut/encryptor/crypto"
	"github.com/jrapoport/chestnut/keystore"
	"github.com/jrapoport/chestnut/storage/nuts"
)

// use nutsdb
store := nuts.NewStore(path)

// use a simple text secret
textSecret := crypto.TextSecret("i-am-a-good-secret")

// use AES256-CFB encryption
opt := chestnut.WithAES(crypto.Key256, aes.CFB, textSecret)

// open the keystore with nutsdb and the aes encryptor
ks := keystore.NewKeystore(store, opt)
if err := ks.Open(); err != nil {
    return err
}
```

A complete example of the Chestnut `Keystore` can be found
[here](examples/keystore).

### Important Note

```go
package main

import  (
    "github.com/ipfs/go-ipfs/keystore"
    "github.com/libp2p/go-libp2p/core/crypto"
)
```

If you want to work with the Keystore, please make **make sure** you are
importing 
[go-ipfs](github.com/ipfs/go-ipfs) and
[go-libp2p-core](https://github.com/libp2p/go-libp2p-core/), and NOT importing
[go-ipfs-keystore](github.com/ipfs/go-ipfs-keystore) and
[go-libp2p-crypto](github.com/libp2p/go-libp2p-crypto) — which are
**DEPRECATED**, out of date, a/o archived, etc.

## Logging

Chestnut supports logging via the `log.Logger` interface and the
`chestnut.WithLogger()` option. The `log.Logger` interface conforms to
[Logrus](https://github.com/Sirupsen/logrus),
[Zap](https://github.com/uber-go/zap), and the standard Go logger (with the
`log.Std` adapter), for example:

```go
opt := chestnut.WithLogger(myLogger)
```

#### Logrus Logger

Chestnut supports [Logrus](https://github.com/Sirupsen/logrus) logger.

`*logrus.Logger` and `*logrus.Entry` both work with Chestnut's `log.Logger` 
interface:

```go
logger := logrus.New() // *logrus.Logger

opt := chestnut.WithLogger(logger)
```

or 

```go
logger := logrus.New()
logger = logger.WithField("hello", "world") // *logrus.Entry

opt := chestnut.WithLogger(logger)
```

In addition to the `chestnut.WithLogger()` option, you can use the convenience 
option, `chestnut.WithLogrusLogger()`:

```go
opt := chestnut.WithLogrusLogger(log.InfoLevel)
``` 

`chestnut.WithLogrusLogger()` will return a new `*logrus.Entry` set to the
normalized log level you requested. This is equivalent to calling `logrus.New()`
followed by `logrus.SetLevel()`.

#### Zap Logger

Chestnut supports [Zap](https://github.com/uber-go/zap) logger.

`*zap.SugaredLogger` works with Chestnut's `log.Logger` interface:

```go
logger :=zap.NewProduction().Sugar() // *zap.SugaredLogger

opt := chestnut.WithLogger(logger)
```

In addition to the `chestnut.WithLogger()` option, you can use the convenience
option, `chestnut.WithZapLogger()`:

```go
opt := chestnut.WithZapLogger(log.InfoLevel)
```

`chestnut.WithZapLogger()` will return a new `*zap.SugaredLogger` set to the
normalized log level you requested. This is equivalent to calling 
`zap.NewProduction()`, followed by `zap.Core().Enabled()`, and finally, 
`zap.Sugar()`.

#### Standard Logger

Chestnut supports the [Go standard](https://golang.org/pkg/log/) logger.

We provide a lightweight wrapper for Go's standard logger `*log.Logger` which
supports Chestnut's `log.Logger` interface.

```go
import (
    "log"
    "os"

    "github.com/jrapoport/chestnut"
    cnlog "github.com/jrapoport/chestnut/log"
)

logger := cnlog.NewStdLogger(log.InfoLevel os.Stderr, "", log.LstdFlags)
opt := chestnut.WithLogger(logger)
```

`chestnut.WithStdLogger()` will return a new `*log.stdLogger` set to the
normalized log level you requested. This is equivalent to calling
`log.NewStdLogger()` with the specified log level.

#### Storage

Lastly, the Chestnut stores also accept matching options for logging: 
`storage.WithLogger`, `storage.WithLogrusLogger`,  `storage.WithZapLogger`, and
`storage.WithStdLogger`. 

```go
// enable logging 
opt := storage.WithLogger(myLogger)
// use nutsdb
store := nuts.NewStore(path, opt)
```

Enabling logging for the backing store was intentionally kept separate for 
additional flexibility. This allows you to log Chestnut operations without
automatically incurring the noise of store operations and vice versa.

## Examples

Run any example with `make <example-dir>`

```shell
$ make sparse
```

## Known Issues
* Because we use JSON encoding for structs, time.Time will lose resolution when 
  encoded:
  
  ````
  IN:  2021-01-09 17:29:36.349522 -0800 PST m=+0.002746093
  OUT: 2021-01-09 17:29:36.349522 -0800 PST

## Misc

### JSON encoding
We use the [jsoniter](https://github.com/json-iterator/go) JSON encoder 
internally. 
