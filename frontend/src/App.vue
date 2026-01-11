<script setup>
import { ref, onMounted, watch } from 'vue'

// State management
const pbContent = ref('')
const output = ref('')
const fileName = ref('example.proto')
const isGenerating = ref(false)
const activeNav = ref('edit') // edit, generated, settings
const generatedFiles = ref([])
const pbFiles = ref([])
const settings = ref({
  fontSize: 14,
  autoSave: false
})
const selectedFile = ref(null)
const fileContentModal = ref(false)
const fileContent = ref('')
const searchQuery = ref('')

// User management state
const user = ref(null)
const isAuthenticated = ref(false)
const isUserMenuOpen = ref(false)
const isLoginModalOpen = ref(false)
const isRegisterModalOpen = ref(false)
const loginForm = ref({
  username: '',
  password: ''
})
const registerForm = ref({
  username: '',
  email: '',
  password: '',
  confirmPassword: ''
})
const authError = ref('')

// Auto-save timer
let autoSaveTimer = null

// Load protobuf content from file
async function loadPB(fileNameToLoad) {
  try {
    const result = await window['go']['main']['App']['ReadPB'](fileNameToLoad)
    // Check if result is an error message
    if (result.startsWith('Error')) {
      output.value = result
      // Use default content if file reading fails
      pbContent.value = `syntax = "proto3";

package example;

// 示例消息
message Example {
  string id = 1;
  string name = 2;
  int32 value = 3;
}

// 示例服务
service ExampleService {
  rpc GetExample (GetExampleRequest) returns (Example);
  rpc CreateExample (Example) returns (Example);
}

message GetExampleRequest {
  string id = 1;
}
`
    } else {
      pbContent.value = result
      output.value = `Loaded file: ${fileNameToLoad}`
    }
  } catch (e) {
    output.value = `Error loading file: ${e}`
    // Use default content if error occurs
    pbContent.value = `syntax = "proto3";

package example;

// Example message
message Example {
  string id = 1;
  string name = 2;
  int32 value = 3;
}

// Example service
service ExampleService {
  rpc GetExample (GetExampleRequest) returns (Example);
  rpc CreateExample (Example) returns (Example);
}

message GetExampleRequest {
  string id = 1;
}
`
  }
}

// Load protobuf content from file on mount
onMounted(async () => {
  await loadPB(fileName.value)
  await loadGeneratedFiles()
  await loadPBFiles()
  // Load current user
  await loadCurrentUser()
  // Apply initial settings
  applySettings()
})

// 监听设置变化 - 只处理自动保存，不处理主题变化
watch(() => settings.value.autoSave, (newAutoSave) => {
  if (newAutoSave) {
    setupAutoSave()
  } else {
    clearAutoSave()
  }
})

// 监听内容变化以实现自动保存
watch(pbContent, () => {
  if (settings.value.autoSave) {
    setupAutoSave()
  }
})

// 应用设置到UI
function applySettings() {
  // 应用字体大小到编辑器
  const editor = document.querySelector('.feishu-editor')
  if (editor) {
    editor.style.fontSize = `${settings.value.fontSize}px`
  }
}

// 设置自动保存计时器
function setupAutoSave() {
  if (autoSaveTimer) {
    clearTimeout(autoSaveTimer)
  }
  autoSaveTimer = setTimeout(() => {
    savePB()
  }, 3000) // 3秒无操作后保存
}

// 清除自动保存计时器
function clearAutoSave() {
  if (autoSaveTimer) {
    clearTimeout(autoSaveTimer)
    autoSaveTimer = null
  }
}

// 保存protobuf内容
async function savePB() {
  try {
    const result = await window['go']['main']['App']['SavePB'](fileName.value, pbContent.value)
    output.value = result
  } catch (e) {
    output.value = `错误: ${e}`
  }
}

// 生成GRPC代码
async function generateGRPC() {
  try {
    isGenerating.value = true
    const result = await window['go']['main']['App']['GenerateGRPC'](fileName.value, pbContent.value)
    output.value = result
    // 生成后重新加载生成的文件列表
    await loadGeneratedFiles()
  } catch (e) {
    output.value = `错误: ${e}`
  } finally {
    isGenerating.value = false
  }
}

// 加载生成的文件列表
async function loadGeneratedFiles() {
  try {
    const result = await window['go']['main']['App']['GetGeneratedFiles']()
    // 检查结果是否为文件数组
    if (Array.isArray(result)) {
      generatedFiles.value = result
    } else {
      // 如果是字符串，尝试解析为JSON
      try {
        generatedFiles.value = JSON.parse(result)
      } catch (parseError) {
        // 如果解析失败，使用虚拟数据
        generatedFiles.value = [
          {
            name: 'example.pb.go',
            size: 12606,
            modified: new Date().toLocaleString()
          },
          {
            name: 'example_grpc.pb.go',
            size: 11415,
            modified: new Date().toLocaleString()
          },
          {
            name: 'example.pb.gw.go',
            size: 5000,
            modified: new Date().toLocaleString()
          }
        ]
      }
    }
  } catch (e) {
    console.error('加载生成的文件错误:', e)
    // 如果有错误，使用虚拟数据
    generatedFiles.value = [
      {
        name: 'example.pb.go',
        size: 12606,
        modified: new Date().toLocaleString()
      },
      {
        name: 'example_grpc.pb.go',
        size: 11415,
        modified: new Date().toLocaleString()
      },
      {
        name: 'example.pb.gw.go',
        size: 5000,
        modified: new Date().toLocaleString()
      }
    ]
  }
}

