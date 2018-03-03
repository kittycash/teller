# Kittycash Wallet

Where one claims ownership of dem' kitties.

## Run IKO node (test mode)

Run a node with test data and 10 kitties.

```bash
iko \
    -master-public-key 03429869e7e018840dbf5f94369fa6f2ee4b380745a722a84171757a25ac1bb753 \
    -memory \
    -test \
    -test-secret-key 190030fed87872ff67015974d4c1432910724d0c0d4bfbd29d3b593dba936155 \
    -test-injection-count 10
```

RESTful API will be served on port `:8080`.

**Get Kitty of ID:**

Request (for JSON reply):

```text
GET http://127.0.0.1:8080/api/iko/kitty/9.json
```

Response:

```json
{
    "kitty_id": 9,
    "address": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
    "transactions": [
        "40c34bc724643d5b25beea3fdb3b1eeeff61b08b6ba90111126d2571f28aa33a"
    ]
}
```

Request (for encoded reply):

```text
GET http://127.0.0.1:8080/api/iko/kitty/9.enc
```

**Get Address:**

Request (for JSON reply):

```text
GET http://127.0.0.1:8080/api/iko/address/2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7.json
```

Response:

```json
{
    "address": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
    "kitties": [
        0,
        1,
        2,
        3,
        4,
        5,
        6,
        7,
        8,
        9
    ],
    "transactions": [
        "cd7073ed8dc93c3e0d52ab3925887161ff3063e56a95a5503d56b4726b910080",
        "87f006702875485c11bfbfc124915e11d3794a50c458d9a49bd7c2004b383914",
        "c01a57201bc5d0a93d61bf3ca07d110402de8f7acbbff38f40038d501c8f6e45",
        "11446e215bb4bcc39fecf238e850e61fdd491108feef5e9f4bcb7bdc21dc163b",
        "98803f1872ad2bee2bc9e907788231ce0409ea234852dcf46e7f54f5fcd8cfb1",
        "a39e5d3bb3b7da93c10aa29c5c71f8cf84a1ca7adc410c26c6e422aacf069af8",
        "f1003dc6adadd98ab9dac25c836530c613d374862b6efbd16df93ca9aa65c03b",
        "1f78bddf95fd20ec9fd44a0f5ac1795cfa65243dfb5adad9b406f6410cd8e855",
        "c18e2c0421ec6f2b8ea06472d333cd499230a1e6599be960cfb5190d3cfb6d37",
        "40c34bc724643d5b25beea3fdb3b1eeeff61b08b6ba90111126d2571f28aa33a"
    ]
}
```

Request (for encoded reply):

```text
GET http://127.0.0.1:8080/api/iko/address/2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7.enc
```

**Get Transaction of Hash:**

Request (for JSON reply):

```text
GET http://127.0.0.1:8080/api/iko/tx/72e9b929f77d35cd556c4fe3d758d537b72537330790e186b842786da6d8f3cc.json?request=hash
```

Response:

```json
{
    "meta": {
        "hash": "72e9b929f77d35cd556c4fe3d758d537b72537330790e186b842786da6d8f3cc",
        "raw": "3815752563947ba5342fefa059479d476a2586a5544574bd9605c0135bbc483208000000000000004fd8ed48b29c16150800000000000000000427fcd0f0b9461c5c516cd66a4b5ac413978272000427fcd0f0b9461c5c516cd66a4b5ac413978272408980e7c3671fcd3fc7c6258d3de8b4ad477323456850080ff603578f72f99000f712a195bd77393de32be08125436a9d02553448b2e3d3b43dee96dd6a6e7a00"
    },
    "transaction": {
        "prev_hash": "3815752563947ba5342fefa059479d476a2586a5544574bd9605c0135bbc4832",
        "seq": 8,
        "time": 1519574213825779791,
        "kitty_id": 8,
        "from": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
        "to": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
        "sig": "408980e7c3671fcd3fc7c6258d3de8b4ad477323456850080ff603578f72f99000f712a195bd77393de32be08125436a9d02553448b2e3d3b43dee96dd6a6e7a00"
    }
}
```

Request (for encoded reply):

```text
GET http://127.0.0.1:8080/api/iko/tx/72e9b929f77d35cd556c4fe3d758d537b72537330790e186b842786da6d8f3cc.enc?request=hash
```

**Get Transaction of Sequence:**

Request (for JSON reply):

```text
GET http://127.0.0.1:8080/api/iko/tx/7.json?request=seq
```

Response:

```json
{
    "meta": {
        "hash": "1f78bddf95fd20ec9fd44a0f5ac1795cfa65243dfb5adad9b406f6410cd8e855",
        "raw": "f1003dc6adadd98ab9dac25c836530c613d374862b6efbd16df93ca9aa65c03b0700000000000000507b6202a19f16150700000000000000000427fcd0f0b9461c5c516cd66a4b5ac413978272000427fcd0f0b9461c5c516cd66a4b5ac413978272f9baf19ce3aed213a3008891462107299947dd8e32f077f2396b2d7e81e8562a55f4a9506176219b58646dc6387f81298dd4b23b891e06eb83114ab62eb3f84f00"
    },
    "transaction": {
        "prev_hash": "f1003dc6adadd98ab9dac25c836530c613d374862b6efbd16df93ca9aa65c03b",
        "seq": 7,
        "time": 1519577438162680656,
        "kitty_id": 7,
        "from": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
        "to": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
        "sig": "f9baf19ce3aed213a3008891462107299947dd8e32f077f2396b2d7e81e8562a55f4a9506176219b58646dc6387f81298dd4b23b891e06eb83114ab62eb3f84f00"
    }
}
```

Request (for encoded reply):

```text
GET http://127.0.0.1:8080/api/iko/tx/7.enc?request=seq
```

**Get Head Transaction**

Request (for JSON reply):

```text
GET http://127.0.0.1:8080/api/iko/head_tx.json
```

Response:

```json
{
    "meta": {
        "hash": "40c34bc724643d5b25beea3fdb3b1eeeff61b08b6ba90111126d2571f28aa33a",
        "raw": "c18e2c0421ec6f2b8ea06472d333cd499230a1e6599be960cfb5190d3cfb6d3709000000000000007dafaa02a19f16150900000000000000000427fcd0f0b9461c5c516cd66a4b5ac413978272000427fcd0f0b9461c5c516cd66a4b5ac4139782723bef43f3d326265978014af2589bca4bde89684683dcf85e13f6f118ac5913ec6b45db49462e94eec00fd1bcdbbe48638533a58042cc3c07f17ede877ebb4fa000"
    },
    "transaction": {
        "prev_hash": "c18e2c0421ec6f2b8ea06472d333cd499230a1e6599be960cfb5190d3cfb6d37",
        "seq": 9,
        "time": 1519577438167412605,
        "kitty_id": 9,
        "from": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
        "to": "2fzr9thfdgHCWe8Hp9btr3nNEVTaAmkDk7",
        "sig": "3bef43f3d326265978014af2589bca4bde89684683dcf85e13f6f118ac5913ec6b45db49462e94eec00fd1bcdbbe48638533a58042cc3c07f17ede877ebb4fa000"
    }
}
```

Request (for encoded response):

```text
GET http://127.0.0.1:8080/api/iko/head_tx.enc
```

**Inject Transaction**

Request:

```text
POST http://127.0.0.1:8080/api/iko/inject_tx
Content-Type: application/json or application/octet-stream
```
