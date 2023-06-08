# Self-hosted Brave sync server

A simplified version of the [Brave sync server](https://github.com/brave/go-sync),
made more suitable for self-hosting by replacing the AWS Dynamo and Redis
dependencies with SQLite3 and a local memory cache.

```
brave-browser --sync-url=http://localhost:8295/litesync
```
