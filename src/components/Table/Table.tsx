import styles from './table.module.scss'

import { useMediaQuery } from 'react-responsive'
import { Text } from '@statisticsnorway/ssb-component-library'

export interface TableData {
  columns: {
    id: string
    label: string
  }[]
  data: {
    id: string
    [key: string]: React.ReactNode
  }[]
}

function conditionalStyling(index: number) {
  // Add conditional styling for the first element, then third etc
  return (index + 1) % 2 !== 0 ? styles.greenBackground : undefined
}

const TableMobileView = ({ columns, data }: TableData) => (
  <div className={styles.tableContainerMobile}>
    {data.map((row, index) => {
      return (
        <div key={row.id} className={`${styles.tableMobile} ${conditionalStyling(index)}`}>
          {columns.map((column, index) => (
            <Text small key={column.id}>
              {index !== 0 && <b>{column.label}</b>}
              {row[column.id]}
            </Text>
          ))}
        </div>
      )
    })}
  </div>
)

const TableDesktopView = ({ columns, data }: TableData) => (
  <div className={styles.tableContainer}>
    <table className={styles.table}>
      <thead>
        <tr>
          {columns.map((column) => (
            <th key={column.id}>{column.label}</th>
          ))}
        </tr>
      </thead>
      <tbody>
        {data.map((row, index) => {
          return (
            <tr key={row.id} className={conditionalStyling(index)}>
              {columns.map((column) => (
                <td key={column.id}>{row[column.id]}</td>
              ))}
            </tr>
          )
        })}
      </tbody>
    </table>
  </div>
)

export default function Table({ columns, data }: TableData) {
  const isOnMobile = useMediaQuery({ query: 'screen and (max-width: 767px)' }) // $mobile variable from ssb-component-library
  if (isOnMobile) {
    return <TableMobileView columns={columns} data={data} />
  } else {
    return <TableDesktopView columns={columns} data={data} />
  }
}
