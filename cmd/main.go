package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/coreseekdev/texere/pkg/session"
	"github.com/coreseekdev/texere/pkg/transport"
)

// UserInfoData extends session.UserInfo with color for demo.
type UserInfoData struct {
	*session.UserInfo
	Color string `json:"color"` // Random color for user
}

func main() {
	// Create components
	auth := session.NewTokenAuthenticator()
	content := session.NewMemoryContentStorage()

	// Initialize test files with Chinese and emoji content
	initializeTestFiles(content)

	// Create protocol handler
	protocolHandler := transport.NewProtocolHandler(content, auth)

	// Create a single HTTP mux for all routes
	mux := http.NewServeMux()

	// Create WebSocket server (without starting its own HTTP server)
	wsServer := transport.NewWebSocketServer("")
	protocolHandler.SetServer(wsServer)

	// Register WebSocket handler with our mux
	wsServer.RegisterHandler(mux)

	// Setup HTTP routes (edit page, etc.)
	setupHTTPRoutes(mux, protocolHandler, content, auth)

	// Create single HTTP server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		log.Println("Shutting down server...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		server.Shutdown(ctx)
		wsServer.Close()
		os.Exit(0)
	}()

	log.Println("==========================================")
	log.Println("  Texere Collaborative Editor Demo")
	log.Println("==========================================")
	log.Println("WebSocket server started on ws://localhost:8080/ws")
	log.Println("HTTP server started on http://localhost:8080")
	log.Println("")
	log.Println("Access the editor at:")
	log.Println("  http://localhost:8080/edit?token=user1")
	log.Println("  http://localhost:8080/edit?token=user2")
	log.Println("  http://localhost:8080/edit?token=user3")
	log.Println("")
	log.Println("Press Ctrl+C to stop")
	log.Println("==========================================")

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("Server error: %v", err)
	}
}

// initializeTestFiles creates test files with Chinese and emoji content.
func initializeTestFiles(store session.ContentStorage) {
	ctx := context.Background()

	testFiles := map[string]string{
		"test1.txt": `# 项目文档 📝

这是一个协同编辑的测试文档。欢迎大家同时编辑！

## 功能特性 ✨

- 实时协同编辑 ⚡
- 多用户支持 👥
- 操作转换 OT 🔄
- 冲突解决 🤝
- 历史记录 📚
- 版本回滚 ⏮️

## 中文支持

完美支持中文、日文、韩文等多字节字符。
测试各种边界情况和特殊字符。

## Emoji 支持 🎉

😀 😃 🄻 🌍 🎨 🎭 🎪 🎢 🎼 🎯

让我们开始协同编辑吧！🚀

---

**注意**: 这是一个演示文档，用于测试协同编辑功能。
`,
		"test2.txt": `# 代码编辑测试 💻

function helloWorld() {
    console.log("Hello, 世界! 🌏");
    return "Welcome to the collaborative editor! 🎉";
}

// 测试各种编辑场景
const data = [
    { name: "张三", age: 25, city: "北京" },
    { name: "李四", age: 30, city: "上海" },
    { name: "王五", age: 28, city: "深圳" },
];

console.log("用户列表:", data);

## 编辑操作 ⌨️

- ✅ 插入文本
- ✅ 删除内容
- ✅ 替换字符
- ✅ 多行编辑
- ✅ 代码高亮

## 验证中文字符

这是一段包含中英文混合的文本，用于测试 OT 算法的正确性。
The quick brown fox jumps over the lazy dog.
敏捷的棕色狐狸跳过了懒狗。

MathJax 公式测试: E = mc²

Good luck! 🍀
`,
		"test3.txt": `# 会议记录 📋

**时间**: 2026年1月31日 下午3:00
**地点**: 会议室 A 📍
**参与人**: 开发团队全体 👨‍💼👩‍💼

## 议程 📝

1. 项目进度汇报
2. 技术方案讨论
3. 下季度规划 🎯
4. 自由交流 ☕

## 讨论要点 💡

- 性能优化方案
- 用户体验改进
- 新功能开发计划
- 测试覆盖率提升

## 行动项 ✅

- [ ] 实现缓存优化（张三负责）
- [ ] 重构 API 接口（李四负责）
- [ ] 添加单元测试（王五负责）
- [ ] 更新文档（全员参与）

Emoji 测试: 🎊🎉🎈🎁🏆🥇🥈🥉🏅️🎂

让我们一起加油！💪

---

*下次会议时间: 下周五下午3点*
`,
	}

	// Create test files
	for filePath, fileContent := range testFiles {
		model := &session.ContentModel{
			Name:     filePath,
			Type:     "file",
			Format:   "text",
			MimeType: "text/plain",
			Path:     filePath,
			Content:  fileContent,
			Size:     int64(len(fileContent)),
		}

		if _, err := store.Save(ctx, filePath, model, nil); err != nil {
			log.Printf("Failed to create test file %s: %v", filePath, err)
		} else {
			log.Printf("Created test file: %s (%d bytes)", filePath, len(fileContent))
		}
	}
}

