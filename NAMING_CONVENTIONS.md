# Texere 命名体系文档

> 全部使用拉丁语词根，保持语言一致性

## 🏛️ 主项目

### **Texere** (拉丁语)
- **含义**：编织、纺织
- **发音**：/ˈteks.ɛ.re/ (拉丁语) 或 /ˈtɛksəri/ (英语化)
- **词源**：拉丁动词 *texere* (to weave)
- **Slogan**：*Weave Knowledge Together* - 编织知识，连接智慧
- **应用范围**：整个文档协作与 AI 生成平台

---

## 📦 核心子包（全部拉丁语）

### 1. **Concordia** - OT 协调核心 ✅

- **词源**：拉丁语 *concordia*
- **含义**：和谐、一致、共识
- **职责**：Operational Transformation 核心算法
- **包路径**：`github.com/coreseekdev/texere/pkg/concordia`

```go
import "github.com/coreseekdev/texere/pkg/concordia"

// OT 操作
op := concordia.NewInsert(0, "Hello")
transformed := concordia.Transform(op1, op2)
```

**为什么选择 Concordia？**
- ✅ 拉丁语，与 Texere 保持一致
- ✅ "和谐"完美描述 OT 的目标：多方编辑达成一致
- ✅ Concordia 是罗马神话中的和谐女神，有文化底蕴
- ✅ 比 Syntaxis（希腊语）更符合整体命名体系

**对比 Syntaxis**：
| 维度 | Syntaxis (希腊语) | Concordia (拉丁语) |
|------|------------------|-------------------|
| 语言一致性 | ⚠️ 不一致 | ✅ 一致 |
| 技术准确性 | ⭐⭐⭐⭐⭐ 完美 | ⭐⭐⭐⭐ 很好 |
| 文化底蕴 | ⭐⭐⭐ 学术 | ⭐⭐⭐⭐⭐ 神话 |
| 发音难度 | ⭐⭐⭐ 中等 | ⭐⭐⭐⭐ 简单 |
| 品牌感 | ⭐⭐⭐ 技术 | ⭐⭐⭐⭐ 优雅 |

---

### 2. **Unio** - 统一与排序

- **词源**：拉丁语 *unio*
- **含义**：统一、联合、合一
- **职责**：逻辑时钟、向量时钟、版本管理、全局排序
- **包路径**：`github.com/coreseekdev/texere/pkg/unio`

```go
import "github.com/coreseekdev/texere/pkg/unio"

// 逻辑时钟
clock := unio.NewLamportClock()
timestamp := unio.Tick(clock)

// 向量时钟
vclock := unio.NewVectorClock()
unio.Increment(vclock, "user-1")
```

**为什么选择 Unio？**
- ✅ 拉丁语
- ✅ "统一"描述了时间排序的目标：统一多地的操作
- ✅ 简短（4 个字母）
- ✅ 与 "union"（联合）相关

**对比 Ordo**：
- *Ordo* (秩序)：更偏向秩序、规则
- *Unio* (统一)：更强调合而为一的过程
- 推荐：**Unio** 更适合分布式系统的语境

---

### 3. **Textor** - 文本处理

- **词源**：拉丁语 *textor*
- **含义**：编织者、纺织者
- **职责**：Rope、Piece Table、光标操作、文本选择
- **包路径**：`github.com/coreseekdev/texere/pkg/textor`

```go
import "github.com/coreseekdev/texere/pkg/textor"

// Rope 数据结构
rope := textor.NewRope("Hello World")
rope.Insert(5, "Beautiful")

// Piece Table
pt := textor.NewPieceTable()
pt.Insert(0, "Hello")
```

**为什么选择 Textor？**
- ✅ 拉丁语
- ✅ "编织者"与 Texere "编织"完美呼应
- ✅ Textor 是从事纺织的人，隐喻很贴切
- ✅ 与 "text"（文本）相关

---

### 4. **Fabric** - 文档织物

- **词源**：拉丁语 *fabricum* (作坊) → 法语 *fabrique* → 英语 *fabric*
- **含义**：织物、结构、构造
- **职责**：文档模型、文档块、增量变更、补丁应用
- **包路径**：`github.com/coreseekdev/texere/pkg/fabric`

