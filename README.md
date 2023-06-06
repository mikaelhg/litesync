# Self-hosted Brave sync server

A simplfied version of the [Brave sync server](https://github.com/brave/go-sync),
made more suitable for self-hosting by replacing the AWS Dynamo and Redis
dependencies with SQLite3 and a local memory cache.

```
--sync-url=https://sync-v2.brave.com/v2
```
