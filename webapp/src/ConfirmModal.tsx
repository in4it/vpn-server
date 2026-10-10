import { Button, Group, Modal, Text } from '@mantine/core'
import { ReactNode } from 'react'

type Props = {
  opened: boolean
  title: string
  confirmLabel: string
  onConfirm: () => void
  onClose: () => void
  loading?: boolean
  children: ReactNode
}

/** Confirmation dialog for destructive actions (delete, regenerate, ...). */
export function ConfirmModal({ opened, title, confirmLabel, onConfirm, onClose, loading, children }: Props) {
  return (
    <Modal opened={opened} onClose={onClose} title={title}>
      <Text size="sm">{children}</Text>
      <Group justify="flex-end" mt="lg">
        <Button variant="default" onClick={onClose}>Cancel</Button>
        <Button color="red" onClick={onConfirm} loading={loading}>{confirmLabel}</Button>
      </Group>
    </Modal>
  )
}
