import styles from './table.module.scss'

import { useEffect, useState } from 'react'
import { useMediaQuery } from 'react-responsive'
import { Title, Dropdown, Input, Text } from '@statisticsnorway/ssb-component-library'

interface TableProps {
  title: string
  dropdownAriaLabel?: string
  dropdownFilterItems?: Array<object>
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
  sortData?: CallableFunction
}

const conditionalStyling = (index: number) => {
  // Add conditional styling for the first element, then third etc
  return (index + 1) % 2 !== 0 ? styles.greenBackground : undefined
}

const NoResultText = () => <p className={styles.noResult}>Fant ingen resultater</p>

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
      <NoResultText />
    )}
  </div>
)

const TableDesktopView = ({ columns, data, sortData }: TableData) => {
  const [sortByDescending, setSortByDescending] = useState(false)

  return (
    <div className={styles.tableContainer}>
      <table className={styles.table}>
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.id}
                onClick={() => {
                  setSortByDescending(!sortByDescending)
                  if (sortData) sortData(column.id, sortByDescending)
                }}
              >
                {column.label}
              </th>
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
              <td colSpan={columns.length}>
                <NoResultText />
              </td>
            </tr>
          )}
        </tbody>
      </table>
    </div>
  )
}

const Table = ({ title, dropdownAriaLabel, dropdownFilterItems, columns, data }: TableProps) => {
  const [searchFilterKeyword, setSearchFilterKeyword] = useState('')
  const [filteredTableData, setFilteredTableData] = useState(data)

  const isOnMobile = useMediaQuery({ query: 'screen and (max-width: 767px)' }) // $mobile variable from ssb-component-library

  useEffect(() => {
    if (searchFilterKeyword !== '') {
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

  // TODO: Currently not working for name tab.
  const handleSort = (id: string, sortByDescending: boolean) => {
    if (filteredTableData) {
      filteredTableData.sort((a, b) => {
        const valueA = a[id]
        const valueB = b[id]

        // Sort by number
        if (typeof valueA === 'number' && typeof valueB === 'number')
          return sortByDescending ? valueA - valueB : valueB - valueA

        // Sort by alphabet
        if (typeof valueA === 'string' && typeof valueB === 'string') {
          if (valueA.toLowerCase() < valueB.toLowerCase()) return sortByDescending ? -1 : 1
          if (valueA.toLowerCase() > valueB.toLowerCase()) return sortByDescending ? 1 : -1
        }
        return 0
      })

      // TODO: Sort by date
    }
  }

  return (
    <>
      <div className={styles.tableTitleContainer}>
        {title && (
          <Title size={2} className={styles.tableTitleWrapper}>
            {title}
          </Title>
        )}
        <div className={styles.tableFilterWrapper}>
          {dropdownFilterItems?.length && (
            <Dropdown
              className={styles.tableFilterDropdown}
              ariaLabel={dropdownAriaLabel}
              selectedItem={dropdownFilterItems[0]}
              items={dropdownFilterItems}
            />
          )}
          <Input placeholder='Filtrer liste...' value={searchFilterKeyword} handleChange={handleChange} searchField />
        </div>
      </div>
      {isOnMobile ? (
        <TableMobileView columns={columns} data={filteredTableData} />
      ) : (
        <TableDesktopView columns={columns} data={filteredTableData} sortData={handleSort} />
      )}
    </>
  )
}

export default Table