```go
import "github.com/coreseekdev/texere/pkg/fabric"

// 文档织物
doc := fabric.NewDocument()
block := fabric.NewBlock("heading", "Title")
fabric.Append(doc, block)
```

**为什么选择 Fabric？**
- ✅ 源于拉丁语
- ✅ "织物"与 Texere "编织"形成隐喻体系
- ✅ 在编程中常用（如 Fabric.js, System Fabric）
- ✅ 暗示文档是由多个部分编织而成的结构

---

### 5. **Weave** - 编织引擎

- **词源**：古英语 *weawfan* < 拉丁语 *texere*
- **含义**：编织、交织
- **职责**：核心编织引擎、AI + 人工协同、流水线
- **包路径**：`github.com/coreseekdev/texere/pkg/weave`

```go
import "github.com/coreseekdev/texere/pkg/weave/engine"

engine := weave.NewEngine()
engine.WeaveHuman(&humanEdit)
engine.WeaveAI(&aiRequest)
```

**为什么选择 Weave？**
- ✅ 词源可追溯至拉丁语 *texere*
- ✅ 直接使用"编织"这个动词，语义清晰
- ✅ 主项目叫 Texere（编织），引擎叫 Weave（编织），形成呼应
- ✅ 简单易懂，无需解释

**替代方案**：
- *Texo* (拉丁语：我编织) - 稍显古奥
- *Plecto* (拉丁语：编织、缠绕) - 过于生僻
- **Weave** - 最佳选择 ✅

---

### 6. **Flux** - 数据流动

- **词源**：拉丁语 *fluxus*
- **含义**：流动、流动的
- **职责**：WebSocket、WebRTC、传输层、同步协议
- **包路径**：`github.com/coreseekdev/texere/pkg/flux`

```go
import "github.com/coreseekdev/texere/pkg/flux"

transport := flux.NewWebSocket()
flux.Subscribe(transport, "doc-001", handler)
```

**为什么选择 Flux？**
- ✅ 拉丁语
- ✅ "流动"完美描述实时同步的数据流
- ✅ 在技术圈流行（如 Flux architecture, Redux）
- ✅ 简短有力

---

### 7. **Store** - 持久化存储

- **词源**：古法语 *estore* < 拉丁语 *instaurare* (建立、恢复)
- **含义**：存储、仓库
- **职责**：数据库、仓库模式、快照、缓存
- **包路径**：`github.com/coreseekdev/texere/pkg/store`

```go
import "github.com/coreseekdev/texere/pkg/store"

db := store.NewDatabase()
repo := store.NewRepository(db)
store.SaveSnapshot(doc)
```

**为什么选择 Store？**
- ✅ 源于拉丁语
- ✅ 技术圈通用术语
- ✅ 语义清晰，无需解释

**替代方案**：
- *Repositorium* (拉丁语：仓库) - 过于冗长
- *Arca* (拉丁语：箱子、柜子) - 生僻
- **Store** - 最佳选择 ✅

---

### 8. **AI** - 人工智能集成

- **词源**：英语 Artificial Intelligence
- **含义**：人工智能
- **职责**：LLM 集成、提示工程、流式生成
- **包路径**：`github.com/coreseekdev/texere/pkg/ai`

```go
import "github.com/coreseekdev/texere/pkg/ai"

llm := ai.NewLLM("gpt-4")
response := ai.Generate(llm, prompt)
```

**注**：AI 是技术通用术语，无需拉丁化。

---

## 🎯 完整命名体系

```
Texere (编织) - 主项目
│
├── Concordia (和谐) - OT 操作协调
├── Unio (统一) - 时间与版本统一
├── Textor (编织者) - 文本处理
├── Fabric (织物) - 文档结构
├── Weave (编织) - 核心引擎
├── Flux (流动) - 数据流与同步
├── Store (存储) - 持久化
└── AI (人工智能) - AI 集成
```

---

## 📊 命名原则总结

### ✅ 遵循的原则

1. **语言一致性**：全部使用拉丁语或拉丁语源词汇
2. **隐喻统一性**：围绕"编织"构建隐喻体系
3. **语义准确性**：名称准确描述包的职责
4. **文化深度**：优先选择有神话或历史背景的词汇
5. **简洁易记**：避免过长的词汇

### ❌ 避免的问题

