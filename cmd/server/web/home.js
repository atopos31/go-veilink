// 加载客户端列表
function loadClients() {
    fetch('/api/clients/')
        .then(response => response.json())
        .then(clients => {
            const select = document.getElementById('clientSelect');
            select.innerHTML = '<option value="">选择客户端</option>';
            clients.forEach(client => {
                select.innerHTML += `<option value="${client}">${client}</option>`;
            });
        })
        .catch(error => {
            console.error('加载客户端失败:', error);
            showFeedback(false, '加载客户端失败');
        });
}

// 添加轮询状态的函数
function pollClientStatus() {
    const clientId = document.getElementById('clientSelect').value;
    if (!clientId) return;

    const statusBadge = document.getElementById('statusBadge');
    fetch(`/api/clients/${clientId}/online`)
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            return response.json();
        })
        .then(online => {
            statusBadge.className = `badge ${online ? 'badge-success' : 'badge-error'}`;
            statusBadge.textContent = online ? '在线' : '离线';
        })
        .catch(error => {
            console.error('获取客户端状态失败:', error);
            statusBadge.className = 'badge badge-error';
            statusBadge.textContent = '未知';
        });
}

// 修改 loadTunnels 函数，移除原有的状态检查代码
function loadTunnels() {
    const clientId = document.getElementById('clientSelect').value;
    const tunnelCard = document.getElementById('tunnelCard');
    const clientStatus = document.getElementById('clientStatus');
    const showKeyBtn = document.getElementById('showKeyBtn');
    const deleteClientBtn = document.getElementById('deleteClientBtn');

    document.getElementById('addTunnelBtn').disabled = !clientId;
    showKeyBtn.disabled = !clientId;
    deleteClientBtn.disabled = !clientId;

    if (!clientId) {
        tunnelCard.style.display = 'none';
        clientStatus.classList.add('hidden');
        return;
    }

    // 显示客户端状态
    clientStatus.classList.remove('hidden');
    pollClientStatus(); // 立即获取一次状态

    // 显示隧道卡片
    tunnelCard.style.display = 'block';

    // 从专门的隧道 API 获取隧道列表
    fetch(`/api/clients/${clientId}/tunnels`)
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            return response.json();
        })
        .then(tunnels => {
            const tunnelList = document.getElementById('tunnelList');
            tunnelList.innerHTML = '';
            tunnels.forEach(tunnel => {
                const row = document.createElement('tr');
                row.innerHTML = `
                        <td>${tunnel.public_protocol.toUpperCase()}</td>
                        <td>${tunnel.public_ip}:${tunnel.public_port}</td>
                        <td>${tunnel.internal_ip}:${tunnel.internal_port}</td>
                        <td>
                            <div class="badge ${tunnel.encrypt ? 'badge-success' : 'badge-error'}">
                                ${tunnel.encrypt ? '是' : '否'}
                            </div>
                        </td>
                        <td class="flex gap-2 justify-center">
                            <button class="btn btn-xs btn-primary" onclick="editTunnel('${tunnel.uuid}')">编辑</button>
                            <button class="btn btn-xs btn-error" onclick="showDeleteConfirm('${tunnel.uuid}')">删除</button>
                        </td>
                    `;
                tunnelList.appendChild(row);
            });
        })
        .catch(error => {
            console.error('加载隧道列表失败:', error);
            showFeedback(false, error || '加载隧道列表失败');
        });
}

// 客户端相关操作
function openClientModal() {
    document.getElementById('clientModal').showModal();
}

function closeClientModal() {
    document.getElementById('clientModal').close();
    document.getElementById('clientId').value = '';
}

