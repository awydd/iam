const PREFIX = import.meta.env.VITE_STORAGE_PREFIX ?? 'iam-web'

export const withPrefix = (key: string): string => {
  return `${PREFIX}_${key}`
}