1. ❌ **语言混杂**：希腊语（Syntaxis）+ 拉丁语（Texere）
2. ❌ **隐喻断裂**：纺织隐喻 + 其他无关隐喻
3. ❌ **过度技术化**：纯技术术语（如 OperationTransform）
4. ❌ **文化浅薄**：直白的描述性命名

---

## 🎨 品牌一致性

### Slogan

**主 Slogan**：
> **Texere: Weave Knowledge Together**
> 编织知识，连接智慧

**子产品 Slogan**：
- **Concordia**: *Harmony in Collaboration* (协作中的和谐)
- **Unio**: *Unify in Time* (时间中的统一)
- **Textor**: *The Text Weaver* (文本编织者)
- **Flux**: *Flow of Ideas* (思想的流动)

### 视觉元素

- 🧵 线与织物纹理
- 🕸️ 编织与连接的网络
- 🌊 流动与变化的水波
- ⏱️ 时间与秩序的钟表

---

## 📝 API 命名示例

### 包导入

```go
import (
    "github.com/coreseekdev/texere/pkg/concordia"  // OT
    "github.com/coreseekdev/texere/pkg/unio"       // 时间
    "github.com/coreseekdev/texere/pkg/textor"     // 文本
    "github.com/coreseekdev/texere/pkg/fabric"     // 文档
    "github.com/coreseekdev/texere/pkg/weave"      // 引擎
    "github.com/coreseekdev/texere/pkg/flux"       // 同步
)
```

### 函数命名

```go
// Concordia - OT 操作
concordia.NewInsert(pos, text)
concordia.Transform(op1, op2)
concordia.Compose(ops...)

// Unio - 时间排序
unio.NewLamportClock()
unio.Tick(clock)
unio.Compare(ts1, ts2)

// Textor - 文本处理
textor.NewRope(content)
textor.Insert(rope, pos, text)
textor.Delete(rope, pos, length)

// Fabric - 文档结构
fabric.NewDocument(id)
fabric.AddBlock(doc, block)
fabric.ApplyDelta(doc, delta)

// Weave - 编织引擎
weave.NewEngine(config)
weave.WeaveHuman(engine, op)
weave.WeaveAI(engine, request)

// Flux - 数据流动
flux.NewWebSocket()
flux.Subscribe(transport, topic)
flux.Publish(transport, msg)
```

---

## 🏆 命名体系的优势

### 1. 一致性 ✅
- 全部拉丁语，语言统一
- 编织隐喻贯穿始终
- 品牌识别度高

### 2. 可扩展性 ✅
- 可以继续添加拉丁语词汇的子包
- 命名模式清晰易懂
- 便于社区贡献

### 3. 文化深度 ✅
- Concordia (罗马女神)
- Unio (政治/哲学概念)
- Textor (历史职业)
- 有故事可讲

### 4. 国际化 ✅
- 拉丁语是欧洲语言的共同词根
- 在欧美技术圈认可度高
- 容易翻译和本地化

---

## 🎓 学习资源

### 拉丁语词汇表

| 拉丁语 | 英语 | 中文 | 应用 |
|--------|------|------|------|
| *Texere* | To weave | 编织 | 主项目 |
| *Concordia* | Harmony | 和谐 | OT 核心 |
| *Unio* | Union | 统一 | 时间排序 |
| *Textor* | Weaver | 编织者 | 文本处理 |
| *Fabrica* | Fabric | 织物 | 文档结构 |
| *Fluxus* | Flow | 流动 | 数据同步 |
| *Instaurare* | To store | 存储 | 持久化 |

### 推荐阅读

- *Latin for Beginners* - Benjamin L. D'Ooge
- *Word Power Made Easy* - Norman Lewis（词根词缀）
- *The Etymologicon* - Mark Forsyth（词源故事）

---

## 🔮 未来扩展

如果需要添加新的子包，可以考虑以下拉丁语词汇：

- **Notatio** (记号)：注释、评论系统
- **Versio** (版本)：版本控制
- **Copia** (丰富)：副本、备份
- **Index** (索引)：搜索与索引
- **Limen** (阈值)：权限与边界
- **Spatium** (空间)：文档空间管理
- **Tempus** (时间)：时间线管理
- **Vocabulum** (词汇)：词典与术语

---

**生成时间**：2026-01-28
**项目**：Texere - 文档编织引擎
**语言原则**：全部使用拉丁语或拉丁语源词汇
