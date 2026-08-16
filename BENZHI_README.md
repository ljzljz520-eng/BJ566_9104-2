# 食堂订餐单体系统

这是一个使用 Go 1.24.13 编写的单体演示项目。服务进程同时提供简单菜单页面和 JSON API，数据全部保存在进程内固定内存结构中，不需要数据库、网络服务、时钟或随机数。

## 运行

```bash
go run ./cmd/cafeteria
```

打开 `http://127.0.0.1:8080/` 可查看今日菜单。常用接口如下：

如端口已被占用，可执行 `go run ./cmd/cafeteria -addr 127.0.0.1:8081` 指定其他监听地址。

- `GET /api/menu`：今日菜单
- `POST /api/orders`：提交订餐，JSON 字段为 `request_id`、`user_id`、`menu_item_id`、`quantity`
- `GET /api/orders?user_id=student-001`：查询订餐记录
- `POST /api/orders/{order_id}/cancel?user_id=student-001`：取消订单
- `GET /api/balance?user_id=student-001`：查询余额

固定用户 `student-001` 初始余额为 100.00，菜单和库存也在每次启动时恢复为固定值。

## 业务链路测试

```bash
CGO_ENABLED=0 go test -count=1 ./...
```

测试覆盖菜单读取、订餐、查询、取消和余额返还。重复请求号的测试按照业务约定断言第二次提交应返回第一次结果且不产生额外副作用；当前实现保留了一个用于验收的幂等校验缺失，因此该测试会稳定失败并复现重复扣款、重复扣库存问题。
