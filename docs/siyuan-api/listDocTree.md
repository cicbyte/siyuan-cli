[← 返回 FileTree API 列表](../pages/filetree.html)

# 列出文档树

[在 GitHub 上查看源码](https://github.com/siyuan-note/siyuan/blob/master/kernel/api/filetree.go)

`POST /api/filetree/listDocTree`

## 描述

获取指定笔记本或文档的文档树结构，返回文档的层级关系和基本信息。可用于构建文档树视图或获取文档层次结构。

## 请求参数

请求体需要是 JSON 格式。

| 参数名 | 类型 | 描述 | 必需 |
| :--- | :--- | :--- | :--- |
| `notebook` | `string` | 笔记本 ID | 是 |
| `path` | `string` | 文档路径，从根路径开始（如 "/a/b"）。如果为空，则返回整个笔记本的文档树 | 否 |
| `sort` | `number` | 排序方式（0：按名称，1：按更新时间，2：按创建时间，3：自定义） | 否 |

## 响应结果

响应体是 JSON 格式。`code` 为 0 表示成功。`data` 包含文档树信息。

```json
{
  "code": 0,
  "msg": "",
  "data": {
    "path": "/path/to/doc",  // 文档路径
    "files": [
      {
        "path": "/path/to/doc/doc1.sy",  // 文档文件路径
        "name": "doc1",                  // 文档名称
        "icon": "📄",                     // 文档图标
        "id": "20200812220555-lj3enxa",  // 文档块 ID
        "count": 5,                      // 文档中的块数量
        "hSize": "1.2 KB",               // 文档人类可读大小
        "size": 1235,                    // 文档大小（字节）
        "updated": "20220428104712",     // 更新时间
        "created": "20220428104712"      // 创建时间
      },
      {
        "path": "/path/to/doc/folder1",  // 文件夹路径
        "name": "folder1",               // 文件夹名称
        "icon": "",                      // 文件夹图标
        "count": 0,                      // 文件夹中的文档数量
        "hSize": "",                     // 文件夹人类可读大小
        "size": 0,                       // 文件夹大小（字节）
        "subFileCount": 3                // 子文件数量
      }
    ]
  }
}
```

> 注意：这是一个社区维护的文档，可能与官方最新版本存在差异。
> 
> 如果您觉得本文档有帮助，可以考虑赞助支持：[爱发电](https://afdian.com/a/leolee9086?tab=feed)