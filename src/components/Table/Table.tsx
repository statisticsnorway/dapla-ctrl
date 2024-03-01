import styles from './table.module.scss'

import { useEffect, useState } from 'react'
import { useMediaQuery } from 'react-responsive'
import { Title, Dropdown, Input, Text } from '@statisticsnorway/ssb-component-library'

import { ArrowUp, ArrowDown } from 'react-feather'

interface TableProps extends TableData {
  title: string
  dropdownAriaLabel?: string
  dropdownFilterItems?: Array<object>
}

interface TableDesktopViewProps extends TableData {
  activeTab?: string
}
export interface TableData {
  columns: {
    id: string
    label: string
    unsortable?: boolean
  }[]
  data: {
    id: string
    [key: string]: React.ReactNode
  }[]
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

const TableDesktopView = ({ columns, data, activeTab: activeTab }: TableDesktopViewProps) => {
  const defaultState = {
    sortBy: '',
    sortByDirection: 'asc',
  }
  const [sortBy, setSortBy] = useState(defaultState.sortBy)
  const [sortByDirection, setSortByDirection] = useState(defaultState.sortByDirection)

  useEffect(() => {
    setSortBy(defaultState.sortBy)
    setSortByDirection(defaultState.sortByDirection)
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [activeTab])

  const sortTableData = (id: string) => {
    data.sort((a, b) => {
      // Sort by id for the first column;
      const valueA = typeof a[id] === 'object' ? a['id'] : a[id]
      const valueB = typeof b[id] === 'object' ? b['id'] : b[id]

      // Sort by number
      if (typeof valueA === 'number' && typeof valueB === 'number')
        return sortByDirection === 'asc' ? valueA - valueB : valueB - valueA

      // Sort by alphabet
      if (typeof valueA === 'string' && typeof valueB === 'string') {
        if (valueA.toLowerCase() < valueB.toLowerCase()) return sortByDirection === 'asc' ? -1 : 1
        if (valueA.toLowerCase() > valueB.toLowerCase()) return sortByDirection === 'asc' ? 1 : -1
      }
      return 0

      // TODO: Sort by date
    })
  }

  const handleSortBy = (id: string) => {
    setSortBy(id)
    //TODO: Sort by direction should also take account of the cell as well
    setSortByDirection((prevState) => (prevState === 'asc' ? 'desc' : 'asc'))
    sortTableData(id)
  }

  const renderSortByArrow = (selectedColumn: boolean, sortByDirection: string) => {
    if (selectedColumn && sortByDirection === 'asc') return <ArrowDown size={18} />
    return <ArrowUp size={18} />
  }

  return (
    <div className={styles.tableContainer}>
      <table className={styles.table}>
        <thead>
          <tr>
            {columns.map((column) => (
              <th
                key={column.id}
                className={!column.unsortable ? styles.sortableColumn : undefined}
                onClick={!column.unsortable ? () => handleSortBy(column.id) : undefined}
              >
                {!column.unsortable ? (
                  <span>
                    {column.label}
                    {renderSortByArrow(sortBy === column.id, sortByDirection)}
                  </span>
                ) : (
                  column.label
                )}
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
    if (searchFilterKeyword !== '' && data.length) {
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
        <TableDesktopView columns={columns} data={filteredTableData} activeTab={title} /> // Table title changes when switching between tabs
      )}
    </>
  )
}

export default Table
