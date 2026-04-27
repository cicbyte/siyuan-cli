# 内核 API

## 说明

思源笔记是一款前后端分离架构的软件，GoLang 语言作为后端在运行时提供了大量 API 供前端调用。

### API 调用格式

在插件系统中调用后端 API, 可以使用插件 `API` 提供的 `fetchPost` 和 `fetchSyncPost` 两个方法, 这两个 API 都会在内部完成自动鉴权已经发送请求。在使用中，注意点如下:

- 使用 POST 方法
- 第一个参数为 API route
- 第二个参数为 post data, 是一个 js object
- 如果是 `fetchPost` 方法, 还会有第三个回调函数参数
- 返回值格式为
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {}
  }
  ```
  
  - `code`：非 0 为异常情况
  - `msg`：正常情况下是空字符串，异常情况下会返回错误文案
  - `data`：可能为 `{}`、`[]` 或者 `NULL`，根据不同接口而不同

### 样例代码

typescript

```
import { fetchPost } from "siyuan";

fetchPost("/api/system/currentTime", {}, (response) => {
    dialog.element.querySelector("#time").innerHTML = new Date(response.data).toString();
});
```

## API 大纲

- [说明](#%E8%AF%B4%E6%98%8E)
  
  - [API 调用格式](#api-%E8%B0%83%E7%94%A8%E6%A0%BC%E5%BC%8F)
  - [样例代码](#%E6%A0%B7%E4%BE%8B%E4%BB%A3%E7%A0%81)
- [API 大纲](#api-%E5%A4%A7%E7%BA%B2)
- [笔记本](#%E7%AC%94%E8%AE%B0%E6%9C%AC)
  
  - [列出笔记本](#%E5%88%97%E5%87%BA%E7%AC%94%E8%AE%B0%E6%9C%AC)
  - [打开笔记本](#%E6%89%93%E5%BC%80%E7%AC%94%E8%AE%B0%E6%9C%AC)
  - [关闭笔记本](#%E5%85%B3%E9%97%AD%E7%AC%94%E8%AE%B0%E6%9C%AC)
  - [重命名笔记本](#%E9%87%8D%E5%91%BD%E5%90%8D%E7%AC%94%E8%AE%B0%E6%9C%AC)
  - [创建笔记本](#%E5%88%9B%E5%BB%BA%E7%AC%94%E8%AE%B0%E6%9C%AC)
  - [删除笔记本](#%E5%88%A0%E9%99%A4%E7%AC%94%E8%AE%B0%E6%9C%AC)
  - [获取笔记本配置](#%E8%8E%B7%E5%8F%96%E7%AC%94%E8%AE%B0%E6%9C%AC%E9%85%8D%E7%BD%AE)
  - [保存笔记本配置](#%E4%BF%9D%E5%AD%98%E7%AC%94%E8%AE%B0%E6%9C%AC%E9%85%8D%E7%BD%AE)
- [文档](#%E6%96%87%E6%A1%A3)
  
  - [通过 Markdown 创建文档](#%E9%80%9A%E8%BF%87-markdown-%E5%88%9B%E5%BB%BA%E6%96%87%E6%A1%A3)
  - [重命名文档](#%E9%87%8D%E5%91%BD%E5%90%8D%E6%96%87%E6%A1%A3)
  - [删除文档](#%E5%88%A0%E9%99%A4%E6%96%87%E6%A1%A3)
  - [移动文档](#%E7%A7%BB%E5%8A%A8%E6%96%87%E6%A1%A3)
  - [根据路径获取人类可读路径](#%E6%A0%B9%E6%8D%AE%E8%B7%AF%E5%BE%84%E8%8E%B7%E5%8F%96%E4%BA%BA%E7%B1%BB%E5%8F%AF%E8%AF%BB%E8%B7%AF%E5%BE%84)
  - [根据 ID 获取人类可读路径](#%E6%A0%B9%E6%8D%AE-id-%E8%8E%B7%E5%8F%96%E4%BA%BA%E7%B1%BB%E5%8F%AF%E8%AF%BB%E8%B7%AF%E5%BE%84)
- [资源文件](#%E8%B5%84%E6%BA%90%E6%96%87%E4%BB%B6)
  
  - [上传资源文件](#%E4%B8%8A%E4%BC%A0%E8%B5%84%E6%BA%90%E6%96%87%E4%BB%B6)
- [块](#%E5%9D%97)
  
  - [插入块](#%E6%8F%92%E5%85%A5%E5%9D%97)
  - [插入前置子块](#%E6%8F%92%E5%85%A5%E5%89%8D%E7%BD%AE%E5%AD%90%E5%9D%97)
  - [插入后置子块](#%E6%8F%92%E5%85%A5%E5%90%8E%E7%BD%AE%E5%AD%90%E5%9D%97)
  - [更新块](#%E6%9B%B4%E6%96%B0%E5%9D%97)
  - [删除块](#%E5%88%A0%E9%99%A4%E5%9D%97)
  - [移动块](#%E7%A7%BB%E5%8A%A8%E5%9D%97)
  - [获取块 kramdown 源码](#%E8%8E%B7%E5%8F%96%E5%9D%97-kramdown-%E6%BA%90%E7%A0%81)
  - [获取子块](#%E8%8E%B7%E5%8F%96%E5%AD%90%E5%9D%97)
  - [转移块引用](#%E8%BD%AC%E7%A7%BB%E5%9D%97%E5%BC%95%E7%94%A8)
- [属性](#%E5%B1%9E%E6%80%A7)
  
  - [设置块属性](#%E8%AE%BE%E7%BD%AE%E5%9D%97%E5%B1%9E%E6%80%A7)
  - [获取块属性](#%E8%8E%B7%E5%8F%96%E5%9D%97%E5%B1%9E%E6%80%A7)
- [SQL](#sql)
  
  - [执行 SQL 查询](#%E6%89%A7%E8%A1%8C-sql-%E6%9F%A5%E8%AF%A2)
- [模板](#%E6%A8%A1%E6%9D%BF)
  
  - [渲染模板](#%E6%B8%B2%E6%9F%93%E6%A8%A1%E6%9D%BF)
  - [渲染 Sprig](#%E6%B8%B2%E6%9F%93-sprig)
- [文件](#%E6%96%87%E4%BB%B6)
  
  - [获取文件](#%E8%8E%B7%E5%8F%96%E6%96%87%E4%BB%B6)
  - [写入文件](#%E5%86%99%E5%85%A5%E6%96%87%E4%BB%B6)
  - [删除文件](#%E5%88%A0%E9%99%A4%E6%96%87%E4%BB%B6)
  - [重命名文件](#%E9%87%8D%E5%91%BD%E5%90%8D%E6%96%87%E4%BB%B6)
  - [列出文件](#%E5%88%97%E5%87%BA%E6%96%87%E4%BB%B6)
- [导出](#%E5%AF%BC%E5%87%BA)
  
  - [导出 Markdown 文本](#%E5%AF%BC%E5%87%BA-markdown-%E6%96%87%E6%9C%AC)
- [转换](#%E8%BD%AC%E6%8D%A2)
  
  - [Pandoc](#pandoc)
- [通知](#%E9%80%9A%E7%9F%A5)
  
  - [推送消息](#%E6%8E%A8%E9%80%81%E6%B6%88%E6%81%AF)
  - [推送报错消息](#%E6%8E%A8%E9%80%81%E6%8A%A5%E9%94%99%E6%B6%88%E6%81%AF)
- [系统](#%E7%B3%BB%E7%BB%9F)
  
  - [获取启动进度](#%E8%8E%B7%E5%8F%96%E5%90%AF%E5%8A%A8%E8%BF%9B%E5%BA%A6)
  - [获取系统版本](#%E8%8E%B7%E5%8F%96%E7%B3%BB%E7%BB%9F%E7%89%88%E6%9C%AC)
  - [获取系统当前时间](#%E8%8E%B7%E5%8F%96%E7%B3%BB%E7%BB%9F%E5%BD%93%E5%89%8D%E6%97%B6%E9%97%B4)

***

## 笔记本

### 列出笔记本

- `/api/notebook/lsNotebooks`
- 不带参
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "notebooks": [
              {
                  "id": "20210817205410-2kvfpfn",
                  "name": "测试笔记本",
                  "icon": "1f41b",
                  "sort": 0,
                  "closed": false
              },
              {
                  "id": "20210808180117-czj9bvb",
                  "name": "思源笔记用户指南",
                  "icon": "1f4d4",
                  "sort": 1,
                  "closed": false
              }
          ]
      }
  }
  ```
  
  

