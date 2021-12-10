# pathenvconfig

A lot like https://github.com/kelseyhightower/envconfig, but supports  sourcing secrets from mounted files.

Instead of needing to specify a secret like `POSTGRES_PASSWORD` as an environment variable (which might be a [bad idea](https://blog.nillsf.com/index.php/2020/02/24/dont-use-environment-variables-in-kubernetes-to-consume-secrets/)), you can specify `POSTGRES_PASSWORD_FILE=/path/to/secret.txt`, where `/path/to/secret.txt` contains the password.
