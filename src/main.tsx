import './index.scss'

import React from 'react'
import ReactDOM from 'react-dom/client'
import App from './App.tsx'

import { BrowserRouter } from 'react-router-dom'
import { DaplaCtrlProvider } from './provider/DaplaCtrlProvider'

ReactDOM.createRoot(document.getElementById('root')!).render(
  <React.StrictMode>
    <BrowserRouter>
      <DaplaCtrlProvider>
        <App />
      </DaplaCtrlProvider>
    </BrowserRouter>
  </React.StrictMode>
)
