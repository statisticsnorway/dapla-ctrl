import styles from './table.module.scss'
import { useMediaQuery } from 'react-responsive'

export interface TableData {
    columns: {
        id: string,
        label: string,
    }[],
    data: {
        id: string,
        [key: string]: React.ReactNode
    }[]
}

function conditionalStyling(index: number) {
    // Add conditional styling for first element, then third etc
    return (index + 1) % 2 !== 0 ? styles.conditionalCell : undefined
}

const TableMobileView = ({columns, data}: TableData) => (
    <div className={styles.tableContainerMobile}>
        {data.map((row, index) => {        
            return (
                <div key={row.id} className={`${styles.tableMobile} ${conditionalStyling(index)}`}>
                    {columns.map((column, index) => 
                        <div key={column.id}>
                            {index !== 0 && <b>{column.label}</b>}
                            {row[column.id]} 
                        </div>
                    )}
                </div>
            )
        })}
    </div>
)

const TableDesktopView = ({columns, data}: TableData) => (
    <div className={styles.tableContainer}>
        <table className={styles.table}>
                <thead>
                <tr>
                    {columns.map((column) => (
                        <th key={column.id}>
                            {column.label}
                        </th>
                    ))}
                </tr>
                </thead>
                <tbody>
                    {data.map((row, index) => {
                        return (
                            <tr key={row.id} className={conditionalStyling(index)}>
                                {columns.map((column) => (
                                    <td key={column.id}>
                                        {row[column.id]}
                                    </td>
                                ))}
                            </tr>
                        )
                    })}
                </tbody>
        </table>
    </div>
)

export default function Table ({columns, data}: TableData) {
    const isMobile = useMediaQuery({ query: '(max-width: 767px)'}) // from ssb-component-library
    
    return (
        <>
           {!isMobile && <TableDesktopView 
                columns={columns} 
                data={data} 
            />}
            {isMobile && <TableMobileView 
                columns={columns} 
                data={data} 
            />}
        </>
    )
}