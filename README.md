# zkproxy
a HTTP+JSON proxy of ZooKeeper.

## Install
```bash
$ go get github.com/xgfone/zkproxy
$ cd $GOPATH/src/github.com/xgfone/zkproxy
$ dep ensure && go install github.com/xgfone/zkproxy
```

Notice:
1. The package manager is [dep](https://github.com/golang/dep), which will be added into the go tool chain in future.
2. The go version should be 1.7+.

## Example
```bash
$ curl http://127.0.0.1/zk -X POST -H "application/json" -d '{"cmd":"create", "path":"/test", "data":"test"}'
```
Output:
```json
{"path": "/test"}
```

## HTTP API

### Method
You can use any HTTP method: `GET`, `POST`, `PUT`, `OPTION`, even `DELETE`. But suggest to use `POST`.

### URL
The request url is `http(s)://host[:port]/zk`.

### The Request Body
The request body is JSON, which must contain the field `cmd`.

### The Response Status Code and Body.
- 200: OK. If there is the body, it's JSON.
- 400: The request body arguments is wrong.
- 404: The path/node does not exist.
- 406: The path/node has existed, or the version argument is not consistent with the server..
- 500: The server error or other errors.
- 501: The CMD is not implementation.

**Notice:** If the status code is not 200, the body is an error string.

#### 1. AddAuthInfo
**Request Body**

|  Field  |  Type  | Required | Value
|---------|--------|----------|----------------
| cmd     | string | Y        | `add_auth_info`
| scheme  | string | Y        |
| auth    | string | Y        |

**Response Body**

None.

#### 2. Create
**Request Body**

|    Field   |  Type  | Required | Value
|------------|--------|----------|---------
| cmd        | string | Y        | `create`
| path       | string | Y        |
| data       | string | Y        |
| ephemeral  | bool   | N        |
| sequential | bool   | N        |
| acl        | array  | N        |

The format of each element of the array `acl` is JSON:

| Field  |  Type  | Required
|--------|--------|---------
| id     | string | Y
| scheme | string | Y
| perms  | int    | Y

**Response Body**

| Field | Type
|-------|--------
| path  | string

#### 3. Exists
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|----------|----------
| cmd   | string | Y        | `exists`
| path  | string | Y        |

**Response Body**

If the path exists, the JSON is

| Field | Type | Value
|-------|------|--------
| exist | bool | `true`
| czxid | int |
| mzxid | int |
| pzxid | int |
| ctime | int |
| mtime | int |
| version  | int |
| cversion | int |
| aversion | int |
| data_length  | int |
| num_children | int |
| ephemeral_owner | int |

If the path does not exist, the JSON is

| Field | Type | Value
|-------|------|-------
| exist | bool | `false`

#### 4. GetChildren
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|----------|---------
| cmd   | string | Y        | `get_children`
| path  | string | Y        |

**Response Body**

| Field | Type
|-------|------
| children | array\<string>
| czxid | int
| mzxid | int
| pzxid | int
| ctime | int
| mtime | int
| version  | int
| cversion | int
| aversion | int
| data_length  | int
| num_children | int
| ephemeral_owner | int

#### 5. GetData
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|----------|---------
| cmd   | string | Y        | `get_data`
| path  | string | Y        |

**Response Body**

| Field | Type
|-------|------
| data | string
| czxid | int
| mzxid | int
| pzxid | int
| ctime | int
| mtime | int
| version  | int
| cversion | int
| aversion | int
| data_length  | int
| num_children | int
| ephemeral_owner | int

#### 6. SetData
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|--------- |----------
| cmd   | string | Y        | `set_data`
| path  | string | Y        |
| data  | string | Y        |
| version | int  | Y        |

**Response Body**

| Field | Type
|-------|------
| czxid | int
| mzxid | int
| pzxid | int
| ctime | int
| mtime | int
| version  | int
| cversion | int
| aversion | int
| data_length  | int
| num_children | int
| ephemeral_owner | int

#### 7. GetACL
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|----------|----------
| cmd   | string | Y        | `get_acl`
| path  | string | Y        |

**Response Body**

| Feild | Type
|-------|------
| acl   | array
| czxid | int
| mzxid | int
| pzxid | int
| ctime | int
| mtime | int
| version  | int
| cversion | int
| aversion | int
| data_length  | int
| num_children | int
| ephemeral_owner | int

The format of each element of the array `acl` is JSON:

| Field  |  Type
|--------|--------
| id     | string
| scheme | string
| perms  | int

#### 8. SetACL
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|--------- |----------
| cmd   | string | Y        | `set_acl`
| path  | string | Y        |
| version | int  | Y        |
| acl   | array  | Y        |

The format of each element of the array `acl` is JSON:

| Field  |  Type  | Required
|--------|--------|---------
| id     | string | Y
| scheme | string | Y
| perms  | int    | Y

**Response Body**

| Field | Type
|-------|------
| czxid | int
| mzxid | int
| pzxid | int
| ctime | int
| mtime | int
| version  | int
| cversion | int
| aversion | int
| data_length  | int
| num_children | int
| ephemeral_owner | int

#### 9. Delete
**Request Body**

| Field |  Type  | Required | Value
|-------|--------|----------|----------
| cmd   | string | Y        | `delete`
| path  | string | Y        |
| version | int  | Y        |

**Response Body**
None.

**Notice:** The watch api of `Exists`, `GetData`, `GetChildren`, and `Multi` are not implemented.
