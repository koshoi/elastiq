# elastiq
Query Elasticsearch without Kibana

## Installation

Build the binary from source code
```bash
$ go build -o elastiq ./main.go
```

Copy built binary to some location in your **\$PATH**

```bash
$ cp elastiq /usr/bin/elastiq
```

## Configuration

By default the config file is **\$HOME/.config/elastiq/config.toml**

Config file location can be specified using **-c** command line flag

An example config is presented in file elastiq.toml

```toml
[env.dev]
endpoints = ["http://localhost:9200"]
default   = true
index     = "dev-*"
output    = "pretty"

[env.prod]
endpoints = ["http://localhost:9299"]
default   = false
index     = "prod-*"
output    = "pretty"

[output.pretty]
exclude = ["kubernetes", "message", "@timestamp"]
format = "json"
decode_recursively = true

[output.message]
only = ["message"]
format = "json"

[output.all]
format = "json"
```

The config file lists **environments** and **outputs**

### Environment

Environment specifies endpoints and credentials to your elasticsearch.
It also contains
- default index to use when querying elastic (you can change index using **-i** flag from command line)
- default output to use (you can change ouput using **-o** flag from command line)

To specify the env to use when querying elastic you can use **-e** flag from command line.
However you can set one environment to de default.

In given example environment 'dev' is used by default.

### Output

Output is a small config that changes how records from elasticsearch are printed.

It contains
- **format** (by now it only can do json)
- **exclude** list of top-level fields to delete from final output
- **only** list of top-level fields to output
- **decode_recursively** specify if you want your data to be recursively decoded

**decode_recursively** can be either boolean (if true, it will try to decode using every known decoder)
or a list of strings (list of decoders to use).

For now there are 3 implemented decoders
- json
- http

Using all of them will change record like this
```json
{
	"somehttp": "GET /health HTTP/1.1\r\nHost: localhost:80\r\nContent-Length: 0\r\n\r\n",
	"somejson": "{\"a\":\"b\",\"c\":[1,2,3]}",
}
```

to this
```json
{
	"somehttp": {
		"body": {},
			"headers": {
				"Content-Length": "0",
				"Host": "localhost:80",
			},
			"method": "GET",
			"url": "/health",
			"version": "HTTP/1.1"
	},
	"somejson":{
		"a":"b",
		"c":[1, 2, 3]
	}
}
```

You can change recursive decoding behavior from command line using **-R** argument.
It takes coma separated list of decoders to use (-R json for example)

## Usage

elastiq is command-based tool, but for now only **query** command is implemented (with an alias **q**)

Here are some examples
```bash
$ elastiq query -f level=info -f kubernetes.labels.environment=stage -t -1h/now --limit 1
$ elastiq query -e prod -f level=error -f kubernetes.labels.environment=stage -t -1h/now --limit 1
$ elastiq query -f level=error -f 'request_id in qwe asd zxc' -t -1h/now --limit 100
$ elastiq query -f level=error -f 'http.status_code between 400 500' -t -1h/now --limit 100
```
