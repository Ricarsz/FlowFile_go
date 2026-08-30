# goFileFlow — Go P2P 分布式文件存储

去中心化、内容寻址的 P2P 文件系统 Demo。节点间通过 TCP 直连，文件按 Hash 分目录落盘，支持多节点复制与流式传输。

## 架构
- **P2P**：TCPTransport / TCPPeer / RPC / Encoder·Decoder（Message/Stream 帧）
- **存储**：CASPathTransform (SHA1 → 多级目录) / Store.Write·Read·Has·Delete
- **服务**：FileServer (Store + Transport + peers) / Start / bootstrap / loop / Store→broadcast

## 快速开始
```bash
make run
go vet ./...
```

## 状态
- 已通：P2P 基础、Store 四件套、FileServer 跨节点复制（2节点 peers 1-1 验证）
- 待补：帧定界（Message/Stream + gob）、操作分发（Store/Get）、流控与大文件

