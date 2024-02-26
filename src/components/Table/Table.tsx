import styles from './table.module.scss'

import { useEffect, useState } from 'react'
import { useMediaQuery } from 'react-responsive'
import { Title, Dropdown, Input, Text } from '@statisticsnorway/ssb-component-library'

interface TableProps {
  title?: string // TODO: Make a required prop after testing; remove conditional
  columns: TableData['columns']
  data: TableData['data']
}
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
    {data.length ? (
      data.map((row, index) => {
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
      })
    ) : (
      <p>Fant ingen resultater</p>
    )}
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
        {data.length ? (
          data.map((row, index) => {
            return (
              <tr key={row.id} className={conditionalStyling(index)}>
                {columns.map((column) => (
                  <td key={column.id}>{row[column.id]}</td>
                ))}
              </tr>
            )
          })
        ) : (
          <tr>
            <td colSpan={columns.length}>Fant ingen resultater</td>
          </tr>
        )}
      </tbody>
    </table>
  </div>
)

/* TODO:
 * Add alphabetical, numerical etc sorting when row header is clicked
 * Consider making table sort more visible
 */
export default function Table({ title, columns, data }: TableProps) {
  const [searchFilterKeyword, setSearchFilterKeyword] = useState('')
  const [filteredTableData, setFilteredTableData] = useState(data)

  const isOnMobile = useMediaQuery({ query: 'screen and (max-width: 767px)' }) // $mobile variable from ssb-component-library

  useEffect(() => {
    if (searchFilterKeyword !== '') {
      // TODO: Sanitize input. Implement filter on navn row, currently unsearchable since we're passing a React Element
      const filterTableData = data.filter((row) =>
        Object.values(row).toString().toLowerCase().includes(searchFilterKeyword.toLowerCase())
      )
      setFilteredTableData(filterTableData)
    } else {
      setFilteredTableData(data) // Reset filter
    }
  }, [searchFilterKeyword, data])

  const handleChange = (value: string) => {
    setSearchFilterKeyword(value)
  }

  const dropdownFilterItems = [
    {
      title: 'Alle',
      id: 'all',
    },
  ]
  return (
    <>
      <div className={styles.tableTitleContainer}>
        {title && (
          <Title size={2} className={styles.tableTitleWrapper}>
            {title}
          </Title>
        )}
        <div className={styles.tableFilterWrapper}>
          <Dropdown
            className={styles.tableFilterDropdown}
            ariaLabel='' // TODO: Use aria-label since dropdown header is not visible
            selectedItem={dropdownFilterItems[0]}
            items={dropdownFilterItems}
          />
          <Input placeholder='Filtrer liste...' searchField value={searchFilterKeyword} handleChange={handleChange} />
        </div>
      </div>
      {isOnMobile ? (
        <TableMobileView columns={columns} data={filteredTableData} />
      ) : (
        <TableDesktopView columns={columns} data={filteredTableData} />
      )}
    </>
  )
}
