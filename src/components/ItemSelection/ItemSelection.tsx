import { useMemo, useState } from 'react'
import { Tag } from '@statisticsnorway/ssb-component-library'
import styles from './itemSelection.module.scss'

export interface Item {
  readonly id: number
  readonly text: string
}

interface ItemSelectionProps {
  readonly header: string
  readonly items: Item[]
  readonly onSelectionChange: (selectedItems: Item[]) => void
}
function ItemSelection({ header, items, onSelectionChange }: ItemSelectionProps) {
  const [selectedItems, setSelectedItems] = useState<Item[]>([])

  const selectedItemIds = useMemo(() => selectedItems.map((item) => item.id), [selectedItems])

  const selectItem = (selectedItem: Item) => {
    const newSelectedItems = selectedItemIds.includes(selectedItem.id)
      ? selectedItems.filter((item) => item.id !== selectedItem.id)
      : [...selectedItems, selectedItem]

    setSelectedItems(newSelectedItems)
    onSelectionChange(newSelectedItems)
  }

  return (
    <div className={styles.rootComponent}>
      <div className={styles.header}>{header}</div>
      <div className={styles.itemSelection}>
        {items.map(({ id, text }) => (
          <Tag
            key={id}
            onClick={(event: Event) => {
              selectItem({ id, text })
              event.preventDefault()
            }}
            className={selectedItemIds.includes(id) ? styles.selectedItem : ''}
          >
            {text}
          </Tag>
        ))}
      </div>
    </div>
  )
}

export default ItemSelection