// 在不同部分之间导航
function navigate(section) {
  activeNav.value = section
}

// 切换到不同的文件
async function switchFile(fileNameToSwitch) {
  fileName.value = fileNameToSwitch
  await loadPB(fileNameToSwitch)
}

// 更新设置
function updateSettings(newSettings) {
  settings.value = { ...settings.value, ...newSettings }
  output.value = `设置已更新: ${JSON.stringify(settings.value)}`
}

// 手动保存设置
function saveSettings() {
  // 应用设置到UI
  applySettings()
  // 处理自动保存
  if (settings.value.autoSave) {
    setupAutoSave()
  } else {
    clearAutoSave()
  }
  // 更新输出
  output.value = `设置保存成功: ${JSON.stringify(settings.value)}`
}

// 查看生成的文件内容
async function viewFile(file) {
  try {
    selectedFile.value = file
    const result = await window['go']['main']['App']['ReadGeneratedFile'](file.name)
    fileContent.value = result
    fileContentModal.value = true
  } catch (e) {
    output.value = `查看文件错误: ${e}`
  }
}

// 下载生成的文件
async function downloadFile(file) {
  try {
    const result = await window['go']['main']['App']['DownloadGeneratedFile'](file.name)
    // 从文件内容创建blob
    const blob = new Blob([result], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = file.name
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    output.value = `已下载文件: ${file.name}`
  } catch (e) {
    output.value = `下载文件错误: ${e}`
  }
}

// 关闭文件内容模态框
function closeModal() {
  fileContentModal.value = false
  selectedFile.value = null
  fileContent.value = ''
}

// 从目录加载protobuf文件
async function loadPBFiles() {
  try {
    const result = await window['go']['main']['App']['GetPBFiles']()
    // 检查结果是否为文件数组
    if (Array.isArray(result)) {
      pbFiles.value = result
    } else {
      // 如果是字符串，尝试解析为JSON
      try {
        pbFiles.value = JSON.parse(result)
      } catch (parseError) {
        // 如果解析失败，使用空数组
        pbFiles.value = []
      }
    }
  } catch (e) {
    console.error('加载protobuf文件错误:', e)
    pbFiles.value = []
  }
}

// 刷新生成的文件
async function refreshFiles() {
  await loadGeneratedFiles()
  output.value = '已刷新生成的文件列表'
}

// User authentication methods
async function loadCurrentUser() {
  try {
    const token = localStorage.getItem('authToken')
    if (token) {
      const result = await window['go']['main']['App']['GetCurrentUser'](token)
      const data = JSON.parse(result)
      if (data.success) {
        user.value = data.user
        isAuthenticated.value = true
      } else {
        localStorage.removeItem('authToken')
        isAuthenticated.value = false
        user.value = null
      }
    }
  } catch (e) {
    console.error('Error loading current user:', e)
    localStorage.removeItem('authToken')
    isAuthenticated.value = false
    user.value = null
  }
}

async function login() {
  try {
    authError.value = ''
    if (!loginForm.value.username || !loginForm.value.password) {
      authError.value = '请填写用户名和密码'
      return
    }
    
    const result = await window['go']['main']['App']['LoginUser'](
      loginForm.value.username,
      loginForm.value.password
    )
    const data = JSON.parse(result)
    if (data.success) {
      localStorage.setItem('authToken', data.token)
      user.value = data.user
      isAuthenticated.value = true
      isLoginModalOpen.value = false
      loginForm.value = {
        username: '',
        password: ''
      }
      output.value = '登录成功'
    } else {
      authError.value = data.error || '登录失败'
    }
  } catch (e) {
    authError.value = `登录错误: ${e}`
  }
}

async function register() {
  try {
    authError.value = ''
    if (!registerForm.value.username || !registerForm.value.email || !registerForm.value.password) {
      authError.value = '请填写所有必填字段'
      return
    }
    if (registerForm.value.password !== registerForm.value.confirmPassword) {
      authError.value = '密码不匹配'
      return
    }
    
    const result = await window['go']['main']['App']['RegisterUser'](
      registerForm.value.username,
      registerForm.value.email,
      registerForm.value.password
    )
    const data = JSON.parse(result)
    if (data.success) {
      localStorage.setItem('authToken', data.token)
      user.value = data.user
      isAuthenticated.value = true
      isRegisterModalOpen.value = false
      registerForm.value = {
        username: '',
        email: '',
        password: '',
        confirmPassword: ''
      }
      output.value = '注册成功'
    } else {
      authError.value = data.error || '注册失败'
    }
  } catch (e) {
    authError.value = `注册错误: ${e}`
  }
}

function logout() {
  try {
    const token = localStorage.getItem('authToken')
    if (token) {
      window['go']['main']['App']['LogoutUser'](token)
    }
    localStorage.removeItem('authToken')
    isAuthenticated.value = false
    user.value = null
    output.value = '已登出'
  } catch (e) {
    console.error('Error logging out:', e)
  }
}

// Modal management
function openLoginModal() {
  isLoginModalOpen.value = true
  isRegisterModalOpen.value = false
  authError.value = ''
}

function openRegisterModal() {
  isRegisterModalOpen.value = true
  isLoginModalOpen.value = false
  authError.value = ''
}

function closeAuthModals() {
  isLoginModalOpen.value = false
  isRegisterModalOpen.value = false
  authError.value = ''
  loginForm.value = {
    username: '',
    password: ''
  }
  registerForm.value = {
    username: '',
    email: '',
    password: '',
    confirmPassword: ''
  }
}
</script>

<template>
  <div class="feishu-app">
    <!-- Header -->
    <header class="feishu-header">
      <div class="feishu-header-content">
        <h1 class="feishu-app-title">Protobuf 编辑器</h1>
        <div class="feishu-header-actions">
          <template v-if="isAuthenticated">
            <div class="feishu-user-info dropdown">
              <div class="feishu-user-avatar" @click="isUserMenuOpen = !isUserMenuOpen">👤</div>
              <div class="feishu-user-name" @click="isUserMenuOpen = !isUserMenuOpen">{{ user?.username }}</div>
              <div v-if="isUserMenuOpen" class="feishu-user-menu">
                <div class="feishu-user-menu-item" @click="logout">退出登录</div>
              </div>
            </div>
          </template>
          <template v-else>
            <button @click="openLoginModal" class="feishu-btn feishu-btn-small feishu-btn-secondary" style="margin-right: 8px;">
              登录
            </button>
            <button @click="openRegisterModal" class="feishu-btn feishu-btn-small feishu-btn-primary">
              注册
            </button>
          </template>
        </div>
      </div>
    </header>
    
    <!-- Main Content -->
    <div class="feishu-main">
      <!-- Left Sidebar (Navigation) -->
      <aside class="feishu-sidebar">
        <div class="feishu-sidebar-section">
          <h3 class="feishu-sidebar-title">快速操作</h3>
          <ul class="feishu-nav-list">
            <li 
              class="feishu-nav-item" 
              :class="{ 'feishu-nav-item-active': activeNav === 'edit' }"
              @click="navigate('edit')"
            >
              <span class="feishu-nav-icon">📝</span>
              <span class="feishu-nav-text">编辑 Protobuf</span>
            </li>
            <li 
              class="feishu-nav-item" 
              :class="{ 'feishu-nav-item-active': activeNav === 'generated' }"
              @click="navigate('generated')"
            >
              <span class="feishu-nav-icon">📚</span>
              <span class="feishu-nav-text">生成的代码</span>
            </li>
            <li 
              class="feishu-nav-item" 
              :class="{ 'feishu-nav-item-active': activeNav === 'settings' }"
              @click="navigate('settings')"
            >
              <span class="feishu-nav-icon">⚙️</span>
              <span class="feishu-nav-text">设置</span>
            </li>
          </ul>
        </div>
        
        <div class="feishu-sidebar-section">
          <h3 class="feishu-sidebar-title">Protobuf 文件</h3>
          <ul class="feishu-file-list">
            <li 
              v-for="file in pbFiles" 
              :key="file.name"
              class="feishu-file-item" 
              :class="{ 'feishu-file-item-active': fileName === file.name }"
              @click="switchFile(file.name)"
            >
              <span class="feishu-file-icon">📄</span>
              <span class="feishu-file-name">{{ file.name }}</span>
            </li>
            <li v-if="pbFiles.length === 0" class="feishu-file-item feishu-file-item-empty">
              <span class="feishu-file-icon">📄</span>
              <span class="feishu-file-name">未找到 proto 文件</span>
            </li>
          </ul>
        </div>
      </aside>
      
      <!-- Right Content Area -->
      <main class="feishu-content">
        <!-- Edit Protobuf Section -->
        <template v-if="activeNav === 'edit'">
          <div class="feishu-content-header">
            <div class="feishu-breadcrumb">
            <span class="feishu-breadcrumb-item">首页</span>
            <span class="feishu-breadcrumb-separator">/</span>
            <span class="feishu-breadcrumb-item">Protobuf</span>
            <span class="feishu-breadcrumb-separator">/</span>
            <span class="feishu-breadcrumb-item">{{ fileName }}</span>
          </div>
          
          <div class="feishu-content-actions">
            <button 
              @click="savePB" 
              class="feishu-btn feishu-btn-primary"
              :disabled="!pbContent"
            >
              保存
            </button>
            <button 
              @click="generateGRPC" 
              class="feishu-btn feishu-btn-secondary"
              :disabled="!pbContent || isGenerating"
            >
              {{ isGenerating ? '生成中...' : '生成 GRPC' }}
            </button>
          </div>
          </div>
          
          <!-- Editor Card -->
          <div class="feishu-card">
            <div class="feishu-card-header">
              <h2 class="feishu-card-title">编辑 Protobuf</h2>
            </div>
            
            <div class="feishu-card-body">
              <div class="feishu-form-item">
                <div class="feishu-file-name-display">{{ fileName }}</div>
              </div>
              
              <div class="feishu-form-item feishu-form-item-large">
                <label class="feishu-form-label">Protobuf 内容</label>
                <div class="feishu-editor-container">
                  <textarea 
                    class="feishu-editor" 
                    v-model="pbContent" 
                    placeholder="在此编写你的 protobuf 代码..."
                    spellcheck="false"
                  ></textarea>
                </div>
              </div>
            </div>
          </div>
          
          <!-- 输出卡片 -->
          <div class="feishu-card feishu-card-margin-top">
            <div class="feishu-card-header">
              <h2 class="feishu-card-title">输出</h2>
            </div>
            
            <div class="feishu-card-body">
              <div class="feishu-output-container">
                <pre v-if="output" class="feishu-output">{{ output }}</pre>
                <div v-else class="feishu-output-placeholder">
                  <div class="feishu-output-placeholder-icon">📤</div>
                  <h4 class="feishu-output-placeholder-title">暂无输出</h4>
                  <p class="feishu-output-placeholder-desc">
                    保存 Protobuf 或生成 GRPC 代码以查看结果
                  </p>
                </div>
              </div>
            </div>
          </div>
        </template>
        
        <!-- Generated Code Section -->
        <template v-else-if="activeNav === 'generated'">
          <div class="feishu-content-header">
            <div class="feishu-breadcrumb">
            <span class="feishu-breadcrumb-item">首页</span>
            <span class="feishu-breadcrumb-separator">/</span>
            <span class="feishu-breadcrumb-item">生成的代码</span>
          </div>
          
          <div class="feishu-content-actions">
            <button 
              @click="refreshFiles" 
              class="feishu-btn feishu-btn-secondary"
            >
              🔄 刷新
            </button>
            <button 
              @click="generateGRPC" 
              class="feishu-btn feishu-btn-primary"
              :disabled="!pbContent || isGenerating"
            >
              {{ isGenerating ? '生成中...' : '重新生成代码' }}
            </button>
          </div>
          </div>
          
          <!-- Generated Files Card -->
          <div class="feishu-card">
            <div class="feishu-card-header">
              <h2 class="feishu-card-title">生成的文件</h2>
              <p class="feishu-card-subtitle">从 {{ fileName }} 生成的文件</p>
            </div>
            
            <div class="feishu-card-body">
              <!-- 搜索框 -->
              <div class="feishu-search-container">
                <input 
                  type="text" 
                  v-model="searchQuery" 
                  placeholder="搜索文件..." 
                  class="feishu-search-input"
                />
              </div>
              
              <!-- Files Table -->
              <div class="feishu-table-container">
                <table class="feishu-table">
                  <thead class="feishu-table-header">
                    <tr>
                      <th class="feishu-table-th">文件名</th>
                      <th class="feishu-table-th">大小</th>
                      <th class="feishu-table-th">修改时间</th>
                      <th class="feishu-table-th">操作</th>
                    </tr>
                  </thead>
                  <tbody class="feishu-table-body">
                    <tr 
                      v-for="file in generatedFiles" 
                      :key="file.name" 
                      class="feishu-table-row"
                      v-if="!searchQuery || file.name.includes(searchQuery)"
                    >
                      <td class="feishu-table-td">
                        <span class="feishu-file-icon">📄</span>
                        <span class="feishu-table-file-name">{{ file.name }}</span>
                      </td>
                      <td class="feishu-table-td">{{ (file.size / 1024).toFixed(2) }} KB</td>
                      <td class="feishu-table-td">{{ file.modified }}</td>
                      <td class="feishu-table-td">
                        <button 
                          class="feishu-btn feishu-btn-small feishu-btn-secondary"
                          @click="viewFile(file)"
                          style="margin-right: 8px;"
                        >
                          查看
                        </button>
                        <button 
                          class="feishu-btn feishu-btn-small feishu-btn-secondary"
                          @click="downloadFile(file)"
                        >
                          下载
                        </button>
                      </td>
                    </tr>
                    <tr v-if="generatedFiles.length === 0" class="feishu-table-row feishu-table-empty">
                      <td colspan="4" class="feishu-table-td">
                        <div class="feishu-empty-state">
                          <span class="feishu-empty-icon">📁</span>
                          <p class="feishu-empty-text">未找到生成的文件</p>
                          <p class="feishu-empty-hint">点击"重新生成代码"以生成文件</p>
                        </div>
                      </td>
                    </tr>
                    <tr v-else-if="generatedFiles.filter(file => file.name.includes(searchQuery)).length === 0" class="feishu-table-row feishu-table-empty">
                      <td colspan="4" class="feishu-table-td">
                        <div class="feishu-empty-state">
                          <span class="feishu-empty-icon">🔍</span>
                          <p class="feishu-empty-text">没有匹配的文件</p>
                          <p class="feishu-empty-hint">尝试其他搜索词</p>
                        </div>
                      </td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>
          </div>
        </template>
        
        <!-- Settings Section -->
        <template v-else-if="activeNav === 'settings'">
          <div class="feishu-content-header">
            <div class="feishu-breadcrumb">
            <span class="feishu-breadcrumb-item">首页</span>
            <span class="feishu-breadcrumb-separator">/</span>
            <span class="feishu-breadcrumb-item">设置</span>
          </div>
          </div>
          
          <!-- Settings Card -->
          <div class="feishu-card">
            <div class="feishu-card-header">
              <h2 class="feishu-card-title">设置</h2>
              <p class="feishu-card-subtitle">自定义你的 Protobuf 编辑器体验</p>
            </div>
            
            <div class="feishu-card-body">
              <div class="feishu-settings-section">
                <h3 class="feishu-settings-section-title">编辑器设置</h3>
                
                <div class="feishu-form-item">
                  <label class="feishu-form-label">字体大小</label>
                  <div class="feishu-font-size-container">
                    <input 
                      type="range" 
                      v-model.number="settings.fontSize" 
                      min="10" 
                      max="24" 
                      class="feishu-slider"
                    />
                    <span class="feishu-font-size-value">{{ settings.fontSize }}px</span>
                  </div>
                </div>
                
                <div class="feishu-form-item">
                  <label class="feishu-form-label">自动保存</label>
                  <div class="feishu-switch-container">
                    <label class="feishu-switch">
                      <input 
                        type="checkbox" 
                        v-model="settings.autoSave"
                      />
                      <span class="feishu-switch-slider"></span>
                    </label>
                    <span class="feishu-switch-label">{{ settings.autoSave ? '已启用' : '已禁用' }}</span>
                  </div>
                </div>
              </div>
              
              <div class="feishu-settings-actions">
                <button 
                  @click="saveSettings" 
                  class="feishu-btn feishu-btn-primary"
                >
                  保存设置
                </button>
              </div>
            </div>
          </div>
        </template>
      </main>
    </div>
    
    <!-- File Content Modal -->
    <div v-if="fileContentModal" class="feishu-modal-overlay" @click="closeModal">
      <div class="feishu-modal" @click.stop>
        <div class="feishu-modal-header">
          <h3 class="feishu-modal-title">{{ selectedFile?.name }}</h3>
          <button class="feishu-modal-close" @click="closeModal" title="关闭">×</button>
        </div>
        <div class="feishu-modal-body">
          <pre class="feishu-file-content">{{ fileContent }}</pre>
        </div>
        <div class="feishu-modal-footer">
          <button class="feishu-btn feishu-btn-secondary" @click="closeModal">关闭</button>
        </div>
      </div>
    </div>

    <!-- Login Modal -->
    <div v-if="isLoginModalOpen" class="feishu-modal-overlay" @click="closeAuthModals">
      <div class="feishu-modal" @click.stop>
        <div class="feishu-modal-header">
          <h3 class="feishu-modal-title">登录</h3>
          <button class="feishu-modal-close" @click="closeAuthModals" title="关闭">×</button>
        </div>
        <div class="feishu-modal-body">
          <div class="feishu-auth-form">
            <div v-if="authError" class="feishu-auth-error">{{ authError }}</div>
            <div class="feishu-form-item">
              <label class="feishu-form-label">用户名</label>
              <input 
                type="text" 
                v-model="loginForm.username" 
                class="feishu-input" 
                placeholder="请输入用户名"
              >
            </div>
            <div class="feishu-form-item">
              <label class="feishu-form-label">密码</label>
              <input 
                type="password" 
                v-model="loginForm.password" 
                class="feishu-input" 
                placeholder="请输入密码"
              >
            </div>
          </div>
        </div>
        <div class="feishu-modal-footer">
          <button class="feishu-btn feishu-btn-secondary" @click="closeAuthModals">取消</button>
          <button class="feishu-btn feishu-btn-primary" @click="login">登录</button>
        </div>
      </div>
    </div>

    <!-- Register Modal -->
    <div v-if="isRegisterModalOpen" class="feishu-modal-overlay" @click="closeAuthModals">
      <div class="feishu-modal" @click.stop>
        <div class="feishu-modal-header">
          <h3 class="feishu-modal-title">注册</h3>
          <button class="feishu-modal-close" @click="closeAuthModals" title="关闭">×</button>
        </div>
        <div class="feishu-modal-body">
          <div class="feishu-auth-form">
            <div v-if="authError" class="feishu-auth-error">{{ authError }}</div>
            <div class="feishu-form-item">
              <label class="feishu-form-label">用户名</label>
              <input 
                type="text" 
                v-model="registerForm.username" 
                class="feishu-input" 
                placeholder="请输入用户名"
              >
            </div>
            <div class="feishu-form-item">
              <label class="feishu-form-label">邮箱</label>
              <input 
                type="email" 
                v-model="registerForm.email" 
                class="feishu-input" 
                placeholder="请输入邮箱"
              >
            </div>
            <div class="feishu-form-item">
              <label class="feishu-form-label">密码</label>
              <input 
                type="password" 
                v-model="registerForm.password" 
                class="feishu-input" 
                placeholder="请输入密码"
              >
            </div>
            <div class="feishu-form-item">
              <label class="feishu-form-label">确认密码</label>
              <input 
                type="password" 
                v-model="registerForm.confirmPassword" 
                class="feishu-input" 
                placeholder="请再次输入密码"
              >
            </div>
          </div>
        </div>
        <div class="feishu-modal-footer">
          <button class="feishu-btn feishu-btn-secondary" @click="closeAuthModals">取消</button>
          <button class="feishu-btn feishu-btn-primary" @click="register">注册</button>
        </div>
      </div>
    </div>
  </div>
</template>

<style>
/* Global styles */
body {
  margin: 0;
  padding: 0;
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
  background-color: #f5f5f5;
  color: #212529;
}

/* Feishu Design System Styles */
:root {
  --feishu-primary-color: #0078d4;
  --feishu-primary-hover: #106ebe;
  --feishu-secondary-color: #6c757d;
  --feishu-secondary-hover: #5a6268;
  --feishu-bg-color: #f5f5f5;
  --feishu-card-bg: #ffffff;
  --feishu-text-primary: #212529;
  --feishu-text-secondary: #6c757d;
  --feishu-border-color: #e0e0e0;
  --feishu-header-bg: #ffffff;
  --feishu-sidebar-bg: #ffffff;
  --feishu-sidebar-hover: #f0f0f0;
  --feishu-sidebar-active: #e8f0fe;
  --feishu-sidebar-active-text: #0078d4;
  --feishu-input-bg: #ffffff;
  --feishu-input-border: #e0e0e0;
  --feishu-input-focus: #0078d4;
  --feishu-btn-primary: #0078d4;
  --feishu-btn-primary-hover: #106ebe;
  --feishu-btn-secondary: #ffffff;
  --feishu-btn-secondary-hover: #f5f5f5;
  --feishu-btn-secondary-border: #e0e0e0;
  --feishu-btn-secondary-text: #212529;
  --feishu-btn-disabled: #e0e0e0;
  --feishu-card-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  --feishu-modal-overlay: rgba(0, 0, 0, 0.5);
  --feishu-slider-track: #e0e0e0;
  --feishu-slider-thumb: #0078d4;
}
</style>

<style scoped>
/* Reset */
* {
  box-sizing: border-box;
  margin: 0;
  padding: 0;
}

/* App layout */
.feishu-app {
  display: flex;
  flex-direction: column;
  height: 100vh;
  width: 100vw;
  background-color: var(--feishu-bg-color);
  color: var(--feishu-text-primary);
  font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
}

/* Header */
.feishu-header {
  background-color: var(--feishu-header-bg);
  border-bottom: 1px solid var(--feishu-border-color);
  padding: 0 20px;
  height: 60px;
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.feishu-header-content {
  display: flex;
  align-items: center;
  justify-content: space-between;
  width: 100%;
}

.feishu-app-title {
  font-size: 20px;
  font-weight: 600;
  color: var(--feishu-text-primary);
  text-align: left;
}

.feishu-header-actions {
  display: flex;
  align-items: center;
  gap: 12px;
}

.feishu-user-info {
  display: flex;
  align-items: center;
  gap: 8px;
  position: relative;
}

.feishu-user-menu {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: 8px;
  background-color: white;
  border: 1px solid var(--feishu-border-color);
  border-radius: 6px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  min-width: 120px;
  z-index: 1000;
}

.feishu-user-menu-item {
  padding: 10px 16px;
  cursor: pointer;
  transition: background-color 0.2s ease;
  font-size: 14px;
  text-align: left;
}

.feishu-user-menu-item:hover {
  background-color: var(--feishu-sidebar-hover);
}

/* Auth Form Styles */
.feishu-auth-form {
  max-width: 400px;
  margin: 0 auto;
}

.feishu-auth-error {
  background-color: #fee;
  color: #d32f2f;
  padding: 8px 12px;
  border-radius: 6px;
  margin-bottom: 16px;
  font-size: 14px;
}

.feishu-auth-form .feishu-form-item {
  margin-bottom: 16px;
}

.feishu-user-avatar {
  font-size: 24px;
}

.feishu-user-name {
  font-size: 14px;
  color: var(--feishu-text-secondary);
}

/* Main layout */
.feishu-main {
  display: flex;
  flex: 1;
  overflow: hidden;
}

/* Sidebar */
.feishu-sidebar {
  width: 240px;
  background-color: var(--feishu-sidebar-bg);
  border-right: 1px solid var(--feishu-border-color);
  padding: 20px 0;
  overflow-y: auto;
}

.feishu-sidebar-section {
  margin-bottom: 24px;
}

.feishu-sidebar-title {
  font-size: 12px;
  font-weight: 600;
  color: var(--feishu-text-secondary);
  padding: 0 20px 8px;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  text-align: left;
}

.feishu-nav-list {
  list-style: none;
}

.feishu-nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 12px 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 0 20px 20px 0;
}

.feishu-nav-item:hover {
  background-color: var(--feishu-sidebar-hover);
}

.feishu-nav-item-active {
  background-color: var(--feishu-sidebar-active);
  color: var(--feishu-sidebar-active-text);
  font-weight: 500;
}

.feishu-nav-icon {
  font-size: 18px;
  width: 20px;
  text-align: center;
}

.feishu-nav-text {
  font-size: 14px;
}

.feishu-file-list {
  list-style: none;
}

.feishu-file-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 20px;
  cursor: pointer;
  transition: all 0.2s ease;
  border-radius: 0 20px 20px 0;
  font-size: 14px;
}

