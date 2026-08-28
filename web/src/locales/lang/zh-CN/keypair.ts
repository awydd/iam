export default {
  field: {
    kid: 'Key ID',
    algorithm: '算法',
    status: '状态',
    activatedAt: '启用时间',
    retireAt: '停用时间',
  },
  rotate: '轮换密钥',
  rotateConfirm: '将生成新的签名密钥，原密钥进入过渡期，确定继续吗？',
  downgrade: '降级',
  downgradeConfirm: '降级后该密钥将不再用于新签名，仅用于验证旧 token，确定继续吗？',
  retire: '停用',
  retireConfirm: '停用后该密钥将完全失效，确定继续吗？',
}