### 打开笔记本

- `/api/notebook/openNotebook`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  

### 关闭笔记本

- `/api/notebook/closeNotebook`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  

### 重命名笔记本

- `/api/notebook/renameNotebook`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0",
      "name": "笔记本的新名称"
  }
  ```
  
  - `notebook`：笔记本 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  

### 创建笔记本

- `/api/notebook/createNotebook`
- 参数
  
  json
  
  ```
  {
      "name": "笔记本的名称"
  }
  ```
  
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "notebook": {
              "id": "20220126215949-r1wvoch",
              "name": "笔记本的名称",
              "icon": "",
              "sort": 0,
              "closed": false
          }
      }
  }
  ```
  
  

### 删除笔记本

- `/api/notebook/removeNotebook`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

### 获取笔记本配置

- `/api/notebook/getNotebookConf`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210817205410-2kvfpfn"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "box": "20210817205410-2kvfpfn",
          "conf": {
              "name": "测试笔记本",
              "closed": false,
              "refCreateSavePath": "",
              "createDocNameTemplate": "",
              "dailyNoteSavePath": "/daily note/{{now | date \"2006/01\"}}/{{now | date \"2006-01-02\"}}",
              "dailyNoteTemplatePath": ""
          },
          "name": "测试笔记本"
      }
  }
  ```
  
  

### 保存笔记本配置

- `/api/notebook/setNotebookConf`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210817205410-2kvfpfn",
      "conf": {
          "name": "测试笔记本",
          "closed": false,
          "refCreateSavePath": "",
          "createDocNameTemplate": "",
          "dailyNoteSavePath": "/daily note/{{now | date \"2006/01\"}}/{{now | date \"2006-01-02\"}}",
          "dailyNoteTemplatePath": ""
      }
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "name": "测试笔记本",
          "closed": false,
          "refCreateSavePath": "",
          "createDocNameTemplate": "",
          "dailyNoteSavePath": "/daily note/{{now | date \"2006/01\"}}/{{now | date \"2006-01-02\"}}",
          "dailyNoteTemplatePath": ""
      }
  }
  ```
  
  

