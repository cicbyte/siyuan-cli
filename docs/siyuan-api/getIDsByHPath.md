# 根据人类可读路径获取文档ID (getIDsByHPath)

POST /api/filetree/getIDsByHPath

## 接口描述

本接口用于根据人类可读路径（HPath）获取对应的文档ID列表。HPath 是文档在笔记本中的层级结构路径，以文档标题组成，例如 `/笔记/子文件夹/文档标题`。

由于可能存在同名文档，所以接口返回的是一个ID列表。如果需要精确匹配单个文档，可能需要结合其他信息进一步筛选。

## 请求参数

请求体为一个 JSON 对象，包含以下字段：

| 参数名 | 类型 | 必需 | 描述 |
| --- | --- | --- | --- |
| `notebook` | string | 是 | 笔记本 ID，例如 `20200812223209-lj3enxa`。 |
| `path` | string | 是 | 人类可读路径（HPath），例如 `/笔记/子文件夹/文档标题`。 |

## 响应体

成功时，响应体为一个 JSON 对象，包含以下字段：

| 参数名 | 类型 | 描述 |
| --- | --- | --- |
| `code` | number | 状态码，0 表示成功。 |
| `msg` | string | 错误信息，成功时为空字符串。 |
| `data` | array | 文档 ID 的字符串数组，例如 `["20200813125239-hbwpz87"]`。 |

## 请求示例

```json
{
  "notebook": "20200812223209-lj3enxa",
  "path": "/笔记/子文件夹/文档标题"
}
```

## 响应示例

成功响应（找到一个或多个匹配文档）：

```json
{
  "code": 0,
  "msg": "",
  "data": ["20200813125239-hbwpz87", "20200814103045-abcdef"]
}
```

成功响应（未找到匹配文档）：

```json
{
  "code": 0,
  "msg": "",
  "data": []
}
```

失败响应（笔记本不存在）：

```json
{
  "code": -1,
  "msg": "笔记本不存在",
  "data": null
}
```

## 错误码

| 错误码 | 描述 |
| --- | --- |
| `0` | 成功 |
| `-1` | 失败，可能原因包括：笔记本 ID 格式无效、笔记本不存在等 |

## 源码定位

后端实现相关的核心逻辑可以在以下文件中找到：

- 路由定义：`siyuan/kernel/api/router.go`（路由 `/api/filetree/getIDsByHPath`）
- API 处理函数：`siyuan/kernel/api/filetree.go`（函数 `getIDsByHPath`）
- 模型层实现：`siyuan/kernel/model/file.go`（函数 `GetIDsByHPath`）