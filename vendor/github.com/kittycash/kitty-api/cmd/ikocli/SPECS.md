# Commands

## Overview

`ikocli` is a way for the kc admin to submit kitties to `iko-chain` and `kitty-api`. Nodes will check submissions via given signature of the submissions. Submissions should be generated and signed offline.

### Offline commands

**Initiate file**

```bash
ikocli init --index-range=[2,32] --file=iko.json
```
Creates a `iko.json` file (as specified below). Essentially a list of `iko.Kitty` data types of the given index range.

**Edit file**

```bash
ikocli edit --field=breed --value=default --index-range=[2,6] --file=iko.json
```
This is a way of mass-editing the `iko.json`. The `index-range` flag can be a single number (e.g. `--index-range=5` should be interpreted as `--index-range=[5,5]`).

**Finalize file**

```bash
ikocli finalize --secret-key=39cc961167a27452d76cc7cfcfc3a97581a3c38f150d0a14597296996c5e1b46 --file=iko.json --output=out.bin
```

The format of `out.bin` is defined below. This is what is to be submitted to the servers directly.

## Format of iko.json

This is a iko.json file for 2 kitties (indexes 0 & 1).
It is an array of `iko.Kitty` elements.

```json
[
    {
        "kitty_id": 0,
        "name": "",
        "description": "",
        "breed": "",
        "price_btc": 0,
        "price_sky": 0,
        "box_open": false,
        "birth_date": 0,
        "kitty_dna": "",
        "box_image_url": "",
        "kitty_image_url": ""
    },
    {
        "kitty_id": 1,
        "name": "",
        "description": "",
        "breed": "",
        "price_btc": 0,
        "price_sky": 0,
        "box_open": false,
        "birth_date": 0,
        "kitty_dna": "",
        "box_image_url": "",
        "kitty_image_url": ""
    }
]
```

# Format of out.bin

This is made up of 2 parts:

- `Entries` is to be submitted to `kitty-api`.
    - Type: `database.Entry` of `"github.com/kittycash/kitty-api/src/database"`.
    - Remember to sign it using `func (e *Entry) Sign(sk cipher.SecKey)`.
    
- `Transactions` is to be submitted to `iko-chain`.
    - Should be generated via `func NewGenTx(kittyID KittyID, sk cipher.SecKey) *Transaction`.

and is encoded using `func Serialize(data interface{}) []byte` in `"github.com/skycoin/skycoin/src/cipher/encoder"`.

```go
type OutBin struct {
    Entries      []database.Entry
    Transactions []iko.Transaction 
}
```