## 文档

### 通过 Markdown 创建文档

- `/api/filetree/createDocWithMd`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210817205410-2kvfpfn",
      "path": "/foo/bar",
      "markdown": ""
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
  - `path`：文档路径，需要以 / 开头，中间使用 / 分隔层级（这里的 path 对应数据库 hpath 字段）
  - `markdown`：GFM Markdown 内容
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": "20210914223645-oj2vnx2"
  }
  ```
  
  
  
  - `data`：创建好的文档 ID
  - 如果使用同一个 `path` 重复调用该接口，不会覆盖已有文档

### 重命名文档

- `/api/filetree/renameDoc`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0",
      "path": "/20210902210113-0avi12f.sy",
      "title": "文档新标题"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
  - `path`：文档路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

### 删除文档

- `/api/filetree/removeDoc`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0",
      "path": "/20210902210113-0avi12f.sy"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
  - `path`：文档路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

### 移动文档

- `/api/filetree/moveDocs`
- 参数
  
  json
  
  ```
  {
      "fromPaths": ["/20210917220056-yxtyl7i.sy"],
      "toNotebook": "20210817205410-2kvfpfn",
      "toPath": "/"
  }
  ```
  
  
  
  - `fromPaths`：源路径
  - `toNotebook`：目标笔记本 ID
  - `toPath`：目标路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

### 根据路径获取人类可读路径

- `/api/filetree/getHPathByPath`
- 参数
  
  json
  
  ```
  {
      "notebook": "20210831090520-7dvbdv0",
      "path": "/20210917220500-sz588nq/20210917220056-yxtyl7i.sy"
  }
  ```
  
  
  
  - `notebook`：笔记本 ID
  - `path`：路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": "/foo/bar"
  }
  ```
  
  

### 根据 ID 获取人类可读路径

