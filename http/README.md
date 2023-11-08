# yazl/utils/http

## Example

### Method Get
```go
header := http.DefaultJsonHeader()
params := http.NewParams()

resp := http.Default().Get(ctx, url, params, header)
```

### Method Post
```go
header := http.DefaultFormUrlEncodedHeader()
form := http.NewForm().Add("name", "superwhys").Encode()

resp := http.Default().Post(ctx, url, nil, header, form)
```

### To get string resp
```go
respStr, err := resp.BodyString()
```

### To get Bytes resp
```go
respBytes, err := resp.BodyBytes()
```

### To get json resp
```go
err := resp.BodyJson(&respStruct)
```

### Headers
#### creat a new header
`http.NewHeader()`
#### Add value to header
`header.Add(key, value)`

#### It also has a number of different headers built in
`http.DefaultJsonHeader()`
`http.DefaultFormUrlEncodedHeader()`
`http.DefaultFormHeader()`

### Params
#### creat a new params
`http.NewParams()`
#### Add value to params
`params.Add(key, value)`
#### Get value from params
`params.Get(key)`

### Form
#### creat a new form
`http.NewForm()`
#### Add value to form
`form.Add(key, value)`
#### Encode form
`form.Encode()`
