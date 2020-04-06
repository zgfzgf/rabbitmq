# rabbitmq
[========]
### Engine queue
```seq
readermq->engine: Message(1)
Note right of readermq: store reader queue
engine->storemq: Message Log(N)
Note left of storemq: store transaction logs(commit)
engine->infomq: Message info(M)
Note left of infomq: store infos
```
[========]
### send->engine->recieve
```seq
send->engine: Message
Note right of send: store reader queue
engine->recieve: Message Log
Note right of engine: store store queue
```
### End
