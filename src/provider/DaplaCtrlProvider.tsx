import { FC, ReactNode, createContext, useState } from 'react'

interface BreadcrumbUserProfileDisplayName {
  displayName: string
}

interface BreadcrumbTeamDetailDisplayName {
  displayName: string
}

interface DaplaCtrlContextType {
  breadcrumbUserProfileDisplayName: BreadcrumbUserProfileDisplayName | null
  setBreadcrumbUserProfileDisplayName: (
    breadcrumbUserProfileDisplayName: BreadcrumbUserProfileDisplayName | null
  ) => void

  breadcrumbTeamDetailDisplayName: BreadcrumbTeamDetailDisplayName | null
  setBreadcrumbTeamDetailDisplayName: (
    breadcrumbTeamDetailDisplayName: BreadcrumbTeamDetailDisplayName | null
  ) => void
}

const DaplaCtrlContext = createContext<DaplaCtrlContextType>({
  breadcrumbUserProfileDisplayName: null,
  setBreadcrumbUserProfileDisplayName: () => { },
  breadcrumbTeamDetailDisplayName: null,
  setBreadcrumbTeamDetailDisplayName: () => { },
})

const DaplaCtrlProvider: FC<{ children: ReactNode }> = ({ children }) => {
  const [breadcrumbUserProfileDisplayName, setBreadcrumbUserProfileDisplayName] =
    useState<BreadcrumbUserProfileDisplayName | null>(null)
  const [breadcrumbTeamDetailDisplayName, setBreadcrumbTeamDetailDisplayName] =
    useState<BreadcrumbUserProfileDisplayName | null>(null)

  return (
    <DaplaCtrlContext.Provider value={{
      breadcrumbUserProfileDisplayName,
      setBreadcrumbUserProfileDisplayName,
      breadcrumbTeamDetailDisplayName,
      setBreadcrumbTeamDetailDisplayName
    }}>
      {children}
    </DaplaCtrlContext.Provider>
  )
}

export { DaplaCtrlContext, DaplaCtrlProvider }