- `/api/filetree/getHPathByID`
- 参数
  
  json
  
  ```
  {
      "id": "20210917220056-yxtyl7i"
  }
  ```
  
  - `id`：块 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": "/foo/bar"
  }
  ```
  
  

## 资源文件

### 上传资源文件

- `/api/asset/upload`
- 参数为 HTTP Multipart 表单
  
  - `assetsDirPath`：资源文件存放的文件夹路径，以 data 文件夹作为根路径，比如：
    
    - `"/assets/"`：工作空间/data/assets/ 文件夹
    - `"/assets/sub/"`：工作空间/data/assets/sub/ 文件夹
    
    常规情况下建议用第一种，统一存放到工作空间资源文件夹下，放在子目录有一些副作用，请参考用户指南资源文件章节。
  - `file[]`：上传的文件列表
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "errFiles": [""],
          "succMap": {
              "foo.png": "assets/foo-20210719092549-9j5y79r.png"
          }
      }
  }
  ```
  
  - `errFiles`：处理时遇到错误的文件名
  - `succMap`：处理成功的文件，key 为上传时的文件名，value 为 assets/foo-id.png，用于将已有 Markdown 内容中的资源文件链接地址替换为上传后的地址

## 块

### 插入块

- `/api/block/insertBlock`
- 参数
  
  json
  
  ```
  {
      "dataType": "markdown",
      "data": "foo**bar**{: style=\"color: var(--b3-font-color8);\"}baz",
      "nextID": "",
      "previousID": "20211229114650-vrek5x6",
      "parentID": ""
  }
  
  ```
  
  - `dataType`：待插入数据类型，值可选择 `markdown` 或者 `dom`
  - `data`：待插入的数据
  - `nextID`：后一个块的 ID，用于锚定插入位置
  - `previousID`：前一个块的 ID，用于锚定插入位置
  - `parentID`：父块 ID，用于锚定插入位置
  
  `nextID`、`previousID`、`parentID` 三个参数必须至少存在一个有值，优先级为 `nextID` &gt; `previousID` &gt; `parentID`
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "doOperations": [
                  {
                      "action": "insert",
                      "data": "<div data-node-id=\"20211230115020-g02dfx0\" data-node-index=\"1\" data-type=\"NodeParagraph\" class=\"p\"><div contenteditable=\"true\" spellcheck=\"false\">foo<strong style=\"color: var(--b3-font-color8);\">bar</strong>baz</div><div class=\"protyle-attr\" contenteditable=\"false\"></div></div>",
                      "id": "20211230115020-g02dfx0",
                      "parentID": "",
                      "previousID": "20211229114650-vrek5x6",
                      "retData": null
                  }
              ],
              "undoOperations": null
          }
      ]
  }
  ```
  
  
  
  - `action.data`：新插入块生成的 DOM
  - `action.id`：新插入块的 ID

### 插入前置子块

- `/api/block/prependBlock`
- 参数
  
  json
  
  ```
  {
      "data": "foo**bar**{: style=\"color: var(--b3-font-color8);\"}baz",
      "dataType": "markdown",
      "parentID": "20220107173950-7f9m1nb"
  }
  ```
  
  
  
  - `dataType`：待插入数据类型，值可选择 `markdown` 或者 `dom`
  - `data`：待插入的数据
  - `parentID`：父块的 ID，用于锚定插入位置
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "doOperations": [
                  {
                      "action": "insert",
                      "data": "<div data-node-id=\"20220108003710-hm0x9sc\" data-node-index=\"1\" data-type=\"NodeParagraph\" class=\"p\"><div contenteditable=\"true\" spellcheck=\"false\">foo<strong style=\"color: var(--b3-font-color8);\">bar</strong>baz</div><div class=\"protyle-attr\" contenteditable=\"false\"></div></div>",
                      "id": "20220108003710-hm0x9sc",
                      "parentID": "20220107173950-7f9m1nb",
                      "previousID": "",
                      "retData": null
                  }
              ],
              "undoOperations": null
          }
      ]
  }
  ```
  
  
  
  - `action.data`：新插入块生成的 DOM
  - `action.id`：新插入块的 ID

### 插入后置子块

