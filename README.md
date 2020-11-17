# GoWebsocket
golang实现简单的websocket客户端

### 服务端
单独运行server.go作为服务端（ws://127.0.0.1:8080）去接收连接响应数据（copy他人代码，后期改掉哈哈哈）

### 客户端
全局连接websocket并发送订阅，全局接收数据做处理（简单粗暴）
- logger：日志保存至文件
- http：curl请求封装
- config：可配置路由
