import ViteExpress from 'vite-express'
import { createLightship } from 'lightship'
import express from 'express'
import jwt from 'jsonwebtoken'
import jwksClient from 'jwks-rsa'
import { getReasonPhrase } from 'http-status-codes'
import dotenv from 'dotenv'

if (!process.env.VITE_JWKS_URI) {
  dotenv.config({ path: './.env.local' })
}

const DAPLA_TEAM_API_URL = process.env.VITE_DAPLA_TEAM_API_URL

const app = express()
const PORT = process.env.PORT || 3000

app.use(express.json())

const client = jwksClient({
  jwksUri: process.env.VITE_JWKS_URI,
})

app.post('/api/verify-token', (req, res) => {
  if (!req.headers.authorization.startsWith('Bearer')) {
    return res.status(401).json({ message: 'No token provided' })
  }

  const token = req.headers.authorization.split('Bearer ')[1]

  const decodedToken = jwt.decode(token, { complete: true })
  if (!decodedToken) return res.status(400).json({ message: 'Invalid token format' })

  const kid = decodedToken.header.kid
  getPublicKeyFromKeycloak(kid)
    .then((publicKey) => {
      jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
        if (err) {
          return res.status(401).json({ message: 'Invalid token' })
        }
        res.json({ user: decoded })
      })
    })
    .catch((error) => {
      console.error(error)
      res.status(500).json({ message: 'Server error', error: error.message })
    })
})

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

function getPublicKeyFromKeycloak(kid) {
  return new Promise((resolve, reject) => {
    client.getSigningKey(kid, (err, key) => {
      if (err) {
        reject(err)
        return
      }
      if (!key) {
        reject(new Error('No key found'))
        return
      }
      resolve(key.getPublicKey())
    })
  })
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

app.get('/live', (req, res) => {
  if (lightship.isServerReady()) {
    res.status(200).send({ status: 'ok' });
  } else {
    res.status(503).send({ status: 'error', message: 'Service not ready' });
  }
});

app.get('/health', (req, res) => {
  if (lightship.isServerReady()) {
    res.status(200).send({ status: 'ok' });
  } else {
    res.status(503).send({ status: 'error', message: 'Service not healthy' });
  }
});

ViteExpress.listen(app, PORT, () => {
  lightship.signalReady()
  console.log(`Server is listening on port ${PORT} ... ${process.env.NODE_ENV}`)
}).on('error', () => {
  lightship.shutdown()
})

lightship.registerShutdownHandler(async () => {
  console.log('Server is shutting down...')
})
