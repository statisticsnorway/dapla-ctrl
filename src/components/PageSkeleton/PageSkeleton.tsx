import { Skeleton } from '@mui/material'
import styles from './pageSkeleton.module.scss'

interface PageSkeletonProps {
  hasDescription?: boolean
  hasTab?: boolean
}

const PageSkeleton = ({ hasDescription, hasTab = true }: PageSkeletonProps) => {
  return (
    <>
      {hasDescription && (
        <Skeleton className={styles.description} variant='rectangular' animation='wave' width={240} height={100} />
      )}
      {hasTab && <Skeleton variant='rectangular' animation='wave' height={60} />}
      <Skeleton variant='text' animation='wave' sx={{ fontSize: '5.5rem' }} width={150} />
      <Skeleton variant='rectangular' animation='wave' height={200} />
    </>
  )
}

export default PageSkeleton
