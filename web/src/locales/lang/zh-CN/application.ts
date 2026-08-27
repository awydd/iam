export default {
  searchPlaceholder: '搜索应用名称',
  deleteConfirm: '确定要删除应用 {name} 吗？此操作不可撤销',
  regenerateSecret: '重置密钥',
  regenerateSecretConfirm: '重置后旧密钥将立即失效，确定继续吗？',
  newSecretTitle: '新密钥（仅显示一次，请妥善保存）',
  field: {
    name: '应用名称',
    clientId: 'Client ID',
    redirectUris: '回调地址',
    type: '客户端类型',
    status: '状态',
    accessTokenTtl: '访问有效期（秒）',
    refreshTokenTtl: '刷新有效期（秒）',
  },
  type: {
    confidential: '机密客户端',
    public: '公开客户端',
  },
  status: {
    active: '启用',
    disabled: '禁用',
  },
  form: {
    createTitle: '新建应用',
    editTitle: '编辑应用',
    nameRequired: '请输入应用名称',
    clientIdRequired: '请输入 Client ID',
    redirectUrisRequired: '请至少输入一个回调地址',
    redirectUrisPlaceholder: '每行一个 URL',
  },
}
