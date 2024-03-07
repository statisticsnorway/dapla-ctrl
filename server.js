import ViteExpress from 'vite-express'
import { createLightship } from 'lightship'
import express from 'express'
import { getReasonPhrase } from 'http-status-codes'
import dotenv from 'dotenv'

if (!process.env.VITE_DAPLA_TEAM_API_URL) {
  dotenv.config({ path: './.env.local' })
}

const DAPLA_TEAM_API_URL = process.env.VITE_DAPLA_TEAM_API_URL

const app = express()
const PORT = process.env.PORT || 3000

app.use(express.json())

// DO NOT REMOVE, NECCESSARY FOR FRONTEND
app.get('/api/photo/:principalName', async (req, res, next) => {
  const accessToken = req.headers.authorization.split(' ')[1]
  const principalName = req.params.principalName
  const userPhotoUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/photo`

  try {
    const photoData = await fetchPhoto(accessToken, userPhotoUrl, 'could not fetch photo')

    return res.send({ photo: photoData })
  } catch (error) {
    next(error)
  }
})

async function fetchPhoto(accessToken, url, fallbackErrorMessage) {
  const response = await fetch(url, getFetchOptions(accessToken))

  if (!response.ok) {
    throw new Error(fallbackErrorMessage)
  }

  const arrayBuffer = await response.arrayBuffer()
  const photoBuffer = Buffer.from(arrayBuffer)
  return photoBuffer.toString('base64')
}

app.get('/api/fetch-token', (req, res) => {
  if (!req.headers.authorization || !req.headers.authorization.startsWith('Bearer')) {
    return res.status(401).json({ message: 'No token provided' })
  }

  const token = req.headers.authorization.split('Bearer ')[1]

  res.json({ token })
})

function getFetchOptions(token) {
  return {
    method: 'GET',
    headers: {
      accept: '*/*',
      Authorization: `Bearer ${token}`,
    },
  }
}

// eslint-disable-next-line @typescript-eslint/no-unused-vars
app.use((err, req, res, next) => {
  const statusCode = err.statusCode || 500

  return res.status(statusCode).json({
    success: false,
    error: {
      code: getReasonPhrase(statusCode),
      message: err.message,
    },
  })
})


const lightship = await createLightship()

ViteExpress.listen(app, PORT, () => {
  lightship.signalReady()
  console.log(`Server is listening on port ${PORT} ... ${process.env.NODE_ENV}`)
}).on('error', () => {
  lightship.shutdown()
})

lightship.registerShutdownHandler(async () => {
  console.log('Server is shutting down...')
})
