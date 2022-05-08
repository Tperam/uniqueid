# README

当前项目用于生成唯一ID



#### 原理

由于snowflake依赖机器的时间戳。对于我们当前运维要求较高，具体维护过于繁琐。并且其实现对于部分语言不友好（javascript只能存储53bit的数字类型），所以此处参考美团Leaf-Segment 唯一ID。

通过数据库实现批量获取id范围自增，每次程序内部自增，我们只需要控制好从数据库获取批量id的时机与范围即可。

数据库表

```mysql
CREATE TABLE `unique_id` (
    biz_tag VARCHAR(120) NOT NULL,
    max_id BIGINT UNSIGNED NOT NULL,
    step INT UNSIGNED NOT NULL,
    desc VARCHAR(256),
    update_time timestamp,
	PRIMARY KEY(`biz_tag`)
);
```

业务逻辑

A 线程处理请求

- 当A线程检测阈值发现即将超标，则去数据库申请ID

B 线程根据信号去数据库拉取数据



### 实现思路 & 评价

当前实现过于混乱，chan 与 ringBuffer方式并没有完全实现完成，思路过于混乱。

后续重新调整实现

分为几个部分

1. 管理层（门面，外部简单调用）
   - 将 数据结构层 与 数据填充层结合，使其可以
2. 数据结构层（ ringBuffer or Chan）
   - 本层只需要提供三个方法，（ fill, get ）
     - fill 填充数据
     - get 获取id
   - 让底层可以使用更多丰富的实现。
3. 数据填充层
   - 要求单例
   - 从各种数据库获取后，调用数据结构层的 fill 方法，填充相应的id进入。
   - 当前方法只做读取数据，并将数据写入，具体何时读取数据，管理层处理



#### 数据填充层

##### ringBuffer 

优点

- 相比于chan，去锁化，性能较高

缺点

- 需要考虑数据填充时可能带来的等待，以一种优雅的方式处理等待

- 可能出现饿死请求问题

  

##### chan

优点

- 便于实现，几乎无Bug
  - 数据填充层保留chan
  - 数据结构层直接返回chan
- 由于有优先队列，不会出现饿死goroutine问题

缺点

- 每次获取时都会触发锁，性能较低