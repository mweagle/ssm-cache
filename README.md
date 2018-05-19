# ssm-cache

An in memory cache that provides typed accessors to
the [SSM Parameter Store](https://docs.aws.amazon.com/systems-manager/latest/userguide/systems-manager-paramstore.html), inspired by [ssm-cache-python](https://github.com/alexcasalboni/ssm-cache-python).

# Usage

## Create a Client

Create a single ssmcache.Client instance for your
lambda application. This client should only
be initialized once as part of an `init()` function.

The single argument is the default expiry for cached
values. To create a cache whose elements do not expire by
default, provide the [NoExpiration](https://godoc.org/github.com/patrickmn/go-cache#pkg-constants) value.

```go

var cacheClient ssmcache.Client
func init() {
	cacheClient = ssmcache.NewClient(5 * time.Minute)
}
```

## Fetch Value

### Default Expiry

```go
  stringVal, stringValErr := cacheClient.GetString("MyParam")
  stringSliceVal, stringSliceValErr := cacheClient.GetStringList("MyParam")
  decryptedStringVal, decryptedStringValErr := cacheClient.GetSecureString("MyParam")
```

### Custom Expiry

```go
  stringVal, stringValErr := cacheClient.GetExpiringString("MyParam", 30*time.Second)
  stringSliceVal, stringSliceValErr := cacheClient.GetExpiringStringList("MyParam", 30*time.Second)
  decryptedStringVal, decryptedStringValErr := cacheClient.GetExpiringSecureString("MyParam", 30*time.Second)
```

## Force Refresh

Use `Purge(keyname)` to force delete a cached entry and reload the value from SSM.

```go
  stringVal, stringValErr := cacheClient.Purge("MyParam").GetString("MyParam")
```

## Fetch Group

### Default Expiry

```go
  paramMap, paramMapErr := cacheClient.GetParameterGroup("StorageKey", "/my/custom/ssm-path")
```

### Custom Expiry

```go
  paramMap, paramMapErr := cacheClient.GetExpiringParameterGroup("StorageKey", "/my/custom/ssm-path")
```

## References

  * [Sharing Secrets with AWS Lambda Using AWS Systems Manager Parameter Store](https://aws.amazon.com/blogs/compute/sharing-secrets-with-aws-lambda-using-aws-systems-manager-parameter-store/)