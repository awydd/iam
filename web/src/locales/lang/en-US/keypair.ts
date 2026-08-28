export default {
  field: {
    kid: 'Key ID',
    algorithm: 'Algorithm',
    status: 'Status',
    activatedAt: 'Activated At',
    retireAt: 'Retired At',
  },
  rotate: 'Rotate Keypair',
  rotateConfirm:
    'This will generate a new signing keypair and move the current one to grace period. Continue?',
  downgrade: 'Downgrade',
  downgradeConfirm:
    'This key will stop signing new tokens and only be used to verify old ones. Continue?',
  retire: 'Retire',
  retireConfirm: 'This key will be fully retired and unusable. Continue?',
}
