# Archived
本节为该项目存档说明和总结。

## 项目目标与实现情况
项目最初的目标是编写一个简单的桌面端软件，为`导出国服Build数据到国际服POB提供支持`，包括：

- 网关，接受浏览器插件的websocket连接请求，接受POB的http请求，并将http请求通过websocket连接委托给前端插件处理，然后返回结果给POB
- 配置管理
- 日志记录
- 桌面UI

除了桌面UI部分，其余部分需求都完成了。

最后使用C#编写了一个UI壳，而本项目成为了一个后台程序，两者使用`config`文件和Windows进程管理来实现简单的IPC通信。UI壳负责config的配置、backend的启动和关闭；backend读取config并提供网关服务。

## 桌面UI
尝试过使用`walk`（见ui分支）来实现UI，但由于该UI框架功能不够完善，BUG较多，最终放弃了。

我从中学习到的经验是，除非项目需求你硬着头皮必须使用相应的技术来做UI，又或者你是桌面端UI的老手，那么尽量避免使用不成熟的方案。

扩展开来，对于一个项目，技术和实现你必须要有一个是很熟悉的。如果你对技术很熟悉，那么你可以用旧的工具来完成新的项目；如果你对实现很熟悉，那么你可以通过用新的工具来重写旧的项目，从而学习新的工具。

## 网关服务
网关服务是本项目的功能核心，在理解`websocket`协议的基础上，构建一个`websocket`连接池，并分派任务。

得益于`gorilla/websocket`库清晰的示例代码，以及`go`简洁的`协程`语法，这部分实现起来很顺手。

由于时间关系，细节难以在此叙述，请重看代码。