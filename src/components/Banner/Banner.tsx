import styles from './banner.module.scss'
import { Info } from 'react-feather'

const Banner = () => {
  return (
    <div className={styles.header}>
      <Info />
      <p>
        Vi jobber med en ny utgave av Dapla Ctrl som er raskere og har mer funksjonalitet. Prøv den{' '}
        <a href='https://dapla-ctrl-alt-del.intern.ssb.no/'>her</a>.
      </p>
    </div>
  )
}

export default Banner
