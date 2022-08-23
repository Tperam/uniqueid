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
    `desc` VARCHAR(255) NOT NULL DEFAULT '',
    start_time timestamp DEFAULT CURRENT_TIMESTAMP,
    update_time timestamp DEFAULT CURRENT_TIMESTAMP,
	PRIMARY KEY(`biz_tag`)
);
```

业务逻辑

A 线程处理请求

- 当A线程检测阈值发现即将超标，则去数据库申请ID

B 线程根据信号去数据库拉取数据

- 