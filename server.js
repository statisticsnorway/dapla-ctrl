import ViteExpress from 'vite-express'
import { createLightship } from 'lightship'
import express from 'express'
import { getReasonPhrase } from 'http-status-codes'
import proxy from 'express-http-proxy'
import dotenv from 'dotenv'

if (!process.env.DAPLA_TEAM_API_URL) {
  dotenv.config({ path: './.env.local' })
}

const app = express()
const PORT = process.env.PORT || 3000
const DAPLA_TEAM_API_URL = process.env.DAPLA_TEAM_API_URL || 'https://dapla-team-api-v2.staging-bip-app.ssb.no'

app.use(
  '/api',
  proxy(DAPLA_TEAM_API_URL, {
    proxyReqBodyDecorator: function (bodyContent, srcReq) {
      console.log(`Request Body: ${bodyContent}`)
      return bodyContent
    },
    proxyReqOptDecorator: function (proxyReqOpts, srcReq) {
      console.log(`Request Headers:`, srcReq.headers)
      if (srcReq.body) {
        console.log(`Request Body:`, srcReq.body)
      }
      return proxyReqOpts
    },
    proxyReqPathResolver: function (req) {
      const newPath = req.originalUrl.replace(/^\/api/, '')
      console.log(`Forwarding to: ${DAPLA_TEAM_API_URL}${newPath}`)
      return newPath
    },
    userResDecorator: function (proxyRes, proxyResData, userReq, userRes) {
      console.log(`Response Status: ${proxyRes.statusCode}`)
      console.log(`Response Headers:`, proxyRes.headers)
      return proxyResData
    },
    proxyErrorHandler: function (err, res) {
      console.error('Proxy Error:', err)
      res.status(500).send('Proxy Error')
    },
  })
)

app.use(express.json())

// DO NOT REMOVE, NECCESSARY FOR FRONTEND
app.get('/localApi/photo/:principalName', async (req, res, next) => {
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

//TODO: Remove me once DELETE with proxy is fixed
app.delete('/localApi/groups/:groupUniformName/:userPrincipalName', async (req, res) => {
  const token = req.headers.authorization
  const groupUniformName = req.params.groupUniformName
  const userPrincipalName = req.params.userPrincipalName
  const groupsUrl = `${DAPLA_TEAM_API_URL}/groups/${groupUniformName}/users`

  try {
    const response = await fetch(groupsUrl, {
      method: 'DELETE',
      headers: {
        Accept: '*/*',
        'Content-Type': 'application/json',
        Authorization: token,
      },
      body: JSON.stringify({
        users: [userPrincipalName],
      }),
    })

    if (!response.ok) {
      const err = await response.text()
      res.status(response.status).send(err)
    } else {
      const data = await response.json()
      res.status(response.status).send(data)
    }
  } catch (error) {
    console.log(error)
    res.status(500).send('Internal Server Error')
  }
})

app.get('/localApi/fetch-token', (req, res) => {
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
