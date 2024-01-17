import styles from './table.module.scss'

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

export default function Table ({columns, data}: TableData) {
    return (
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
                        // Add conditional styling for first element, then third etc
                        const conditionalStyling = (index + 1) % 2 !== 0 ? styles.conditionalCell : undefined
                        return (
                            <tr key={row.id} className={conditionalStyling}>
                                {columns.map((column) => (
                                    <td key={column.id}>
                                        {row[column.id]}
                                    </td>
                                ))}
                            </tr>)
                    })}
                </tbody>
            </table>
        </div>
    )
}