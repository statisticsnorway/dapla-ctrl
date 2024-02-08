import styles from './sidebar.module.scss'

import { useState, useEffect, useRef } from 'react'
import { Link } from '@statisticsnorway/ssb-component-library'
import { X } from 'react-feather'

interface SidebarModal {
  isOpen: boolean;
  closeSidebar: () => void;

  header?: JSX.Element
  body?: JSX.Element
  button?: JSX.Element
}

const SidebarModal = ({ isOpen, closeSidebar, header, body }: SidebarModal) => {
  const sidebarRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (isOpen && sidebarRef.current && !sidebarRef.current.contains(event.target as Node)) {
        console.log("handleClickOutside moment, closed sidebar");
        closeSidebar();
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    }
  }, [isOpen, closeSidebar]);

  return (
    <div ref={sidebarRef} className={`${styles.sidebar} ${isOpen ? styles.open : ''}`}>
      <div className={styles.header}>
        <button onClick={closeSidebar}>X</button>
      </div>
      <div className={styles.body}>
        {body}
      </div>
    </div>
  );
}

export default SidebarModal
