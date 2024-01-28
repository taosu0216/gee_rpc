# 构建客户端
## Call
```text
the method’s type is exported.
the method is exported.
the method has two arguments, both exported (or builtin) types.
the method’s second argument is a pointer.
the method has return type error.
```
## 注册方法
在 Go 语言的 RPC 系统中，`Client` 结构体是客户端的核心，它负责管理和发送所有的 RPC 请求。在 `Client` 结构体中，有两个互斥锁：`sending` 和 `mu`，以及一个 `Call` 的映射 `pending`。

`sending` 是一个互斥锁，主要用于保护请求的发送过程。在一个 RPC 系统中，可能会有多个请求同时发送，如果不加锁，可能会导致请求的数据混乱。因此，我们需要一个互斥锁来确保一次只能有一个请求在发送。这就是 `sending` 锁的作用。

`mu` 也是一个互斥锁，主要用于保护 `Client` 结构体中的状态和数据。在 `Client` 结构体中，有一些字段（如 `seq`，`pending`，`closing` 和 `shutdown`）可能会在多个 goroutine 中同时被访问和修改，这可能会导致数据竞争的问题。为了避免这种情况，我们需要一个互斥锁来确保在修改这些字段时不会被其他 goroutine 干扰。这就是 `mu` 锁的作用。

`pending` 是一个映射，其键是 `Call` 的 `Seq` 字段，值是 `Call` 对象本身。在 RPC 系统中，每个请求都会被封装成一个 `Call` 对象，然后通过 `Seq` 字段来唯一标识。当我们发送一个请求时，会先将这个 `Call` 对象注册到 `pending` 映射中，然后再发送请求。当我们接收到一个响应时，会通过 `Seq` 字段来查找对应的 `Call` 对象，然后处理响应。这就是 `pending` 映射的作用。

在 `Client` 结构体中，我们是将 `Call` 注册到 `Client` 中，而不是将 `Client` 注册到 `Call` 中。这是通过 `registerCall` 方法实现的。这个方法首先会锁定 `mu` 锁，然后检查 `Client` 是否正在关闭或已经关闭。如果是，那么返回一个错误。否则，将 `Call` 的 `Seq` 设置为 `Client` 的 `seq`，然后将 `Call` 添加到 `Client` 的 `pending` 映射中，并将 `Client` 的 `seq` 加一。最后，返回 `Call` 的 `Seq` 和 `nil` 错误。这就是 `registerCall` 方法的工作流程。

## removeCall
removeCall 方法返回一个 Call 对象的原因是为了在移除 Call 对象后，还能继续对这个 Call 对象进行操作。  在 Client 结构体中，pending 字段是一个映射，用于存储所有未完成的 Call 对象。当一个 Call 对象的 RPC 请求完成时，我们需要从 pending 映射中移除这个 Call 对象。这就是 removeCall 方法的主要功能。  然而，虽然 Call 对象的 RPC 请求已经完成，我们可能还需要对这个 Call 对象进行一些后续操作，比如检查它的 Error 字段是否有错误，或者处理它的 Reply 字段中的响应数据。因此，removeCall 方法在移除 Call 对象后，还需要返回这个 Call 对象，以便进行后续操作。


客户端向服务端发送一个请求,返回的响应的seq和发送的请求的seq是相同
流程是客户端构建请求,然后把这个请求放到一个pending的map中,当有对应seq号的响应发回,则把这个请求从map中移除
## terminateCalls
terminateCalls 是一个方法，它的作用是在客户端关闭或出现错误时，终止所有正在进行的调用，并将错误信息设置到每个调用的 Error 字段。  这个方法首先获取 sending 和 mu 两个互斥锁，然后将 shutdown 字段设置为 true，表示客户端已经关闭。然后，它遍历 pending 映射中的所有 Call 对象，将错误信息设置到 Call 的 Error 字段，并调用 done 方法来通知调用已经完成。

## 三大要素
创建连接
- 交换option
  - MagicNumber
  - CodecType
- 
发送请求
接收响应