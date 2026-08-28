export default {
  searchPlaceholder: 'Search application name',
  deleteConfirm: 'Are you sure you want to delete application {name}? This cannot be undone.',
  regenerateSecret: 'Regenerate Secret',
  regenerateSecretConfirm: 'The old secret will be invalidated immediately. Continue?',
  newSecretTitle: 'New secret (shown once, please save it)',
  field: {
    name: 'Name',
    clientId: 'Client ID',
    redirectUris: 'Redirect URIs',
    type: 'Client Type',
    status: 'Status',
    accessTokenTtl: 'Access Token TTL (seconds)',
    refreshTokenTtl: 'Refresh Token TTL (seconds)',
  },
  type: {
    confidential: 'Confidential',
    public: 'Public',
  },
  status: {
    active: 'Active',
    disabled: 'Disabled',
  },
  form: {
    createTitle: 'Create Application',
    editTitle: 'Edit Application',
    nameRequired: 'Please enter an application name',
    clientIdRequired: 'Please enter a Client ID',
    redirectUrisRequired: 'Please enter at least one redirect URI',
    redirectUrisPlaceholder: 'One URL per line',
  },
}