.feishu-file-item:hover {
  background-color: var(--feishu-sidebar-hover);
}

.feishu-file-item-active {
  background-color: var(--feishu-sidebar-active);
  color: var(--feishu-sidebar-active-text);
  font-weight: 500;
}

.feishu-file-icon {
  font-size: 16px;
}

.feishu-file-name {
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* Content area */
.feishu-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}

/* Content header */
.feishu-content-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 20px;
}

.feishu-breadcrumb {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 14px;
  color: var(--feishu-text-secondary);
}

.feishu-breadcrumb-item {
  cursor: pointer;
  transition: color 0.2s ease;
}

.feishu-breadcrumb-item:hover {
  color: var(--feishu-primary-color);
}

.feishu-breadcrumb-separator {
  color: var(--feishu-text-secondary);
}

.feishu-content-actions {
  display: flex;
  gap: 12px;
}

/* Buttons */
.feishu-btn {
  padding: 8px 16px;
  border: none;
  border-radius: 6px;
  font-size: 14px;
  font-weight: 500;
  cursor: pointer;
  transition: all 0.2s ease;
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.feishu-btn:disabled {
  opacity: 0.6;
  cursor: not-allowed;
}

.feishu-btn-primary {
  background-color: #0078d4;
  color: white;
  border: none;
}

.feishu-btn-primary:hover:not(:disabled) {
  background-color: #106ebe;
}

.feishu-btn-secondary {
  background-color: #ffffff;
  color: #212529;
  border: 1px solid #e0e0e0;
}

.feishu-btn-secondary:hover:not(:disabled) {
  background-color: #f5f5f5;
}

.feishu-btn-small {
  padding: 6px 12px;
  font-size: 12px;
}

/* Cards */
.feishu-card {
  background-color: var(--feishu-card-bg);
  border-radius: 8px;
  box-shadow: var(--feishu-card-shadow);
  margin-bottom: 20px;
  overflow: hidden;
}

.feishu-card-margin-top {
  margin-top: 20px;
}

.feishu-card-header {
  padding: 20px;
  border-bottom: 1px solid var(--feishu-border-color);
}

.feishu-card-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--feishu-text-primary);
  margin-bottom: 4px;
  text-align: left;
}

