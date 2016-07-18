# generate

Generates Go (golang) Structs from JSON schema.

# Examples

```
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Example",
  "id": "Example",
  "type": "object",
  "description": "example",
  "definitions": {
    "address": {
      "id": "address",
      "type": "object",
      "description": "Address",
      "properties": {
        "houseName": {
          "type": "string",
          "description": "House Name",
          "maxLength": 30
        },
        "houseNumber": {
          "type": "string",
          "description": "House Number",
          "maxLength": 4
        },
        "flatNumber": {
          "type": "string",
          "description": "Flat",
          "maxLength": 15
        },
        "street": {
          "type": "string",
          "description": "Address 1",
          "maxLength": 40
        },
        "district": {
          "type": "string",
          "description": "Address 2",
          "maxLength": 30
        },
        "town": {
          "type": "string",
          "description": "City",
          "maxLength": 20
        },
        "county": {
          "type": "string",
          "description": "County",
          "maxLength": 20
        },
        "postcode": {
          "type": "string",
          "description": "Postcode",
          "maxLength": 8
        }
      }
    },
    "status": {
      "type": "object",
      "properties": {
        "favouritecat": {
          "enum": [
            "A",
            "B",
            "C",
            "D",
            "E",
            "F"
          ],
          "type": "string",
          "description": "The favourite cat.",
          "maxLength": 1
        }
      }
    }
  },
  "properties": {
    "name": {
      "type": "string"
    },
    "address": {
      "$ref": "#/definitions/address"
    },
    "status": {
      "$ref": "#/definitions/status"
    }
  }
}
```

Run this:

```
go run main.go -i exampleschema.json
```

Get this:

```
package main

type Address struct {
  County string `json:"county"`
  District string `json:"district"`
  FlatNumber string `json:"flatNumber"`
  HouseName string `json:"houseName"`
  HouseNumber string `json:"houseNumber"`
  Postcode string `json:"postcode"`
  Street string `json:"street"`
  Town string `json:"town"`
}

type Example struct {
  Address Address `json:"address"`
  Name string `json:"name"`
  Status Status `json:"status"`
}

type Status struct {
  Favouritecat string `json:"favouritecat"`
}
```