function saveClient() {
    const clientId = document.getElementById('clientId').value;
    if (!clientId) {
        showFeedback(false, '请输入客户端ID');
        return;
    }

    fetch(`/api/clients/${clientId}`, {
        method: 'POST',
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            closeClientModal();
            // 先加载客户端列表
            return fetch('/api/clients/')
                .then(response => response.json())
                .then(clients => {
                    const select = document.getElementById('clientSelect');
                    select.innerHTML = '<option value="">选择客户端</option>';
                    clients.forEach(client => {
                        select.innerHTML += `<option value="${client}">${client}</option>`;
                    });
                    // 选择新添加的客户端
                    select.value = clientId;
                    // 触发 change 事件以加载隧道列表
                    select.dispatchEvent(new Event('change'));
                    showFeedback(true, '客户端添加成功');
                });
        })
        .catch(error => {
            console.error('添加客户端失败:', error);
            showFeedback(false, error || '添加客户端失败');
        });
}

// 隧道相关操作
let editingTunnelId = null;

function openTunnelModal() {
    editingTunnelId = null;
    document.getElementById('tunnelModal').showModal();
}

function closeTunnelModal() {
    document.getElementById('tunnelModal').close();
    document.getElementById('protocol').value = 'tcp';
    document.getElementById('publicIP').value = '0.0.0.0';
    document.getElementById('publicPort').value = '';
    document.getElementById('internalIP').value = '127.0.0.1';
    document.getElementById('internalPort').value = '';
    document.getElementById('encrypt').checked = false;
    editingTunnelId = null;
}

function saveTunnel() {
    const clientId = document.getElementById('clientSelect').value;
    const publicPort = parseInt(document.getElementById('publicPort').value);
    const internalPort = parseInt(document.getElementById('internalPort').value);

    const tunnelData = {
        client_id: clientId,
        public_protocol: document.getElementById('protocol').value,
        public_ip: document.getElementById('publicIP').value,
        public_port: publicPort,
        internal_ip: document.getElementById('internalIP').value,
        internal_port: internalPort,
        encrypt: document.getElementById('encrypt').checked
    };

    // 验证必填字段
    if (!tunnelData.public_ip || !tunnelData.internal_ip) {
        showFeedback(false, '请填写完整的地址信息');
        return;
    }

    // 验证端口号
    if (!publicPort || publicPort < 1 || publicPort > 65535 ||
        !internalPort || internalPort < 1 || internalPort > 65535) {
        showFeedback(false, '请输入有效的端口(1-65535)');
        return;
    }
    console.log(editingTunnelId);
    const method = editingTunnelId ? 'PUT' : 'POST';
    const url = editingTunnelId
        ? `/api/clients/${clientId}/tunnels/${editingTunnelId}`
        : `/api/clients/${clientId}/tunnels`;

    fetch(url, {
        method: method,
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(tunnelData)
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            closeTunnelModal();
            loadTunnels();
            showFeedback(true, method === 'PUT' ? '隧道更新成功' : '隧道添加成功');

        })
        .catch(error => {
            console.error(method === 'PUT' ? '更新隧道失败:' : '添加隧道失败:', error);
            loadTunnels();
            showFeedback(false, error || (method === 'PUT' ? '更新隧道失败' : '添加隧道失败'));
        });

}

function deleteTunnel(btn) {
    if (confirm('确定要删除这个隧道吗？')) {
        // 这里应调用API删除隧道
        btn.closest('tr').remove();
    }
}

// 添加查看客户端密钥功能
function showClientKey() {
    const clientId = document.getElementById('clientSelect').value;
    if (!clientId) return;

    fetch(`/api/clients/${clientId}/key`)
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            return response.text();
        })
        .then(key => {
            document.getElementById('keyModalClientId').textContent = clientId;
            document.getElementById('clientKey').textContent = key;
            document.getElementById('keyModal').showModal();
        })
        .catch(error => {
            console.error('获取密钥失败:', error);
            showFeedback(false, error || '获取密钥失败');
        });
}

// 修改删除确认功能
let tunnelToDelete = null;
let clientToDelete = null;

function showDeleteConfirm(tunnelId) {
    tunnelToDelete = tunnelId;
    clientToDelete = document.getElementById('clientSelect').value;
    document.getElementById('deleteModal').showModal();
}