.feishu-card-subtitle {
  font-size: 14px;
  color: var(--feishu-text-secondary);
  text-align: left;
}

.feishu-card-body {
  padding: 20px;
}

/* Forms */
.feishu-form-item {
  margin-bottom: 20px;
}

.feishu-form-item-large {
  margin-bottom: 0;
}

.feishu-form-label {
  display: block;
  font-size: 14px;
  font-weight: 500;
  color: var(--feishu-text-primary);
  margin-bottom: 8px;
}

.feishu-file-name-display {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--feishu-input-border);
  border-radius: 6px;
  font-size: 14px;
  background-color: var(--feishu-input-bg);
  color: var(--feishu-text-primary);
  font-weight: 500;
  cursor: default;
}

.feishu-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--feishu-input-border);
  border-radius: 6px;
  font-size: 14px;
  background-color: var(--feishu-input-bg);
  color: var(--feishu-text-primary);
  transition: all 0.2s ease;
}

.feishu-input:focus {
  outline: none;
  border-color: var(--feishu-input-focus);
  box-shadow: 0 0 0 2px rgba(0, 120, 212, 0.2);
}

/* Editor */
.feishu-editor-container {
  position: relative;
  border: 1px solid var(--feishu-input-border);
  border-radius: 6px;
  overflow: hidden;
  min-height: 400px;
}