// setupHTTPRoutes configures HTTP routes.
func setupHTTPRoutes(mux *http.ServeMux, handler *transport.ProtocolHandler, content session.ContentStorage, auth session.Authenticator) {
	// Serve the editor page
	mux.HandleFunc("/edit", handleEditPage(content, auth))

	// API endpoint for creating tokens
	mux.HandleFunc("/api/token", handleCreateToken(auth))

	// Health check
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "ok",
			"message": "Texere Collaborative Editor Demo",
		})
	})
}

// handleEditPage serves the editor page.
func handleEditPage(content session.ContentStorage, auth session.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get token from query parameter
		token := r.URL.Query().Get("token")
		if token == "" {
			// If no token, generate one and redirect
			newToken, err := auth.GenerateToken(r.Context(), "user-"+randToken(6))
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Redirect with token
			http.Redirect(w, r, "/edit?token="+newToken, http.StatusFound)
			return
		}

	// Validate token and get user info
	_, userData, err := authenticateAndGetUser(auth, r.Context(), token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Render HTML page with inline CSS/JS
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, renderEditorPage(token, userData.Color))
	}
}

// handleCreateToken creates a new authentication token.
func handleCreateToken(auth session.Authenticator) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			UserID string `json:"user_id"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Generate token
		token, err := auth.GenerateToken(r.Context(), req.UserID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token":   token,
			"user_id": req.UserID,
		})
	}
}

// authenticateAndGetUser validates token and returns user info with color.
func authenticateAndGetUser(auth session.Authenticator, ctx context.Context, token string) (bool, *UserInfoData, error) {
	// Check if token exists and is valid
	valid, userInfo := auth.ValidateToken(ctx, token)
	if !valid {
		// Token doesn't exist, create new user
		userID := "user-" + randToken(6)
		_, err := auth.GenerateToken(ctx, userID)
		if err != nil {
			return false, nil, err
		}

		// Create user info
		user := &session.UserInfo{
			UserID:       userID,
			Username:     userID,
			Name:         userID,
			AuthProvider: "token",
			Metadata:     make(map[string]interface{}),
		}

		// Store color in metadata
		color := getRandomColor()
		user.Metadata["color"] = color

		// Return user info with color
		return true, &UserInfoData{
			UserInfo: user,
			Color:    color,
		}, nil
	}

	// Check if user has color, assign if not
	if userInfo != nil {
		color := getRandomColor()

		// Check if metadata has color, use it if exists
		if userInfo.Metadata != nil {
			if c, ok := userInfo.Metadata["color"].(string); ok && c != "" {
				color = c
			}
		}

		// Save color to metadata
		if userInfo.Metadata == nil {
			userInfo.Metadata = make(map[string]interface{})
		}
		userInfo.Metadata["color"] = color

		return true, &UserInfoData{
			UserInfo: userInfo,
			Color:    color,
		}, nil
	}

	return false, nil, fmt.Errorf("invalid token")
}

// renderEditorPage returns the HTML page with inline CSS/JS.
func renderEditorPage(token, userColor string) string {
	return `<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Texere 协同编辑演示</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }

        body {
            font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
            background: #f5f5f5;
            height: 100vh;
            overflow: hidden;
            margin: 0;
            padding: 0;
        }

        .header {
            background: white;
            border-bottom: 1px solid #ddd;
            padding: 6px 15px;
            display: flex;
            justify-content: space-between;
            align-items: center;
            box-shadow: 0 2px 4px rgba(0,0,0,0.1);
            min-height: 36px;
        }

        .header h1 {
            font-size: 14px;
            color: #333;
            margin: 0;
        }

        .user-info {
            display: flex;
            align-items: center;
            gap: 15px;
        }

        .user-badge {
            padding: 5px 12px;
            border-radius: 20px;
            background: #e0e0e0;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .user-color {
            width: 20px;
            height: 20px;
            border-radius: 50%;
            border: 2px solid rgba(0,0,0,0.1);
        }

        .connection-status {
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 13px;
            color: #666;
        }

        .status-dot {
            width: 10px;
            height: 10px;
            border-radius: 50%;
            background: #ccc;
        }

        .status-dot.connected {
            background: #4caf50;
        }

        .main {
            display: flex;
            height: calc(100vh - 36px);
        }

        .editor-container {
            flex: 1;
            display: flex;
            flex-direction: column;
            width: 100%;
            height: 100%; /* 100% height */
            min-height: 0; /* Important for flex child */
        }

        .editor-content {
            flex: 1;
            display: flex;
            flex-direction: column;
            height: 100%; /* 100% height */
            min-height: 0; /* Important for flex child */
        }

        .editor-header {
            background: #fafafa;
            border-bottom: 1px solid #ddd;
            padding: 4px 10px;
            font-size: 12px;
            color: #666;
            display: flex;
            justify-content: space-between;
            align-items: center;
            flex-shrink: 0; /* Don't grow */
        }

        .editor-tabs {
            display: flex;
            gap: 2px;
        }

        .editor-tab {
            padding: 4px 10px;
            background: #e0e0e0;
            border: none;
            border-radius: 4px 4px 0 0;
            cursor: pointer;
            font-size: 12px;
            color: #666;
            transition: all 0.2s;
        }

        .editor-tab:hover {
            background: #d0d0d0;
        }

        .editor-tab.active {
            background: #2196f3;
            color: white;
            font-weight: 500;
        }

        .editor-actions {
            display: flex;
            gap: 8px;
        }

        .editor-textarea {
            flex: 1;
            border: none;
            outline: none;
            padding: 20px;
            font-family: "Consolas", "Monaco", "Courier New", monospace;
            font-size: 16px;
            line-height: 1.8;
            resize: none;
            background: white;
            color: #333;
            min-height: 0; /* Important for flex child */
        }

        .editor-textarea:focus {
            background: #fafafa;
        }

        .sidebar {
            width: 250px;
            background: white;
            border-left: 1px solid #ddd;
            overflow-y: auto;
        }

        .sidebar-section {
            border-bottom: 1px solid #eee;
        }

        .sidebar-header {
            padding: 10px 15px;
            background: #fafafa;
            font-size: 12px;
            font-weight: 600;
            color: #666;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .user-list {
            padding: 10px;
        }

        .user-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 6px 8px;
            border-radius: 6px;
            margin-bottom: 4px;
            font-size: 13px;
        }

        .user-item:hover {
            background: #f5f5f5;
        }

        .info-section {
            padding: 15px;
            font-size: 12px;
            color: #666;
        }

        .info-item {
            margin-bottom: 8px;
        }

        .info-label {
            font-weight: 600;
            margin-bottom: 2px;
        }

        .file-list {
            padding: 10px;
        }

        .file-item {
            padding: 8px;
            border-radius: 6px;
            margin-bottom: 4px;
            cursor: pointer;
            font-size: 13px;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .file-item:hover {
            background: #f5f5f5;
        }

        .file-item.active {
            background: #e3f2fd;
        }

        .file-icon {
            font-size: 16px;
        }

        .toast {
            position: fixed;
            bottom: 20px;
            right: 20px;
            background: #323232;
            color: white;
            padding: 12px 20px;
            border-radius: 8px;
            font-size: 14px;
            box-shadow: 0 4px 12px rgba(0,0,0,0.15);
            z-index: 1000;
            opacity: 0;
            transition: opacity 0.3s;
            pointer-events: none;
        }

        .toast.show {
            opacity: 1;
        }

        .loading {
            display: inline-block;
            width: 16px;
            height: 16px;
            border: 2px solid #f3f3f3;
            border-top-color: #2196f3;
            border-radius: 50%;
            animation: spin 1s linear infinite;
        }

        @keyframes spin {
            to { transform: rotate(360deg); }
        }

        .editor-content {
            flex: 1;
            display: flex;
            flex-direction: column;
        }

        .editor-textarea-wrapper {
            flex: 1;
            position: relative;
            height: 100%; /* 100% height */
            min-height: 0; /* Important for flex child */
            display: flex;
            flex-direction: column;
        }

        .char-count {
            position: absolute;
            bottom: 10px;
            right: 15px;
            font-size: 12px;
            color: #999;
            background: rgba(255,255,255,0.9);
            padding: 2px 6px;
            border-radius: 4px;
        }

        .status-badge {
            display: inline-flex;
            align-items: center;
            gap: 4px;
            padding: 2px 8px;
            border-radius: 12px;
            background: #f0f0f0;
            font-size: 11px;
        }

        .remote-cursor {
            position: absolute;
            width: 2px;
            background: red;
            pointer-events: none;
            animation: blink 1s infinite;
        }

        @keyframes blink {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.3; }
        }

        .user-label {
            position: absolute;
            left: 4px;
            top: -20px;
            padding: 2px 6px;
            border-radius: 4px;
            font-size: 11px;
            color: white;
            white-space: nowrap;
        }
    </style>
