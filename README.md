# generate

A library to generate go models from given json files

![A test image](https://miro.medium.com/max/256/0*3CCVZH7RBPWlBVFA.png)


# Requirements

* Go 1.17+

# Usage

Install

```console
$ go get -u github.com/azarc-io/json-schema-to-go-struct-generator
```

or



# Example

This schema

```json
{
  "$schema": "http://json-schema.org/draft-04/schema#",
  "title": "Example",
  "id": "http://example.com/exampleschema.json",
  "type": "object",
  "description": "An example JSON Schema",
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
  },
  "definitions": {
    "address": {
      "id": "address",
      "type": "object",
      "description": "Address",
      "properties": {
        "street": {
          "type": "string",
          "description": "Address 1",
          "maxLength": 40
        },
        "houseNumber": {
          "type": "integer",
          "description": "House Number"
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
            "C"
          ],
          "type": "string",
          "description": "The favourite cat.",
          "maxLength": 1
        }
      }
    }
  }
}
```

generates

```go
package main

type Address struct {
  HouseNumber int `json:"houseNumber,omitempty"`
  Street string `json:"street,omitempty"`
}

type Example struct {
  Address *Address `json:"address,omitempty"`
  Name string `json:"name,omitempty"`
  Status *Status `json:"status,omitempty"`
}

type Status struct {
  Favouritecat string `json:"favouritecat,omitempty"`
}
```

See the [test/](./test/) directory for more examples.

# Running Tests

In order to run the tests, you must first generate the sample outputs which will produce
the code required to compile and run the tests.

```shell
go generate github.com/azarc-io/json-schema-to-go-struct-generator/test
go test ./test/...
```
