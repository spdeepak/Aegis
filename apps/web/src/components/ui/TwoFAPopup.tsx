import { useState, useEffect, useRef, type FormEvent } from 'react'
import { Button } from './Button'
import { Input } from './Input'

interface TwoFAPopupProps {
  open: boolean
  onClose: () => void
  onSubmit: (code: string) => void | Promise<void>
  title?: string
  error?: string
  loading?: boolean
}

export function TwoFAPopup({ open, onClose, onSubmit, title = 'Two-Factor Authentication', error, loading }: TwoFAPopupProps) {
  const [code, setCode] = useState('')
  const dialogRef = useRef<HTMLDialogElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    const dialog = dialogRef.current
    if (!dialog) return
    if (open) {
      dialog.showModal()
      setTimeout(() => inputRef.current?.focus(), 100)
    } else {
      dialog.close()
      setCode('')
    }
  }, [open])

  function handleSubmit(e: FormEvent) {
    e.preventDefault()
    if (code.length === 6) {
      onSubmit(code)
    }
  }

  function handleClose() {
    setCode('')
    onClose()
  }

  return (
    <dialog
      ref={dialogRef}
      onClose={handleClose}
      className="backdrop:bg-black/50 rounded-xl shadow-xl border border-gray-200 p-0 w-full max-w-sm fixed top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2"
    >
      <div className="p-6">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-gray-900">{title}</h3>
          <button onClick={handleClose} className="text-gray-400 hover:text-gray-600 transition-colors">
            <svg className="w-5 h-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>

        <p className="text-sm text-gray-500 mb-4">
          Enter the 6-digit code from your authenticator app.
        </p>

        <form onSubmit={handleSubmit} className="space-y-4">
          {error && (
            <div className="bg-red-50 border border-red-200 text-red-700 text-sm rounded-lg px-4 py-3">
              {error}
            </div>
          )}

          <Input
            ref={inputRef}
            value={code}
            onChange={(e) => setCode(e.target.value.replace(/\D/g, '').slice(0, 6))}
            placeholder="000000"
            maxLength={6}
            required
          />

          <div className="flex justify-end gap-3">
            <Button type="button" variant="secondary" onClick={handleClose} disabled={loading}>
              Cancel
            </Button>
            <Button type="submit" disabled={loading || code.length !== 6}>
              {loading ? 'Verifying...' : 'Verify'}
            </Button>
          </div>
        </form>
      </div>
    </dialog>
  )
}

export function isTwoFARequired(err: unknown): boolean {
  return (
    typeof err === 'object' &&
    err !== null &&
    'status' in err &&
    (err as { status: number }).status === 403 &&
    'errorCode' in err &&
    (err as { errorCode: string }).errorCode === 'JWT0028'
  )
}
