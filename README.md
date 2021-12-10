# pathenvconfig

A lot like https://github.com/kelseyhightower/envconfig, but supports  sourcing secrets from mounted files.

Instead of needing to specify a secret like `POSTGRES_PASSWORD` as an environment variable (which might be a [bad idea](https://blog.nillsf.com/index.php/2020/02/24/dont-use-environment-variables-in-kubernetes-to-consume-secrets/)), you can specify `POSTGRES_PASSWORD_FILE=/path/to/secret.txt`, where `/path/to/secret.txt` contains the password.

## Usage

Define a struct to hold your configuration values:

```go
type MyConfig struct {
	DatabaseConnectionString string `required:"true"`
	TimeoutSeconds           int    `default:"10"`
}
```

Then you can load values into an instance of this struct my calling:

```go
config := MyConfig{}
pathenvconfig.Process("MY_APP", &config)
```

This will populate the fields of the struct by inspecting environment variables. The first field can be specified with the environment variable

```bash
export MY_APP_DATABASE_CONNECTION_STRING="user=someone password=xyz dbname=mydb host=postgres port=5432"
```

Or you can provide an environment variable with the suffix `_FILE` that contains a path to a file that contains the actual connection string:

```bash
export MY_APP_DATABASE_CONNECTION_STRING_FILE="/path/to/a/file.txt"
```

## Limitations

The implementation is currenly very basic. Field types can only be primitives.