- `/api/block/appendBlock`
- 参数
  
  json
  
  ```
  {
      "data": "foo**bar**{: style=\"color: var(--b3-font-color8);\"}baz",
      "dataType": "markdown",
      "parentID": "20220107173950-7f9m1nb"
  }
  ```
  
  
  
  - `dataType`：待插入数据类型，值可选择 `markdown` 或者 `dom`
  - `data`：待插入的数据
  - `parentID`：父块的 ID，用于锚定插入位置
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "doOperations": [
                  {
                      "action": "insert",
                      "data": "<div data-node-id=\"20220108003642-y2wmpcv\" data-node-index=\"1\" data-type=\"NodeParagraph\" class=\"p\"><div contenteditable=\"true\" spellcheck=\"false\">foo<strong style=\"color: var(--b3-font-color8);\">bar</strong>baz</div><div class=\"protyle-attr\" contenteditable=\"false\"></div></div>",
                      "id": "20220108003642-y2wmpcv",
                      "parentID": "20220107173950-7f9m1nb",
                      "previousID": "20220108003615-7rk41t1",
                      "retData": null
                  }
              ],
              "undoOperations": null
          }
      ]
  }
  ```
  
  
  
  - `action.data`：新插入块生成的 DOM
  - `action.id`：新插入块的 ID

### 更新块

- `/api/block/updateBlock`
- 参数
  
  json
  
  ```
  {
      "dataType": "markdown",
      "data": "foo bar baz",
      "id": "20211230161520-querkps"
  }
  ```
  
  
  
  - `dataType`：待更新数据类型，值可选择 `markdown` 或者 `dom`
  - `data`：待更新的数据
  - `id`：待更新块的 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "doOperations": [
                  {
                      "action": "update",
                      "data": "<div data-node-id=\"20211230161520-querkps\" data-node-index=\"1\" data-type=\"NodeParagraph\" class=\"p\"><div contenteditable=\"true\" spellcheck=\"false\">foo<strong>bar</strong>baz</div><div class=\"protyle-attr\" contenteditable=\"false\"></div></div>",
                      "id": "20211230161520-querkps",
                      "parentID": "",
                      "previousID": "",
                      "retData": null
                  }
              ],
              "undoOperations": null
          }
      ]
  }
  ```
  
  
  
  - `action.data`：更新块生成的 DOM

### 删除块

- `/api/block/deleteBlock`
- 参数
  
  json
  
  ```
  {
      "id": "20211230161520-querkps"
  }
  ```
  
  1  
  2  
  3
  
  - `id`：待删除块的 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "doOperations": [
                  {
                      "action": "delete",
                      "data": null,
                      "id": "20211230162439-vtm09qo",
                      "parentID": "",
                      "previousID": "",
                      "retData": null
                  }
              ],
              "undoOperations": null
          }
      ]
  }
  ```
  
  

### 移动块

- `/api/block/moveBlock`
- 参数
  
  json
  
  ```
  {
      "id": "20230406180530-3o1rqkc",
      "previousID": "20230406152734-if5kyx6",
      "parentID": "20230404183855-woe52ko"
  }
  ```
  
  
  
  - `id`：待移动块 ID
  - `previousID`：前一个块的 ID，用于锚定插入位置
  - `parentID`：父块的 ID，用于锚定插入位置，`previousID` 和 `parentID` 不能同时为空，同时存在的话优先使用 `previousID`
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "doOperations": [
                  {
                      "action": "move",
                      "data": null,
                      "id": "20230406180530-3o1rqkc",
                      "parentID": "20230404183855-woe52ko",
                      "previousID": "20230406152734-if5kyx6",
                      "nextID": "",
                      "retData": null,
                      "srcIDs": null,
                      "name": "",
                      "type": ""
                  }
              ],
              "undoOperations": null
          }
      ]
  }
  ```
  
  

### 获取块 kramdown 源码

