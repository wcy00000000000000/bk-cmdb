# 主机转移类工具参考

## 1. open_transfer_host_module

业务内主机转移到普通模块，不可以转移到空闲机、故障机、待回收模块。
`is_increment` 参数表示是否将主机追加到目标模块，如果设置为 true 则仅会新增主机和目标模块的关系，否则会删除原有关系。如果源模块不是普通模块则该参数不能设置为true。

### 调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_host_id": [12345, 67890, 11111],
    "bk_module_id": [201, 202],
    "is_increment": false
  }
}
```

---

## 2. open_transfer_host_to_idlemodule

将业务内的主机从当前模块转移到该业务的空闲机模块。

### 调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_host_id": [12345, 67890, 11111]
  }
}
```

---

## 3. open_transfer_host_to_faultmodule

将业务内的主机从当前模块转移到该业务的故障机模块。

### 调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_host_id": [12345, 67890]
  }
}
```

---

## 4. open_transfer_host_to_recyclemodule

将业务内的主机从当前模块转移到该业务的待回收模块。

### 调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_host_id": [12345, 67890]
  }
}
```

---

## 5. open_transfer_host_to_resourcemodule

将主机从业务中转移到资源池，主机必须已在该业务的空闲机模块下，如果不是空闲机模块的主机需要先使用 `open_transfer_host_to_idlemodule` 转到空闲机。
`bk_module_id` 参数表示转移到的主机池目录ID，不填默认转移到主机池的空闲机目录，如果用户没有指定的话默认不填。

### 转移到默认空闲机目录的调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_host_id": [12345, 67890, 11111]
  }
}
```

### 转移到指定资源池目录的调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_module_id": 5,
    "bk_host_id": [12345, 67890, 11111]
  }
}
```

---

## 6. open_transfer_resourcehost_to_idlemodule

将资源池中的主机分配到目标业务的空闲机模块，如果最终目标是业务内的某个具体模块，需在此步之后再调用 `open_transfer_host_module` 进行二次转移。

### 调用示例

```json
{
  "body_param": {
    "bk_biz_id": 3,
    "bk_host_id": [12345, 67890, 11111]
  }
}
```

---

## 7. open_transfer_host_across_biz

- 跨业务转移主机，只能将源业务空闲机池集群中的主机转移到目标业务的空闲机池集群。`bk_module_id` 参数表示主机要转移到的模块ID，该模块ID必须为下空闲机池集群下的模块ID。
- 如果源模块不是空闲机需要先使用 `open_transfer_host_to_idlemodule` 转到空闲机。
- 如果目标模块是业务内的某个具体模块，需要先使用 `open_get_biz_internal_module` 获取目标业务的空闲机模块 ID 再调用该工具，并且需在此步之后再调用 `open_transfer_host_module` 进行二次转移。

### 调用示例

```json
{
  "body_param": {
    "src_bk_biz_id": 3,
    "dst_bk_biz_id": 5,
    "bk_host_id": [12345, 67890, 11111],
    "bk_module_id": 7
  }
}
```

# 8. open_list_resource_pool_hosts

查询资源池中的主机。

## 调用示例

```json
{
  "body_param": {
    "host_property_filter": {
      "condition": "OR",
      "rules": [
        { "field": "bk_host_innerip", "operator": "in", "value": ["10.0.0.1", "10.0.0.2", "10.0.0.3"] },
        { "field": "bk_asset_id", "operator": "in", "value": ["SVR-001", "SVR-002"] }
      ]
    },
    "page": { "start": 0, "limit": 200 },
    "fields": ["bk_host_id", "bk_host_innerip", "bk_cloud_id", "bk_asset_id"]
  }
}
```
