import { Text, Link } from '@statisticsnorway/ssb-component-library'

interface FormattedTableColumnProps {
  href: string
  linkText: string
  text: string
}

const FormattedTableColumn = (props: FormattedTableColumnProps) => {
  const { href, linkText, text } = props

  return (
    <>
      <span>
        <Link href={href}>
          <b>{linkText}</b>
        </Link>
      </span>
      <Text>{text}</Text>
    </>
  )
}

export default FormattedTableColumn