// 修改删除确认事件监听
document.getElementById('confirmDeleteBtn').addEventListener('click', function () {
    if (tunnelToDelete && clientToDelete) {
        fetch(`/api/clients/${clientToDelete}/tunnels/${tunnelToDelete}`, {
            method: 'DELETE'
        })
            .then(response => {
                if (!response.ok) {
                    return response.text().then(text => Promise.reject(text));
                }
                document.getElementById('deleteModal').close();
                loadTunnels();
                showFeedback(true, '隧道删除成功');
            })
            .catch(error => {
                console.error('删除隧道失败:', error);
                showFeedback(false, error || '删除隧道失败');
            });
        tunnelToDelete = null;
        clientToDelete = null;
    }
});

// 修改编辑隧道功能
function editTunnel(tunnelId) {
    const clientId = document.getElementById('clientSelect').value;
    fetch(`/api/clients/${clientId}/tunnels`)
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            return response.json();
        })
        .then(tunnels => {
            const tunnel = tunnels.find(t => t.uuid === tunnelId);
            if (!tunnel) {
                throw new Error('隧道不存在');
            }

            // 设置编辑状态
            editingTunnelId = tunnelId;

            // 填写表单
            document.getElementById('protocol').value = tunnel.public_protocol;
            document.getElementById('publicIP').value = tunnel.public_ip;
            document.getElementById('publicPort').value = tunnel.public_port;
            document.getElementById('internalIP').value = tunnel.internal_ip;
            document.getElementById('internalPort').value = tunnel.internal_port;
            document.getElementById('encrypt').checked = tunnel.encrypt;

            // 打开模态框
            document.getElementById('tunnelModal').showModal();
        })
        .catch(error => {
            console.error('获取隧道信息失败:', error);
            showFeedback(false, error || '获取隧道信息失��');
        });
}

// 初始化
document.addEventListener('DOMContentLoaded', function () {
    loadClients();
    // 每 5 秒轮询一次状态
    setInterval(pollClientStatus, 2000);
});

// 修改 showFeedback 函数
function showFeedback(success, message) {
    const feedbackContent = document.getElementById('feedbackContent');
    const feedbackModal = document.getElementById('feedbackModal');

    // 设置样式
    if (success) {
        feedbackContent.className = 'alert alert-success shadow-lg';
        feedbackContent.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                <span>${message || '操作成功'}</span>
            `;
    } else {
        feedbackContent.className = 'alert alert-error shadow-lg';
        feedbackContent.innerHTML = `
                <svg xmlns="http://www.w3.org/2000/svg" class="stroke-current shrink-0 h-6 w-6" fill="none" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M10 14l2-2m0 0l2-2m-2 2l-2-2m2 2l2 2m7-2a9 9 0 11-18 0 9 9 0 0118 0z" /></svg>
                <span>${message || '操作失败'}</span>
            `;
    }

    // 显示反馈
    feedbackModal.classList.add('modal-open');

    // 2秒后自动关闭
    setTimeout(() => {
        feedbackModal.classList.remove('modal-open');
    }, 1500);
}

// 添加删除客户端相关函数
function showDeleteClientConfirm() {
    document.getElementById('deleteClientModal').showModal();
}

function deleteClient() {
    const clientId = document.getElementById('clientSelect').value;
    if (!clientId) return;

    fetch(`/api/clients/${clientId}`, {
        method: 'DELETE',
    })
        .then(response => {
            if (!response.ok) {
                return response.text().then(text => Promise.reject(text));
            }
            document.getElementById('deleteClientModal').close();
            document.getElementById('clientSelect').value = '';
            loadClients();
            loadTunnels();
            showFeedback(true, '客户端删除成功');
        })
        .catch(error => {
            console.error('删除客户端失败:', error);
            showFeedback(false, error || '删除客户端失败');
        });
}

// 添加复制功能的 JavaScript
function copyClientKey() {
    const key = document.getElementById('clientKey').textContent;
    navigator.clipboard.writeText(key)
        .then(() => {
            showFeedback(true, '密钥已复制到剪贴板');
        })
        .catch(err => {
            console.error('复制失败:', err);
            showFeedback(false, '复制失败');
        });
}