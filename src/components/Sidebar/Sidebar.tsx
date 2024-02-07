import React, { useEffect, useRef } from 'react';
import styles from './Sidebar.module.scss';

interface SidebarProps {
    isOpen: boolean;
    closeSidebar: () => void;
    children?: React.ReactNode;
}

const Sidebar: React.FC<SidebarProps> = ({ isOpen, closeSidebar, children }) => {
  const sidebarRef = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      // Check if the sidebar is open and if the click target is not within the sidebar
      if (isOpen && sidebarRef.current && !sidebarRef.current.contains(event.target as Node)) {
        closeSidebar();
      }
    };

    // Add event listener when the component is mounted or when isOpen changes
    document.addEventListener('mousedown', handleClickOutside);

    // Cleanup the event listener when the component is unmounted or when isOpen changes
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen, closeSidebar]);

  return (
    <div ref={sidebarRef} className={`${styles.sidebar} ${isOpen ? styles.open : ''}`}>
      <button onClick={closeSidebar}>X</button>
      {children}
    </div>
  );
};

export default Sidebar;
