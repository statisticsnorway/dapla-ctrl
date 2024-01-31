import { FC, ReactNode, createContext, useState } from 'react'

interface DaplaCtrlContextType {
  breadcrumbUserProfileDisplayName: object | null
  setBreadcrumbUserProfileDisplayName: (breadcrumbUserProfileDisplayName: object | null) => void
}

const DaplaCtrlContext = createContext<DaplaCtrlContextType>({
  breadcrumbUserProfileDisplayName: null,
  setBreadcrumbUserProfileDisplayName: () => { },
})

const DaplaCtrlProvider: FC<{ children: ReactNode }> = ({ children }) => {
  const [breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName] = useState<object | null>(null)

  return <DaplaCtrlContext.Provider value={{ breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName }}>{children}</DaplaCtrlContext.Provider>
}

export { DaplaCtrlContext, DaplaCtrlProvider }
