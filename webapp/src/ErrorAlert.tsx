import { Alert } from '@mantine/core'
import { TbInfoCircle } from 'react-icons/tb'

type Props = {
  message: string
  title?: string
}

/** Red Mantine alert used to show a failed request/mutation. Renders nothing for an empty message. */
export function ErrorAlert({ message, title = 'Error' }: Props) {
  if (message === '') return null
  return (
    <Alert variant="light" color="red" title={title} icon={<TbInfoCircle />} style={{ marginBottom: 20 }}>
      {message}
    </Alert>
  )
}