</head>
<body>
    <div class="header">
        <h1>📝 Texere 协同编辑演示</h1>
        <div class="user-info">
            <div class="connection-status">
                <span class="status-dot"></span>
                <span id="status-text">连接中...</span>
            </div>
            <div class="user-badge">
                <span class="user-color" style="background-color: ` + userColor + `"></span>
                <span id="username">Loading...</span>
            </div>
        </div>
    </div>

    <div class="main">
        <div class="editor-container">
            <div class="editor-header">
                <div class="editor-tabs">
                    <button class="editor-tab active" data-file="test1.txt">
                        📄 Test 1
                    </button>
                    <button class="editor-tab" data-file="test2.txt">
                        📄 Test 2
                    </button>
                    <button class="editor-tab" data-file="test3.txt">
                        📄 Test 3
                    </button>
                </div>
                <div class="editor-actions">
                    <button id="save-btn" class="editor-tab">💾 保存</button>
                </div>
            </div>
            <div class="editor-content">
                <div class="editor-textarea-wrapper">
                    <textarea id="editor" class="editor-textarea" placeholder="正在加载内容..." disabled></textarea>
                    <div class="char-count"><span id="char-count">0</span> 字符</div>
                </div>
            </div>
        </div>

        <div class="sidebar">
            <div class="sidebar-section">
                <div class="sidebar-header">在线用户</div>
                <div class="user-list" id="user-list">
                    <div class="info-item" style="color: #999;">加载中...</div>
                </div>
            </div>
            <div class="sidebar-section">
                <div class="sidebar-header">文件信息</div>
                <div class="file-list">
                    <div class="file-item active" data-file="test1.txt">
                        <span class="file-icon">📄</span>
                        <span>Test 1</span>
                    </div>
                    <div class="file-item" data-file="test2.txt">
                        <span class="file-icon">💻</span>
                        <span>Test 2</span>
                    </div>
                    <div class="file-item" data-file="test3.txt">
                        <span class="file-icon">📋</span>
                        <span>Test 3</span>
                    </div>
                </div>
            </div>
            <div class="sidebar-section">
                <div class="sidebar-header">文档状态</div>
                <div class="info-section">
                    <div class="info-item">
                        <div class="info-label">当前文档</div>
                        <div id="current-file">test1.txt</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">版本号</div>
                        <div id="revision">0</div>
                    </div>
                    <div class="info-item">
                        <div class="info-label">连接用户</div>
                        <div id="user-count">0</div>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <div id="toast" class="toast"></div>

    <script>
        // Configuration
        const TOKEN = "` + token + `";
        const USER_COLOR = "` + userColor + `";
        const WS_URL = "ws://localhost:8080/ws";
        const CURRENT_FILE = "test1.txt";

        // State
        let ws = null;
        let currentFile = CURRENT_FILE;
        let content = "";
        let revision = 0;
        let sessionID = null;
        let clients = {};
        let snapshotLoaded = false;  // Track if snapshot has been loaded

        // DOM Elements
        const editor = document.getElementById("editor");
        const statusText = document.getElementById("status-text");
        const statusDot = document.querySelector(".status-dot");
        const username = document.getElementById("username");
        const charCount = document.getElementById("char-count");
        const userCount = document.getElementById("user-count");
        const currentFileEl = document.getElementById("current-file");
        const revisionEl = document.getElementById("revision");
        const userListEl = document.getElementById("user-list");
        const toast = document.getElementById("toast");

        // Initialize
        function init() {
            connectWebSocket();
            setupEditor();
            setupTabs();
            setupFileList();
        }

        // Load File (sends subscribe message)
        function loadFile(filePath) {
            currentFile = filePath;
            // Note: Actual subscription happens in WebSocket onopen
            console.log("Loading file:", filePath);
        }

        // WebSocket Connection
        function connectWebSocket() {
            ws = new WebSocket(WS_URL + "?client_id=" + TOKEN);

            ws.onopen = () => {
                setConnected(true);
                showToast("\u5df2\u8fde\u63a5\u5230\u670d\u52a1\u966a \u2705", "success");
                // Subscribe to the current file after connecting
                sendSubscribe(currentFile);
                sendHeartbeat();
            };

            ws.onmessage = (event) => {
                handleMessage(JSON.parse(event.data));
            };

            ws.onclose = () => {
                setConnected(false);
                showToast("连接已断开，正在重连...", "warning");
                // Reconnect after 3 seconds
                setTimeout(connectWebSocket, 3000);
            };

            ws.onerror = (error) => {
                console.error("WebSocket error:", error);
                showToast("连接错误 ❌", "error");
            };
        }

        function setConnected(connected) {
            if (connected) {
                statusDot.classList.add("connected");
                statusText.textContent = "已连接";
            } else {
                statusDot.classList.remove("connected");
                statusText.textContent = "未连接";
            }
        }

        // Message Handler
        function handleMessage(msg) {
            const protocolMsg = msg.metadata?.protocol_message;

            switch (protocolMsg?.type) {
                case "snapshot":
                    handleSnapshot(protocolMsg.data);
                    break;
                case "remote_operation":
                    handleRemoteOperation(protocolMsg.data);
                    break;
                case "ack":
                    handleAck(protocolMsg.data);
                    break;
                case "user_joined":
                    handleUserJoined(protocolMsg.data);
                    break;
                case "user_left":
                    handleUserLeft(protocolMsg.data);
                    break;
                case "session_info":
                    handleSessionInfo(protocolMsg.data);
                    break;
                case "error":
                    handleError(protocolMsg.data);
                    break;
                default:
                    console.log("Unknown message type:", protocolMsg?.type, msg);
            }
        }

        // Snapshot Handler
        function handleSnapshot(data) {
            console.log("[Snapshot] Received:", data);
            content = data.content || "";
            revision = data.revision || 0;
            sessionID = data.session_id;
            clients = {};

            // Mark snapshot as loaded
            snapshotLoaded = true;

            // Update username display
            if (username) {
                username.textContent = TOKEN;
            }

            // Build clients map
            if (data.clients) {
                data.clients.forEach(client => {
                    clients[client.client_id] = client;
                });
            }

            updateEditor();
            updateUserInfo();
            showToast("\u5df2\u52a0\u8f7d: " + data.file_path + " (v" + revision + ")");
        }

        // Remote Operation Handler
        function handleRemoteOperation(data) {
            console.log("[RemoteOp] Received:", data);

            // Apply operation to content
            try {
                // Simple OT operation application
                const newContent = applyOT(content, data.operation);
                if (newContent !== null) {
                    content = newContent;
                    revision = data.revision;
                    updateEditor();

                    // IMPORTANT: Sync tracking state with new editor value
                    // This ensures next local operation computes correct diff
                    window._editorLastTrackedContent = editor.value;
                    console.log("[RemoteOp] Synced tracking state to:", editor.value.length);

                    showToast("\u6536\u5230\u6765\u81ea " + data.client_id + " \u7684\u7f16\u8f91 \u26a1");
                }
            } catch (error) {
                console.error("Failed to apply remote operation:", error);
                showToast("\u5e94\u7528\u8fdc\u7a0b\u64cd\u4f5c\u5931\u8d25 \u274c", "error");
            }
        }

        // Apply OT operation to content
        function applyOT(textContent, operation) {
            if (!Array.isArray(operation) || operation.length === 0) {
                return textContent;
            }

            let result = textContent;
            let index = 0;

            for (let i = 0; i < operation.length; i++) {
                const op = operation[i];

                if (typeof op === "number") {
                    if (op > 0) {
                        // Retain
                        index += op;
                    } else {
                        // Delete
                        const deleteCount = -op;
                        result = result.substring(0, index) + result.substring(index + deleteCount);
                    }
                } else if (typeof op === "string") {
                    // Insert
                    result = result.substring(0, index) + op + result.substring(index);
                    index += op.length;
                }
            }

            return result;
        }

        // ACK Handler
        function handleAck(data) {
            revision = data.revision;
            updateRevision();
        }

        // User Joined Handler
        function handleUserJoined(data) {
            console.log("[UserJoined] User joined:", data);
            clients[data.client_id] = data.client;
            updateUserInfo();
            showToast(data.client.client + " \u52a0\u5165\u4e86\u534f\u4f5c \ud83c\udf89");
        }

        // User Left Handler
        function handleUserLeft(data) {
            console.log("[UserLeft] User left:", data);
            delete clients[data.client_id];
            updateUserInfo();
            showToast(data.client.client + " \u79bb\u5f00\u4e86\u534f\u4f5c \ud83d\udc4b");
        }

        // Session Info Handler
        function handleSessionInfo(data) {
            console.log("[SessionInfo] Session info:", data);
            updateUserInfo();
        }

        // Error Handler
        function handleError(data) {
            console.error("[Error] Server error:", data);
            showToast("\u9519\u8bef: " + data.message + " \u274c", "error");
        }

        // Send Operation
        function sendOperation(operation) {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                showToast("未连接到服务器 ❌", "error");
                return;
            }

            const message = {
                type: "operation",
                client_id: TOKEN,
                doc_id: currentFile,
                timestamp: Date.now(),
                metadata: {
                    protocol_message: {
                        type: "operation",
                        session_id: sessionID,
                        timestamp: Date.now(),
                        data: {
                            session_id: sessionID,
                            operation: operation,
                            selection: null,
                        }
                    }
                }
            };

            ws.send(JSON.stringify(message));
        }

        // Send Heartbeat
        function sendHeartbeat() {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                return;
            }

            const message = {
                type: "heartbeat",
                client_id: TOKEN,
                timestamp: Date.now(),
                metadata: {
                    protocol_message: {
                        type: "heartbeat",
                        timestamp: Date.now(),
                        data: {
                            session_ids: [sessionID]
                        }
                    }
                }
            };

            ws.send(JSON.stringify(message));

            // Send heartbeat every 30 seconds
            setTimeout(sendHeartbeat, 30000);
        }

        // Update Editor
        function updateEditor() {
            editor.value = content;
            editor.disabled = false; // 启用编辑器
            charCount.textContent = content.length;
            revisionEl.textContent = revision;

            // Sync tracking state when editor is updated externally (remote ops)
            if (window._editorLastTrackedContent !== undefined) {
                window._editorLastTrackedContent = content;
            }
        }

        // Update User Info
        function updateUserInfo() {
            const userArray = Object.values(clients);
            userCount.textContent = userArray.length;

            userListEl.innerHTML = userArray.map(client =>
                '<div class="user-item">' +
                    '<span class="user-color" style="background-color: ' + (client.color || USER_COLOR) + '"></span>' +
                    '<span>' + client.client_id + '</span>' +
                '</div>'
            ).join("") || '<div class="info-item" style="color: #999;">\u6682\u65e0\u5176\u4ed6\u7528\u6237</div>';
        }

        // Setup Editor
        function setupEditor() {
            let lastSentContent = null;  // Last content that was sent to server
            let typingTimer = null;

            editor.addEventListener("input", () => {
                // Don't process until snapshot is loaded
                if (!snapshotLoaded) {
                    console.log("[Editor] Ignoring input before snapshot loaded");
                    return;
                }

                const currentEditorValue = editor.value;

                // Initialize on first input with global content (from snapshot)
                if (lastSentContent === null) {
                    lastSentContent = content;  // Use snapshot content
                    window._editorLastTrackedContent = content;
                    console.log("[Editor] Initialized with content length:", lastSentContent.length);
                }

                // Clear existing timer (user is still typing)
                if (typingTimer) {
                    clearTimeout(typingTimer);
                }

                // Debounce: only compute and send operation after user stops typing for 500ms
                typingTimer = setTimeout(() => {
                    // User stopped typing - now compute the operation from last sent content
                    const oldContent = lastSentContent;
                    const newContent = editor.value;

                    // Only send if content actually changed
                    if (newContent !== oldContent) {
                        const operation = createOperation(newContent, oldContent);
                        if (operation) {
                            console.log("[Editor] Sending operation:", operation, {
                                oldLength: oldContent.length,
                                newLength: newContent.length
                            });
                            sendOperation(operation);

                            // Update tracking AFTER sending
                            lastSentContent = newContent;
                            content = newContent;
                            window._editorLastTrackedContent = content;
                        }
                    }
                }, 500);
            });
        }

        // Create OT operation from content diff
        function createOperation(newContent, oldContent) {
            // If content is same, return empty
            if (oldContent === newContent) {
                return null;
            }

            // Find common prefix
            const minLen = Math.min(oldContent.length, newContent.length);
            let prefixLen = 0;

            for (let i = 0; i < minLen; i++) {
                if (oldContent[i] !== newContent[i]) {
                    break;
                }
                prefixLen++;
            }

            // Find common suffix
            let suffixLen = 0;
            const maxSuffix = minLen - prefixLen;

            for (let i = 0; i < maxSuffix; i++) {
                if (oldContent[oldContent.length - 1 - i] !== newContent[newContent.length - 1 - i]) {
                    break;
                }
                suffixLen++;
            }

            // Extract the changed parts
            const oldMiddle = oldContent.substring(prefixLen, oldContent.length - suffixLen);
            const newMiddle = newContent.substring(prefixLen, newContent.length - suffixLen);

            console.log("[createOperation] Diff analysis:", {
                prefixLen,
                suffixLen,
                oldMiddle: JSON.stringify(oldMiddle.substring(0, 20)),
                newMiddle: JSON.stringify(newMiddle.substring(0, 20))
            });

            // Build OT operation
            const operation = [];

            // Retain common prefix
            if (prefixLen > 0) {
                operation.push(prefixLen);
            }

            // Delete old content (if any)
            if (oldMiddle.length > 0) {
                operation.push(-oldMiddle.length);  // Negative = delete
            }

            // Insert new content (if any)
            if (newMiddle.length > 0) {
                operation.push(newMiddle);  // String = insert
            }

            // Retain common suffix
            if (suffixLen > 0) {
                operation.push(suffixLen);
            }

            console.log("[createOperation] Generated operation:", operation);
            return operation;
        }

        // Setup Tabs
        function setupTabs() {
            document.querySelectorAll(".editor-tab[data-file]").forEach(tab => {
                tab.addEventListener("click", () => {
                    const file = tab.dataset.file;
                    if (file !== currentFile) {
                        switchFile(file);
                    }
                });
            });
        }

        // Setup File List
        function setupFileList() {
            document.querySelectorAll(".file-item").forEach(item => {
                item.addEventListener("click", () => {
                    const file = item.dataset.file;
                    if (file !== currentFile) {
                        switchFile(file);
                    }
                });
            });
        }

        // Switch File
        function switchFile(filePath) {
            currentFile = filePath;

            // Update active tab
            document.querySelectorAll(".editor-tab").forEach(tab => {
                tab.classList.remove("active");
                if (tab.dataset.file === filePath) {
                    tab.classList.add("active");
                }
            });

            // Update active file item
            document.querySelectorAll(".file-item").forEach(item => {
                item.classList.remove("active");
                if (item.dataset.file === filePath) {
                    item.classList.add("active");
                }
            });

            currentFileEl.textContent = filePath;

            // Send subscribe message
            sendSubscribe(filePath);
        }

        // Send Subscribe
        function sendSubscribe(filePath) {
            if (!ws || ws.readyState !== WebSocket.OPEN) {
                return;
            }

            const message = {
                type: "operation",
                client_id: TOKEN,
                doc_id: filePath,
                timestamp: Date.now(),
                metadata: {
                    protocol_message: {
                        type: "subscribe",
                        timestamp: Date.now(),
                        data: {
                            file_path: filePath,
                            read_only: false
                        }
                    }
                }
            };

            ws.send(JSON.stringify(message));
        }

        // Update Revision
        function updateRevision() {
            revisionEl.textContent = revision;
        }

        // Show Toast
        function showToast(message, type = "info") {
            toast.textContent = message;
            toast.style.background = type === "error" ? "#f44336" :
                                       type === "success" ? "#4caf50" :
                                       "#323232";
            toast.classList.add("show");

            setTimeout(() => {
                toast.classList.remove("show");
            }, 3000);
        }

        // Initialize on page load
        init();
    </script>
</body>
</html>
`
}

// randToken generates a random token string.
func randToken(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}

// getRandomColor returns a random color for user highlighting.
func getRandomColor() string {
	colors := []string{
		"#FF6B6B", // Red
		"#4ECDC4", // Green
		"#45B7D1", // Blue
	"#FFA07A", // Orange
	"#9B59B6", // Gray
	"#E91E63", // Pink
	"#00BCD4", // Cyan
	"#FF9800", // Amber
		"#795548", // Brown
		"#607D8B", // Blue Grey
		"#9C27B0", // Purple
		"#2196F3", // Light Blue
		"#FF5722", // Deep Orange
		"#795548", // Brown
		"#8BC34A", // Light Green
		"#03A9F4", // Light Blue
	"#CDDC39", // Lime
	"#FFEB3B", // Yellow
	"#9C27B0", // Purple
	"#FF9800", // Orange
	}

	return colors[rand.Intn(len(colors))]
}
