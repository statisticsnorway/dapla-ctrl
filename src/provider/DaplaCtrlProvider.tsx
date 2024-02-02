import { FC, ReactNode, createContext, useState } from 'react'

interface BreadcrumbUserProfileDisplayName {
  displayName: string
}

interface DaplaCtrlContextType {
  breadcrumbUserProfileDisplayName: BreadcrumbUserProfileDisplayName | null
  setBreadcrumbUserProfileDisplayName: (
    breadcrumbUserProfileDisplayName: BreadcrumbUserProfileDisplayName | null
  ) => void
}

const DaplaCtrlContext = createContext<DaplaCtrlContextType>({
  breadcrumbUserProfileDisplayName: null,
  setBreadcrumbUserProfileDisplayName: () => {},
})

const DaplaCtrlProvider: FC<{ children: ReactNode }> = ({ children }) => {
  const [breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName] =
    useState<BreadcrumbUserProfileDisplayName | null>(null)

  return (
    <DaplaCtrlContext.Provider value={{ breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName }}>
      {children}
    </DaplaCtrlContext.Provider>
  )
}

export { DaplaCtrlContext, DaplaCtrlProvider }