- `/api/block/getBlockKramdown`
- 参数
  
  json
  
  ```
  {
      "id": "20201225220955-l154bn4"
  }
  ```
  
  1  
  2  
  3
  
  - `id`：待获取块的 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "id": "20201225220955-l154bn4",
          "kramdown": "* {: id=\"20201225220955-2nn1mns\"}新建笔记本，在笔记本下新建文档\n  {: id=\"20210131155408-3t627wc\"}\n* {: id=\"20201225220955-uwhqnug\"}在编辑器中输入 <kbd>/</kbd> 触发功能菜单\n  {: id=\"20210131155408-btnfw88\"}\n* {: id=\"20201225220955-04ymi2j\"}((20200813131152-0wk5akh \"在内容块中遨游\"))、((20200822191536-rm6hwid \"窗口和页签\"))\n  {: id=\"20210131155408-hh1z442\"}"
      }
  }
  ```
  
  

### 获取子块

- `/api/block/getChildBlocks`
- 参数
  
  json
  
  ```
  {
      "id": "20230506212712-vt9ajwj"
  }
  ```
  
  
  
  - `id`：父块 ID
  - 标题下方块也算作子块
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "id": "20230512083858-mjdwkbn",
              "type": "h",
              "subType": "h1"
          },
          {
              "id": "20230513213727-thswvfd",
              "type": "s"
          },
          {
              "id": "20230513213633-9lsj4ew",
              "type": "l",
              "subType": "u"
          }
      ]
  }
  ```
  
  

### 转移块引用

- `/api/block/transferBlockRef`
- 参数
  
  json
  
  ```
  {
      "fromID": "20230612160235-mv6rrh1",
      "toID": "20230613093045-uwcomng",
      "refIDs": ["20230613092230-cpyimmd"]
  }
  ```
  
  
  
  - `fromID`：定义块 ID
  - `toID`：目标块 ID
  - `refIDs`：指向定义块 ID 的引用所在块 ID，可选，如果不指定，所有指向定义块 ID 的引用块 ID 都会被转移
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

## 属性

### 设置块属性

- `/api/attr/setBlockAttrs`
- 参数
  
  json
  
  ```
  {
      "id": "20210912214605-uhi5gco",
      "attrs": {
          "custom-attr1": "line1\nline2"
      }
  }
  ```
  
  
  
  - `id`：块 ID
  - `attrs`：块属性，自定义属性必须以 `custom-` 作为前缀
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

### 获取块属性

- `/api/attr/getBlockAttrs`
- 参数
  
  json
  
  ```
  {
      "id": "20210912214605-uhi5gco"
  }
  ```
  
  
  
  - `id`：块 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "custom-attr1": "line1\nline2",
          "id": "20210912214605-uhi5gco",
          "title": "PDF 标注双链演示",
          "type": "doc",
          "updated": "20210916120715"
      }
  }
  ```
  
  

## SQL

### 执行 SQL 查询

- `/api/query/sql`
- 参数
  
  json
  
  ```
  {
      "stmt": "SELECT * FROM blocks WHERE content LIKE'%content%' LIMIT 7"
  }
  ```
  
  
  
  - `stmt`：SQL 脚本
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [{ "列": "值" }]
  }
  ```
  
  

## 模板

### 渲染模板

- `/api/template/render`
- 参数
  
  json
  
  ```
  {
      "id": "20220724223548-j6g0o87",
      "path": "F:\\SiYuan\\data\\templates\\foo.md"
  }
  ```
  
  
  
  - `id`：调用渲染所在的文档 ID
  - `path`：模板文件绝对路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "content": "<div data-node-id=\"20220729234848-dlgsah7\" data-node-index=\"1\" data-type=\"NodeParagraph\" class=\"p\" updated=\"20220729234840\"><div contenteditable=\"true\" spellcheck=\"false\">foo</div><div class=\"protyle-attr\" contenteditable=\"false\"></div></div>",
          "path": "F:\\SiYuan\\data\\templates\\foo.md"
      }
  }
  ```
  
  

### 渲染 Sprig

- `/api/template/renderSprig`
- 参数
  
  json
  
  ```
  {
      "template": "/daily note/{{now | date \"2006/01\"}}/{{now | date \"2006-01-02\"}}"
  }
  ```
  
  
  
  - `template`：模板内容
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": "/daily note/2023/03/2023-03-24"
  }
  ```
  
  

## 文件

### 获取文件

- `/api/file/getFile`
- 参数
  
  json
  
  ```
  {
      "path": "/data/20210808180117-6v0mkxr/20200923234011-ieuun1p.sy"
  }
  ```
  
  - `path`：工作空间路径下的文件路径
- 返回值
  
  文件内容

### 写入文件

- `/api/file/putFile`
- 参数为 HTTP Multipart 表单
  
  - `path`：工作空间路径下的文件路径
  - `isDir`：是否为创建文件夹，为 `true` 时仅创建文件夹，忽略 `file`
  - `modTime`：最近访问和修改时间，Unix time
  - `file`：上传的文件
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  

