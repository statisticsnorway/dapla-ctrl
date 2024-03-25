import styles from './formattedtablecolumn.module.scss'

import { Link } from '@statisticsnorway/ssb-component-library'

interface FormattedTableColumnProps {
  href?: string
  linkText?: string
  text?: string
}

const FormattedTableColumn = (props: FormattedTableColumnProps) => {
  const { href, linkText, text } = props

  return (
    <>
      <span className={styles.linkTextWrapper}>
        {href && linkText && (
          <Link href={href}>
            <b>{linkText}</b>
          </Link>
        )}
      </span>
      {text && <span className={styles.textWrapper}>{text}</span>}
    </>
  )
}

export default FormattedTableColumn
