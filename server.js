import ViteExpress from 'vite-express'
import { createLightship } from 'lightship'
import express from 'express'
import jwt from 'jsonwebtoken'
import jwksClient from 'jwks-rsa'
import { getReasonPhrase } from 'http-status-codes'
import dotenv from 'dotenv'

// TODO: Do a massive cleanup. There are much of the code that can be re-written for reuseability, and some functions
// may not even be required anymore after dapla-team-api-redux changes.

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

// Middleware to protect APi endpoints, requiring Bearer token every time.
async function tokenVerificationMiddleware(req, res, next) {
  try {
    if (!req.headers.authorization || !req.headers.authorization.startsWith('Bearer')) {
      return res.status(401).json({ message: 'No token provided' })
    }

    const token = req.headers.authorization.split('Bearer ')[1]
    const decodedToken = jwt.decode(token, { complete: true })
    if (!decodedToken) {
      return res.status(400).json({ message: 'Invalid token format' })
    }

    const kid = decodedToken.header.kid
    const publicKey = await getPublicKeyFromKeycloak(kid)
    jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
      if (err) {
        return res.status(401).json({ message: 'Invalid token' })
      }
      req.user = decoded
      req.token = token
      next()
    })
  } catch (error) {
    console.error(error)
    res.status(500).json({ message: 'Server error', error: error.message })
  }
}

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
app.get('/api/photo/:principalName', tokenVerificationMiddleware, async (req, res, next) => {
  const accessToken = req.token
  const principalName = req.params.principalName
  const userPhotoUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/photo`
  
  try {
    const photoData = await fetchPhoto(accessToken, userPhotoUrl)

    return res.send({photo: photoData})
  } catch (error) {
    next(error)
  }
})

async function fetchPhoto(token, url, fallbackErrorMessage) {
  const response = await fetch(url, getFetchOptions(token))

  if (!response.ok) {
    throw new Error(fallbackErrorMessage)
  }

  const arrayBuffer = await response.arrayBuffer()
  const photoBuffer = Buffer.from(arrayBuffer)
  return photoBuffer.toString('base64')
}

app.get('/api/teamDetail/:teamUniformName', tokenVerificationMiddleware, async (req, res, next) => {
  try {
    const token = req.token
    const teamUniformName = req.params.teamUniformName
    const teamInfoUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}`
    const teamUsersUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}/users`

    const [teamInfo, teamUsers] = await Promise.all([
      fetchAPIData(token, teamInfoUrl, 'Failed to fetch team info').then(async (teamInfo) => {
        const manager = await fetchTeamManager(token, teamInfo.uniform_name)
        return { ...teamInfo, manager }
      }),
      fetchAPIData(token, teamUsersUrl, 'Failed to fetch team users').then(async (teamUsers) => {
        const resolvedUsers = await fetchTeamUsersWithGroups(token, teamUsers, teamUniformName)
        return { ...teamUsers, _embedded: { users: resolvedUsers } }
      }),
    ])

    // TODO: Implement shared data tab
    res.json({ teamUsers: { teamInfo, teamUsers: teamUsers._embedded.users } })
  } catch (error) {
    next(error)
  }
})

async function fetchTeamManager(token, teamUniformName) {
  const teamManagerUrl = `${DAPLA_TEAM_API_URL}/groups/${teamUniformName}-managers/users`
  return await fetchAPIData(token, teamManagerUrl, 'Failed to fetch team manager')
    .then((teamManager) => {
      return teamManager.count > 0 ? teamManager._embedded.users[0] : managerFallback()
    })
    .catch(() => managerFallback())
}

async function fetchTeamUsersWithGroups(token, teamUsers, teamUniformName) {
  const userPromises = teamUsers._embedded.users.map(async (user) => {
    const userUrl = `${DAPLA_TEAM_API_URL}/users/${user.principal_name}`
    const userGroupsUrl = `${DAPLA_TEAM_API_URL}/users/${user.principal_name}/groups`
    const currentUser = await fetchAPIData(token, userUrl, 'Failed to fetch user')
    const groups = await fetchAPIData(token, userGroupsUrl, 'Failed to fetch groups').catch(() => groupFallback())

    const flattenedGroups = groups._embedded.groups
      .filter((group) => group !== null && group.uniform_name.startsWith(teamUniformName))
      .flatMap((group) => group)

    currentUser.groups = flattenedGroups

    return { ...currentUser }
  })
  return await Promise.all(userPromises)
}

async function fetchAPIData(token, url, fallbackErrorMessage) {
  const response = await fetch(url, getFetchOptions(token))
  const wwwAuthenticate = response.headers.get('www-authenticate')

  if (!response.ok) {
    const { error_description } = wwwAuthenticate
      ? parseWwwAuthenticate(wwwAuthenticate)
      : { error_description: fallbackErrorMessage }
    throw new APIError(error_description, response.status)
  }

  return response.json()
}

function parseWwwAuthenticate(header) {
  const parts = header.split(',')
  const result = {}

  parts.forEach((part) => {
    const [key, value] = part.trim().split('=')
    result[key] = value.replace(/"/g, '')
  })

  return result
}

class APIError extends Error {
  constructor(message, statusCode) {
    super(message)
    this.statusCode = statusCode
  }
}

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

function managerFallback() {
  return {
    display_name: 'Mangler ansvarlig',
    principal_name: 'Mangler ansvarlig',
  }
}

function sectionFallback(uniformName) {
  return {
    uniform_name: uniformName,
    section_name: 'Mangler seksjon',
  }
}

function groupFallback() {
  return { _embedded: { groups: [] }, count: '0' }
}

//const lightship = await createLightship();
// Replace above with below to get liveness and readiness probes when running locally
const lightship = await createLightship({ detectKubernetes: false })

ViteExpress.listen(app, PORT, () => {
  lightship.signalReady()
  console.log(`Server is listening on port ${PORT} ... ${process.env.NODE_ENV}`)
}).on('error', () => {
  lightship.shutdown()
})

lightship.registerShutdownHandler(async () => {
  console.log('Server is shutting down...')
})
