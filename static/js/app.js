// API 配置
const API_BASE = '/api';
let API_KEY =  '';

// 显示/隐藏标签页
function showTab(tabName) {
    // 更新标签按钮
    document.querySelectorAll('.tab-btn').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    // 显示对应内容
    document.querySelectorAll('.tab-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(`${tabName}-tab`).classList.add('active');
}

// 显示管理后台标签页
function showDashboardTab(tabName) {
    document.querySelectorAll('.dashboard-tab').forEach(btn => {
        btn.classList.remove('active');
    });
    event.target.classList.add('active');
    
    document.querySelectorAll('.dashboard-content').forEach(content => {
        content.classList.remove('active');
    });
    document.getElementById(`${tabName}-tab`).classList.add('active');
    
    // 加载对应数据
    switch(tabName) {
        case 'users':
            loadUsers();
            break;
        case 'aliases':
            loadAliases();
            break;
        case 'fail2ban':
            loadFail2ban();
            break;
        case 'stats':
            loadStats();
            break;
    }
}

// ========== 注册功能 ==========
document.getElementById('registerForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const email = document.getElementById('register-email').value;
    const password = document.getElementById('register-password').value;
    const confirm = document.getElementById('register-confirm').value;
    
    // 验证邮箱域名
    if (!email.endsWith('@yxliuchn.uk')) {
        showMessage('register', '只能注册 @yxliuchn.uk 域名的邮箱', 'error');
        return;
    }
    
    if (password !== confirm) {
        showMessage('register', '两次输入的密码不一致', 'error');
        return;
    }
    
    if (password.length < 8) {
        showMessage('register', '密码至少需要8位', 'error');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/email`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': 'DNtWPAaJwWocqvKR3o8ysBdJToSQHWXW'
            },
            body: JSON.stringify({
                address: email,
                password: password
            })
        });
        
        if (response.ok) {
            showMessage('register', '注册成功！请前往登录', 'success');
            document.getElementById('registerForm').reset();
            setTimeout(() => showTab('login'), 2000);
        } else {
            const error = await response.text();
            showMessage('register', `注册失败: ${error}`, 'error');
        }
    } catch (error) {
        showMessage('register', `网络错误: ${error.message}`, 'error');
    }
});



// ========== 管理员登录 ==========
document.getElementById('adminForm')?.addEventListener('submit', async (e) => {
    e.preventDefault();
    
    const apiKey = document.getElementById('admin-api-key').value;
    
    // 测试 API 密钥
    try {
        const response = await fetch(`${API_BASE}/email`, {
            headers: {
                'X-API-Key': apiKey
            }
        });
        
        if (response.ok) {
            API_KEY = apiKey;
            localStorage.setItem('api_key', apiKey);
            
            // 隐藏登录界面，显示管理后台
            document.querySelector('.container').style.display = 'none';
            document.getElementById('admin-dashboard').style.display = 'block';
            
            // 加载数据
            loadUsers();
            loadFail2ban();
            loadStats();
        } else {
            alert('API 密钥无效');
        }
    } catch (error) {
        alert(`登录失败: ${error.message}`);
    }
});

// ========== 管理后台功能 ==========

// 加载用户列表
async function loadUsers() {
    const usersList = document.getElementById('users-list');
    usersList.innerHTML = '<div class="loading"><div class="spinner"></div>加载中...</div>';
    
    try {
        const response = await fetch(`${API_BASE}/email`, {
            headers: {
                'X-API-Key': API_KEY
            }
        });
        
        if (response.ok) {
            const users = await response.json();
            displayUsers(users);
        } else {
            usersList.innerHTML = '<div class="message error">加载失败</div>';
        }
    } catch (error) {
        usersList.innerHTML = `<div class="message error">错误: ${error.message}</div>`;
    }
}

// 显示用户列表
function displayUsers(users) {
    const usersList = document.getElementById('users-list');
    
    if (!users || users.length === 0) {
        usersList.innerHTML = '<div class="message info">暂无用户</div>';
        return;
    }
    
    let html = '';
    users.forEach(user => {
        html += `
            <div class="data-item">
                <div>
                    <span class="address">${user.address}</span>
                    <span class="quota">${user.quota?.usage || '0'} / ${user.quota?.limit || '∞'}</span>
                    ${user.aliases && user.aliases.length > 0 ? 
                        `<div class="aliases">别名: ${user.aliases.join(', ')}</div>` : ''}
                </div>
                <div class="actions">
                    <button class="edit-btn" onclick="showAliasForm('${user.address}')">管理别名</button>
                    <button class="delete-btn" onclick="deleteUser('${user.address}')">删除</button>
                </div>
            </div>
        `;
    });
    
    usersList.innerHTML = html;
}

// 删除用户
async function deleteUser(email) {
    if (!confirm(`确定要删除 ${email} 吗？`)) return;
    
    try {
        const response = await fetch(`${API_BASE}/email/${encodeURIComponent(email)}`, {
            method: 'DELETE',
            headers: {
                'X-API-Key': API_KEY
            }
        });
        
        if (response.ok) {
            alert('删除成功');
            loadUsers();
        } else {
            alert('删除失败');
        }
    } catch (error) {
        alert(`错误: ${error.message}`);
    }
}

// 显示别名表单
function showAliasForm(email) {
    const aliasInput = prompt(`为 ${email} 添加别名:`, '');
    if (aliasInput) {
        addAlias(email, aliasInput);
    }
}

// 以 addAliasFromForm 函数为例进行修复
async function addAliasFromForm() {
    // 获取主邮箱下拉框的值
    const emailSelect = document.getElementById('alias-main-email');
    if (!emailSelect) {
        alert('找不到主邮箱选择框');
        return;
    }
    const email = emailSelect.value;
    
    // 获取别名输入框的值
    const aliasInput = document.getElementById('alias-address');
    if (!aliasInput) {
        alert('找不到别名输入框');
        return;
    }
    const alias = aliasInput.value.trim();
    
    // **关键检查：确保 email 和 alias 都有值**
    console.log('Debug - 主邮箱:', email, '别名:', alias); // 添加调试输出
    
    if (!email) {
        alert('请选择主邮箱');
        return;
    }
    if (!alias) {
        alert('请输入别名');
        return;
    }
    
    // 调用添加别名的核心函数
    await addAlias(email, alias);
    
    // 清空输入框
    aliasInput.value = '';
}
// 添加别名
async function addAlias(email, alias) {
    if (!email || !alias) {
        console.error('addAlias 收到无效参数:', { email, alias });
        alert('参数错误，无法添加别名');
        return;
    }
    
    try {
        const url = `${API_BASE}/email/${encodeURIComponent(email)}/aliasses`;
        console.log('正在请求:', url, '数据:', alias); // 调试用
        
        const response = await fetch(url, {
            method: 'POST', // **确保是 POST 方法**
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            body: JSON.stringify(alias) // 规范要求直接传字符串
        });        
        
        if (response.ok) {
            alert('别名添加成功');
            loadUsers();
        } else {
            alert('添加失败');
        }
    } catch (error) {
        alert(`错误: ${error.message}`);
    }
}

// 加载别名列表
async function loadAliases() {
    const aliasesList = document.getElementById('aliases-list');
    aliasesList.innerHTML = '<div class="loading"><div class="spinner"></div>加载中...</div>';
    
    try {
        const response = await fetch(`${API_BASE}/email`, {
            headers: {
                'X-API-Key': API_KEY
            }
        });
        
        if (response.ok) {
            const users = await response.json();
            displayAliases(users);
        } else {
            aliasesList.innerHTML = '<div class="message error">加载失败</div>';
        }
    } catch (error) {
        aliasesList.innerHTML = `<div class="message error">错误: ${error.message}</div>`;
    }
}

// 显示别名列表
function displayAliases(users) {
    const aliasesList = document.getElementById('aliases-list');
    
    let html = '';
    users.forEach(user => {
        if (user.aliases && user.aliases.length > 0) {
            user.aliases.forEach(alias => {
                html += `
                    <div class="data-item">
                        <div>
                            <span class="address">${alias}</span>
                            <span class="aliases">→ ${user.address}</span>
                        </div>
                        <div class="actions">
                            <button class="delete-btn" onclick="deleteAlias('${user.address}', '${alias}')">删除</button>
                        </div>
                    </div>
                `;
            });
        }
    });
    
    if (!html) {
        html = '<div class="message info">暂无别名</div>';
    }
    
    aliasesList.innerHTML = html;
}

// 删除别名
async function deleteAlias(email, alias) {
    if (!confirm(`确定要删除别名 ${alias} 吗？`)) return;
    
    try {
        const response = await fetch(`${API_BASE}/email/${encodeURIComponent(email)}/aliasses`, {
            method: 'DELETE',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            body: JSON.stringify(alias)
        });
        
        if (response.ok) {
            alert('别名删除成功');
            loadAliases();
            loadUsers();
        } else {
            alert('删除失败');
        }
    } catch (error) {
        alert(`错误: ${error.message}`);
    }
}

// 加载 Fail2ban 列表
async function loadFail2ban() {
    const fail2banList = document.getElementById('fail2ban-list');
    fail2banList.innerHTML = '<div class="loading"><div class="spinner"></div>加载中...</div>';
    
    try {
        const response = await fetch(`${API_BASE}/fail2ban`, {
            headers: {
                'X-API-Key': API_KEY
            }
        });
        
        if (response.ok) {
            const data = await response.json();
            displayFail2ban(data);
        } else {
            fail2banList.innerHTML = '<div class="message error">加载失败</div>';
        }
    } catch (error) {
        fail2banList.innerHTML = `<div class="message error">错误: ${error.message}</div>`;
    }
}

// 显示 Fail2ban 列表
function displayFail2ban(data) {
    console.log('显示 Fail2ban 数据:', data);
    
    const fail2banList = document.getElementById('fail2ban-list');
    
    // 处理 null 值
    const postfixBans = data.bannedInPostfix || [];
    const dovecotBans = data.bannedInDovecot || [];
    const custom = data.bannedInCustom || [];
    
    let html = `
        <div class="fail2ban-stats">
            <div class="stat-card">
                <h4>📧 Postfix</h4>
                <div class="stat-value">${postfixBans.length}</div>
            </div>
            <div class="stat-card">
                <h4>📨 Dovecot</h4>
                <div class="stat-value">${dovecotBans.length}</div>
            </div>
            <div class="stat-card">
                <h4>🔐 Custom</h4>
                <div class="stat-value">${custom.length}</div>
            </div>
        </div>
    `;
    
    // Dovecot 封禁列表
    html += '<div class="fail2ban-section"><h4>📨 Dovecot 封禁IP:</h4>';
    if (dovecotBans.length > 0) {
        html += '<div class="ip-list">';
        dovecotBans.forEach(ip => {
            html += `
                <div class="ip-item" data-ip="${ip}">
                    <span class="ip-address">${ip}</span>
                    <button class="delete-btn" onclick="unbanIp('${ip}'); event.stopPropagation();">解封</button>
                </div>
            `;
        });
        html += '</div>';
    } else {
        html += '<div class="no-data">暂无封禁IP</div>';
    }
    html += '</div>';
    
    // Postfix 封禁列表
    html += '<div class="fail2ban-section"><h4>📧 Postfix 封禁IP:</h4>';
    if (postfixBans.length > 0) {
        html += '<div class="ip-list">';
        postfixBans.forEach(ip => {
            html += `
                <div class="ip-item" data-ip="${ip}">
                    <span class="ip-address">${ip}</span>
                    <button class="delete-btn" onclick="unbanIp('${ip}'); event.stopPropagation();">解封</button>
                </div>
            `;
        });
        html += '</div>';
    } else {
        html += '<div class="no-data">暂无封禁IP</div>';
    }
    html += '</div>';
    
    // Custom 封禁列表
    html += '<div class="fail2ban-section"><h4>🔐 Custom 封禁IP:</h4>';
    if (custom.length > 0) {
        html += '<div class="ip-list">';
        custom.forEach(ip => {
            html += `
                <div class="ip-item" data-ip="${ip}">
                    <span class="ip-address">${ip}</span>
                    <button class="delete-btn" onclick="unbanIp('${ip}'); event.stopPropagation();">解封</button>
                </div>
            `;
        });
        html += '</div>';
    } else {
        html += '<div class="no-data">暂无封禁IP</div>';
    }
    html += '</div>';
    
    fail2banList.innerHTML = html;
}

// 添加封禁IP
async function addBan() {
    const ip = document.getElementById('ban-ip').value;
    if (!ip) {
        alert('请输入IP地址');
        return;
    }
    
    try {
        const response = await fetch(`${API_BASE}/fail2ban`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-API-Key': API_KEY
            },
            body: JSON.stringify(ip)
        });
        
        if (response.ok) {
            alert('IP封禁成功');
            document.getElementById('ban-ip').value = '';
            loadFail2ban();
        } else {
            alert('封禁失败');
        }
    } catch (error) {
        alert(`错误: ${error.message}`);
    }
}

// 解封IP
// 替换原有的 unbanIp 函数
function unbanIp(ip) {
    // 创建确认对话框
    const dialog = document.createElement('div');
    dialog.className = 'confirm-dialog active';
    dialog.innerHTML = `
        <h3>确认解封</h3>
        <p>确定要解封 IP 地址 <strong>${ip}</strong> 吗？</p>
        <div class="buttons">
            <button class="cancel-btn" onclick="this.closest('.confirm-dialog').remove()">取消</button>
            <button class="confirm-btn" onclick="confirmUnban('${ip}')">确认解封</button>
        </div>
    `;
    document.body.appendChild(dialog);
}

// 确认解封
async function confirmUnban(ip) {
    // 关闭对话框
    document.querySelector('.confirm-dialog')?.remove();
    
    try {
        const response = await fetch(`${API_BASE}/fail2ban/${ip}`, {
            method: 'DELETE',
            headers: {
                'X-API-Key': API_KEY
            }
        });
        
        if (response.ok) {
            alert('✅ IP 解封成功');
            loadFail2ban(); // 刷新列表
        } else {
            const error = await response.text();
            alert(`❌ 解封失败: ${error}`);
        }
    } catch (error) {
        alert(`❌ 网络错误: ${error.message}`);
    }
}

// 加载系统状态
async function loadStats() {
    const statsDiv = document.getElementById('system-stats');
    
    try {
        const response = await fetch(`${API_BASE}/email`, {
            headers: {
                'X-API-Key': API_KEY
            }
        });
        
        if (response.ok) {
            const users = await response.json();
            
            // 计算统计信息
            const totalUsers = users.length;
            let totalAliases = 0;
            let totalStorage = 0;
            
            users.forEach(user => {
                totalAliases += user.aliases?.length || 0;
                // 解析存储使用量
                if (user.quota?.usage) {
                    const usage = parseInt(user.quota.usage.replace('K', ''));
                    totalStorage += usage;
                }
            });
            
            statsDiv.innerHTML = `
                <div class="stat-card">
                    <h4>总用户数</h4>
                    <div class="stat-value">${totalUsers}</div>
                </div>
                <div class="stat-card">
                    <h4>总别名数</h4>
                    <div class="stat-value">${totalAliases}</div>
                </div>
                <div class="stat-card">
                    <h4>总存储使用</h4>
                    <div class="stat-value">${(totalStorage/1024).toFixed(2)} MB</div>
                </div>
                <div class="stat-card">
                    <h4>系统状态</h4>
                    <div class="stat-value">✅ 正常</div>
                </div>
            `;
        }
    } catch (error) {
        statsDiv.innerHTML = `<div class="message error">加载失败: ${error.message}</div>`;
    }
}

// 管理员登出
function logoutAdmin() {
    localStorage.removeItem('api_key');
    API_KEY = '';
    
    document.getElementById('admin-dashboard').style.display = 'none';
    document.querySelector('.container').style.display = 'block';
    
    // 切换到登录标签
    document.querySelectorAll('.tab-btn')[0].click();
}

// 刷新用户列表
function refreshUsers() {
    loadUsers();
}

// 显示消息
function showMessage(tab, text, type) {
    const messageDiv = document.getElementById(`${tab}-message`);
    messageDiv.textContent = text;
    messageDiv.className = `message ${type}`;
    
    // 3秒后自动隐藏成功消息
    if (type === 'success') {
        setTimeout(() => {
            messageDiv.style.display = 'none';
        }, 3000);
    }
}

// 页面加载时检查是否已登录
document.addEventListener('DOMContentLoaded', () => {
    // 从localStorage获取API_KEY
    const storedAPIKey = localStorage.getItem('api_key');
    if (storedAPIKey) {
        API_KEY = storedAPIKey;
        // 自动登录管理员
        document.querySelector('.container').style.display = 'none';
        document.getElementById('admin-dashboard').style.display = 'block';
        loadUsers();
        loadFail2ban();
        loadStats();
    }
});