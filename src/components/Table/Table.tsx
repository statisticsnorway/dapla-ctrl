import styles from './table.module.scss'

export default function Table () {
  return (
      <table className={styles.table}>
        <thead>
          <tr>
              <th>Lorem ipsum</th>
          </tr>
        </thead>
        <tbody>
            <tr>
                <td>
                  Lorem ipsum
                </td>
            </tr>
        </tbody>
      </table>
  )
}