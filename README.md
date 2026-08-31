# goFileFlow — Go P2P 分布式文件存储

去中心化、内容寻址的 P2P 文件系统 **Demo**（简化版 IPFS）。节点间通过 TCP 直连，文件按 Hash 分目录落盘，支持多节点副本复制与流式传输。

## 功能特性

- **P2P 节点互联**：TCP 直连，自定义帧协议（`IncomingMessage` / `IncomingStream` 双类型帧），支持 peer list 自动扩散
- **内容寻址存储**：SHA1 → 两级目录（`CASPathTransform`），同一内容的 key 恒定
- **多副本复制**：`Store` 时按一致性哈希挑选节点推流式传输副本（默认 3 副本）
- **广播拉取**：`Get` 时本地无文件则向全网广播，peer 回传数据后落盘
- **HTTP API**：基于 Echo 的 REST 接口（存 / 取 / 删）

## 架构分层

```
handler/          HTTP 层（echo）
  echo.go         REST API：POST/GET/DELETE /store/:key
server.go         FileServer：peer 管理、消息循环、副本选择、pendingGets
store.go          CAS 存储：路径变换、Write/Read/Has/Delete（RWMutex 保护）
p2p/              传输层
  tcp_transport.go  TCPTransport / TCPPeer（含跨消息复用的 bufio.Reader）
  encoding.go       DefaultEncoder / DefaultDecoder（帧 + gob）
  message.go        消息类型：MessagePeerList / MessageStoreFile / MessageGetFile
  transport.go      Peer / Transport / RPC 接口定义
main.go           入口：p2p :3000，HTTP :8080
```

## 使用

```bash
make build          # 编译到 bin/fs
make run            # 启动单节点（p2p :3000，HTTP :8080）
make test           # 运行全部测试
```

多节点：启动多个实例，在 `main.go` 的 `BootstrapNodes` 中填入已有节点地址即可互联。

### HTTP API

| 方法 | 路径 | 说明 |
|---|---|---|
| `POST` | `/store/:key` | 存储文件（body 为文件内容），并按副本策略复制到 peer |
| `GET` | `/store/:key` | 取文件；本地没有则向 peer 广播拉取（约 2s 超时），成功返回 `application/octet-stream` |
| `DELETE` | `/store/:key` | 删除文件（含空目录清理） |

## 测试

```bash
go test ./... -v
```

覆盖：存储 CRUD（`store_test.go`）、帧编解码与粘包（`p2p/encoding_test.go`）、
stream 流式传输回归（`p2p/tcp_smoke_test.go`）、双节点副本推送 + 广播拉取全链路
（`server_smoke_test.go`）。

## 已知边界

- **Demo 级别**：无节点发现协议（需 bootstrap 地址）、无加密、无 NAT 穿透、无故障转移
- `FileServer.Get` 使用 2s 超时 + channel 完成信号；超时按"未找到"处理（HTTP 404）
- `TCPTransport.Addr()` 返回配置地址而非实际监听地址（`:0` 随机端口场景不适用）
- 流式传输依赖发送方按 `hdr.Size` 精确发送数据

## 备注

- 兼容 Go 1.27：`gob.NewDecoder` 会对非 `io.ByteReader` 的 reader 做 bufio 预读，
  本项目通过 `TCPPeer` 持有跨消息复用的 `bufio.Reader`（`Peer.Reader()`）规避，
  解码与 stream 数据读取共用同一缓冲，避免预读字节随 decoder 丢弃。
