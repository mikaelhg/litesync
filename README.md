# Self-hosted Brave sync server

*On a hiatus...* waiting for the Android Brave browser to gain the feature of
setting a custom sync server URL, without which continuing is pretty pointless.

A simplified version of the [Brave sync server](https://github.com/brave/go-sync),
made more suitable for self-hosting by replacing the AWS Dynamo and Redis
dependencies with SQLite3 and a local memory cache.

```
brave-browser --sync-url=http://localhost:8295/litesync
```
