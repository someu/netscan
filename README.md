## API接口

### 任务

#### 下发任务

#### 获取任务详情

#### 暂停任务

#### 恢复任务

#### 取消任务

### 全局配置

#### 获取全局配置

```shell
GET /api/config
```

#### 设置全局配置

```shell
POST /api/config
```

## 参数设置

### 模式

#### 任务模式

```shell
--mode task
```

#### api模式

```shell
--mode api
```

### 输入

#### 文件读取

```shell
--file
```

#### ip

```shell
-i/--ip
```

#### port

```shell
-p/--port
```

### 输出

#### List

```shell
--output-list <filename>
```

#### JSON

```shell
--output-json <filename>
```

### 速率限制

#### packet per send

```shell
--pps <number>
```

#### request concurrent

```shell
--request-concurrent
```

#### 超时

```shell
--timeout
```

### 持久化

#### sqllite

```shell
--sqllite <sqlliteurl>
```

#### Redis

```shell
--redis <redisurl>
```

#### MongoDB

```shell
--mongodb <mongourl>
```

#### 主动推送

```shell
--pushback <url>
```

### 监听配置

```shell
--listen
```

### 读取配置

```shell
-c/--config
```

### 分布式

