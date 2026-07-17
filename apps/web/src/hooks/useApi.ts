import { useState, useCallback } from 'react'

export function useApi<T>(fetcher: () => Promise<T>) {
  const [data, setData] = useState<T | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState<string | null>(null)

  const execute = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const result = await fetcher()
      setData(result)
      return result
    } catch (err: unknown) {
      const message = (err && typeof err === 'object' && 'description' in err)
        ? String((err as { description: string }).description)
        : (err instanceof Error ? err.message : 'An error occurred')
      setError(message)
      throw err
    } finally {
      setLoading(false)
    }
  }, [fetcher])

  return { data, loading, error, execute, setData }
}