### 删除文件

- `/api/file/removeFile`
- 参数
  
  json
  
  ```
  {
      "path": "/data/20210808180117-6v0mkxr/20200923234011-ieuun1p.sy"
  }
  ```
  
  - `path`：工作空间路径下的文件路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  

### 重命名文件

- `/api/file/renameFile`
- 参数
  
  json
  
  ```
  {
      "path": "/data/assets/image-20230523085812-k3o9t32.png",
      "newPath": "/data/assets/test-20230523085812-k3o9t32.png"
  }
  ```
  
  - `path`：工作空间路径下的文件路径
  - `newPath`：新的文件路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  

### 列出文件

- `/api/file/readDir`
- 参数
  
  json
  
  ```
  {
      "path": "/data/20210808180117-6v0mkxr/20200923234011-ieuun1p.sy"
  }
  ```
  
  - `path`：工作空间路径下的文件路径
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": [
          {
              "isDir": true,
              "name": "20210808180320-abz7w6k"
          },
          {
              "isDir": false,
              "name": "20210808180320-abz7w6k.sy"
          }
      ]
  }
  ```
  

## 导出

### 导出 Markdown 文本

- `/api/export/exportMdContent`
- 参数
  
  json
  
  ```
  {
      "id": ""
  }
  ```
  
  
  
  - `id`：要导出的文档块 ID
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "hPath": "/0 请从这里开始",
          "content": "## 🍫 内容块\n\n在思源中，唯一重要的核心概念是..."
      }
  }
  ```
  
  - `hPath`：人类可读的路径
  - `content`：Markdown 内容

## 转换

### Pandoc

- `/api/convert/pandoc`
- 工作目录
  
  - 执行调用 pandoc 命令时工作目录会被设置在 `工作空间/temp/convert/pandoc/` 下
  - 可先通过 API [`写入文件`](#%E5%86%99%E5%85%A5%E6%96%87%E4%BB%B6) 将待转换文件写入该目录
  - 然后再调用该 API 进行转换，转换后的文件也会被写入该目录
  - 最后调用 API [`获取文件`](#%E8%8E%B7%E5%8F%96%E6%96%87%E4%BB%B6) 获取转换后的文件内容
    
    - 或者调用 API [`通过 Markdown 创建文档`](#%E9%80%9A%E8%BF%87-markdown-%E5%88%9B%E5%BB%BA%E6%96%87%E6%A1%A3)
    - 或者调用内部 API `importStdMd` 将转换后的文件夹直接导入
- 参数
  
  json
  
  ```
  {
      "args": ["--to", "markdown_strict-raw_html", "foo.epub", "-o", "foo.md"]
  }
  ```
  
  
  
  - `args`：Pandoc 命令行参数
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": null
  }
  ```
  
  

## 通知

### 推送消息

- `/api/notification/pushMsg`
- 参数
  
  json
  
  ```
  {
      "msg": "test",
      "timeout": 7000
  }
  ```
  
  
  
  - `timeout`：消息持续显示时间，单位为毫秒。可以不传入该字段，默认为 7000 毫秒
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "id": "62j1m2i"
      }
  }
  ```
  
  
  
  - `id`：消息 ID

### 推送报错消息

- `/api/notification/pushErrMsg`
- 参数
  
  json
  
  ```
  {
      "msg": "test",
      "timeout": 7000
  }
  ```
  
  - `timeout`：消息持续显示时间，单位为毫秒。可以不传入该字段，默认为 7000 毫秒
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "id": "1q2c9z3"
      }
  }
  ```
  
  
  
  - `id`：消息 ID

## 系统

### 获取启动进度

- `/api/system/bootProgress`
- 不带参
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": {
          "details": "Finishing boot...",
          "progress": 100
      }
  }
  ```
  
  

### 获取系统版本

- `/api/system/version`
- 不带参
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": "1.3.5"
  }
  ```
  
  

### 获取系统当前时间

- `/api/system/currentTime`
- 不带参
- 返回值
  
  json
  
  ```
  {
      "code": 0,
      "msg": "",
      "data": 1631850968131
  }
  ```
  
  
  
  - `data`: 精度为毫秒