.feishu-editor {
  width: 100%;
  height: 400px;
  padding: 12px;
  border: none;
  resize: vertical;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  background-color: var(--feishu-input-bg);
  color: var(--feishu-text-primary);
  outline: none;
}

.feishu-editor:focus {
  box-shadow: 0 0 0 2px rgba(0, 120, 212, 0.2);
}

/* Output */
.feishu-output-container {
  background-color: var(--feishu-input-bg);
  border: 1px solid var(--feishu-input-border);
  border-radius: 6px;
  padding: 12px;
  min-height: 200px;
}

.feishu-output {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: var(--feishu-text-primary);
  white-space: pre-wrap;
  word-break: break-all;
  margin: 0;
}

.feishu-output-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  color: var(--feishu-text-secondary);
  text-align: center;
}

.feishu-output-placeholder-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.feishu-output-placeholder-title {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
}

.feishu-output-placeholder-desc {
  font-size: 14px;
  opacity: 0.8;
}

/* Search */
.feishu-search-container {
  margin-bottom: 16px;
}

.feishu-search-input {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid var(--feishu-input-border);
  border-radius: 6px;
  font-size: 14px;
  background-color: var(--feishu-input-bg);
  color: var(--feishu-text-primary);
  transition: all 0.2s ease;
}

.feishu-search-input:focus {
  outline: none;
  border-color: var(--feishu-input-focus);
  box-shadow: 0 0 0 2px rgba(0, 120, 212, 0.2);
}

