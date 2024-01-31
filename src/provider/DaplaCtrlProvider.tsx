import { FC, ReactNode, createContext, useState } from 'react'

interface DaplaCtrlContextType {
  data: object | null
  setData: (data: object | null) => void
}

const DaplaCtrlContext = createContext<DaplaCtrlContextType>({
  data: null,
  setData: () => {},
})

const DaplaCtrlProvider: FC<{ children: ReactNode }> = ({ children }) => {
  const [data, setData] = useState<object | null>(null)

  return <DaplaCtrlContext.Provider value={{ data, setData }}>{children}</DaplaCtrlContext.Provider>
}

export { DaplaCtrlContext, DaplaCtrlProvider }
