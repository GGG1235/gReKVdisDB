# gReKVdisDB
golang实现基于Redis的key-value内存数据库

---
```
.
├── LICENSE
├── README.md
├── aof
│   ├── aof.go
│   └── gkvdb.aof
├── core
│   ├── client.go
│   ├── command.go
│   ├── core.go
│   ├── date_structure
│   │   └── lists
│   │       ├── list.go
│   │       └── node.go
│   ├── db.go
│   ├── pubsub.go
│   ├── server.go
│   ├── zhash.go
│   └── zset.go
├── gkvdb-cli
├── gkvdb-cli.go
├── gkvdb-server
├── gkvdb-server.go
├── go.mod
├── proto
│   ├── decoder.go
│   ├── encoder.go
│   ├── proto.go
│   └── resp.go
└── utils
    ├── error.go
    ├── gbufio
    │   ├── reader.go
    │   └── writer.go
    ├── obj.go
    └── version.go
```