/* Tables */
.feishu-table-container {
  overflow-x: auto;
}

.feishu-table {
  width: 100%;
  border-collapse: collapse;
}

.feishu-table-header {
  background-color: var(--feishu-bg-color);
}

.feishu-table-th {
  padding: 12px 16px;
  text-align: left;
  font-size: 14px;
  font-weight: 600;
  color: var(--feishu-text-primary);
  border-bottom: 2px solid var(--feishu-border-color);
}

.feishu-table-row {
  border-bottom: 1px solid var(--feishu-border-color);
  transition: background-color 0.2s ease;
}

.feishu-table-row:hover {
  background-color: var(--feishu-bg-color);
}

.feishu-table-td {
  padding: 12px 16px;
  font-size: 14px;
  color: var(--feishu-text-primary);
  text-align: left;
}

.feishu-table-file-name {
  font-weight: 500;
}

.feishu-table-empty {
  text-align: center;
}

.feishu-empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 40px 0;
  color: var(--feishu-text-secondary);
}

.feishu-empty-icon {
  font-size: 48px;
  margin-bottom: 16px;
  opacity: 0.5;
}

.feishu-empty-text {
  font-size: 16px;
  font-weight: 500;
  margin-bottom: 8px;
}

.feishu-empty-hint {
  font-size: 14px;
  opacity: 0.8;
}

