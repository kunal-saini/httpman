# httpman

httpman is Lightweight and fast fluent interface wrapper over Http Client to make REST calls easier 

### Features ###

* Fluent request building, Add or Set Request Headers/Query
* Extendability to different endpoints
* Encode structs and key value pair into URL query
* Encode JSON payload into request
* Decode JSON success or failure responses
* Request retries [To Do]
* Fake HTTP responses for testing [To Do]
* File Message Store [To Do]

### Usage ###

```go
import "github.com/kunal-saini/httpman"
```

### Examples ###

```go
        resMap := make(map[string]interface{})
	errMap := make(map[string]interface{})

	type QueryParams struct {
		Foo string `url:"foo"`
	}

	client := httpman.
		New("https://example.com/").
		AddQueryStruct(&QueryParams{Foo: "bar"}).
		AddHeader("X-Key", "value")

	req, err := client.
		NewRequest().
		Get("path/to/resource").
		AddQueryParam("resource_key", "value").
		SetHeader("X-Resource-Key", "value").
		Decode(&resMap, &errMap)
```