/* Settings */
.feishu-settings-section {
  margin-bottom: 24px;
  padding-bottom: 24px;
  border-bottom: 1px solid var(--feishu-border-color);
}

.feishu-settings-section:last-child {
  margin-bottom: 0;
  padding-bottom: 0;
  border-bottom: none;
}

.feishu-settings-section-title {
  font-size: 16px;
  font-weight: 600;
  color: var(--feishu-text-primary);
  margin-bottom: 16px;
  text-align: left;
}

.feishu-font-size-container {
  display: flex;
  align-items: center;
  gap: 16px;
}

.feishu-slider {
  flex: 1;
  -webkit-appearance: none;
  appearance: none;
  height: 6px;
  background: var(--feishu-slider-track);
  border-radius: 3px;
  outline: none;
}

.feishu-slider::-webkit-slider-thumb {
  -webkit-appearance: none;
  appearance: none;
  width: 18px;
  height: 18px;
  background: var(--feishu-slider-thumb);
  border-radius: 50%;
  cursor: pointer;
  transition: all 0.2s ease;
}

.feishu-slider::-webkit-slider-thumb:hover {
  transform: scale(1.1);
}

.feishu-slider::-moz-range-thumb {
  width: 18px;
  height: 18px;
  background: var(--feishu-slider-thumb);
  border-radius: 50%;
  cursor: pointer;
  border: none;
  transition: all 0.2s ease;
}

.feishu-slider::-moz-range-thumb:hover {
  transform: scale(1.1);
}

.feishu-font-size-value {
  font-size: 14px;
  font-weight: 500;
  color: var(--feishu-text-primary);
  min-width: 50px;
}

.feishu-switch-container {
  display: flex;
  align-items: center;
  gap: 12px;
}

.feishu-switch {
  position: relative;
  display: inline-block;
  width: 40px;
  height: 20px;
}

.feishu-switch input {
  opacity: 0;
  width: 0;
  height: 0;
}

.feishu-switch-slider {
  position: absolute;
  cursor: pointer;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: #ccc;
  transition: .4s;
  border-radius: 20px;
}

.feishu-switch-slider:before {
  position: absolute;
  content: "";
  height: 16px;
  width: 16px;
  left: 2px;
  bottom: 2px;
  background-color: white;
  transition: .4s;
  border-radius: 50%;
}

input:checked + .feishu-switch-slider {
  background-color: var(--feishu-primary-color);
}

input:focus + .feishu-switch-slider {
  box-shadow: 0 0 1px var(--feishu-primary-color);
}

input:checked + .feishu-switch-slider:before {
  transform: translateX(20px);
}

.feishu-switch-label {
  font-size: 14px;
  color: var(--feishu-text-primary);
}

.feishu-settings-actions {
  display: flex;
  justify-content: flex-end;
  margin-top: 24px;
  padding-top: 24px;
  border-top: 1px solid var(--feishu-border-color);
}

/* Modal */
.feishu-modal-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  bottom: 0;
  background-color: var(--feishu-modal-overlay);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 1000;
}

.feishu-modal {
  background-color: var(--feishu-card-bg);
  border-radius: 8px;
  box-shadow: 0 4px 24px rgba(0, 0, 0, 0.15);
  width: 90%;
  max-width: 800px;
  max-height: 90vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.feishu-modal-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 16px 20px;
  border-bottom: 1px solid var(--feishu-border-color);
}

.feishu-modal-title {
  font-size: 18px;
  font-weight: 600;
  color: var(--feishu-text-primary);
  text-align: left;
}

.feishu-modal-close {
  background: none;
  border: none;
  font-size: 24px;
  cursor: pointer;
  color: var(--feishu-text-secondary);
  padding: 0;
  width: 24px;
  height: 24px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-radius: 6px;
  transition: all 0.2s ease;
}

.feishu-modal-close:hover {
  background-color: var(--feishu-bg-color);
  color: var(--feishu-text-primary);
}

.feishu-modal-body {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  text-align: left;
}

.feishu-file-content {
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.5;
  color: var(--feishu-text-primary);
  background-color: var(--feishu-bg-color);
  padding: 16px;
  border-radius: 4px;
  white-space: pre;
  word-break: normal;
  overflow-x: auto;
  overflow-y: auto;
  max-height: 60vh;
  text-align: left;
  display: block;
}

.feishu-modal-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
  padding: 16px 20px;
  border-top: 1px solid var(--feishu-border-color);
}
</style>