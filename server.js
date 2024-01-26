import ViteExpress from 'vite-express';
import { createLightship } from 'lightship';
import cache from 'memory-cache';
import express from 'express';
import jwt from 'jsonwebtoken';
import jwksClient from 'jwks-rsa';

import dotenv from 'dotenv';

if (!process.env.VITE_JWKS_URI) {
    dotenv.config({ path: './.env.local' });
}

const DAPLA_TEAM_API_URL = process.env.VITE_DAPLA_TEAM_API_URL;

const app = express();
const PORT = process.env.PORT || 3000;

app.use(express.json());

const client = jwksClient({
    jwksUri: process.env.VITE_JWKS_URI
});

// Middleware to protect APi endpoints, requiring Bearer token every time.
async function tokenVerificationMiddleware(req, res, next) {
    try {
        if (!req.headers.authorization || !req.headers.authorization.startsWith("Bearer")) {
            return res.status(401).json({ message: 'No token provided' });
        }

        const token = req.headers.authorization.split("Bearer ")[1];
        const decodedToken = jwt.decode(token, { complete: true });
        if (!decodedToken) {
            return res.status(400).json({ message: 'Invalid token format' });
        }

        const kid = decodedToken.header.kid;
        const publicKey = await getPublicKeyFromKeycloak(kid);
        jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
            if (err) {
                return res.status(401).json({ message: 'Invalid token' });
            }
            req.user = decoded;
            req.token = token;
            next();
        });
    } catch (error) {
        console.error(error);
        res.status(500).json({ message: 'Server error', error: error.message });
    }
}

app.post('/api/verify-token', (req, res) => {
    if (!req.headers.authorization.startsWith("Bearer")) {
        return res.status(401).json({ message: 'No token provided' });
    }

    const token = req.headers.authorization.split("Bearer ")[1];

    const decodedToken = jwt.decode(token, { complete: true });
    if (!decodedToken) {
        return res.status(400).json({ message: 'Invalid token format' });
    }

    const kid = decodedToken.header.kid;
    getPublicKeyFromKeycloak(kid)
        .then(publicKey => {
            jwt.verify(token, publicKey, { algorithms: ['RS256'] }, (err, decoded) => {
                if (err) {
                    return res.status(401).json({ message: 'Invalid token' });
                }
                res.json({ user: decoded });
            });
        })
        .catch(error => {
            console.error(error);
            res.status(500).json({ message: 'Server error', error: error.message });
        });
});

app.get('/api/teamOverview/allTeams', tokenVerificationMiddleware, async (req, res) => {
    const token = req.token;
    const allteamsUrl = `${DAPLA_TEAM_API_URL}/teams`;

    res.json(await fetchAPIData(token, allteamsUrl, 'Failed to fetch teams').then(teams => getTeamOverviewTeams(token, teams))
        .catch(error => {
            // TODO: Handle errors, 401, 403 etc..
            res.status(500).json({ message: 'Server error', error: error.message });
        }));
});

app.get('/api/teamOverview/myTeams', tokenVerificationMiddleware, async (req, res) => {
    const token = req.token;
    const principalName = req.user.email;
    const myTeamsUrl = `${DAPLA_TEAM_API_URL}/users/${principalName}/teams`;

    res.json(await fetchAPIData(token, myTeamsUrl, 'Failed to fetch my teams').then(teams => getTeamOverviewTeams(token, teams)).catch(error => {
        // TODO: Handle errors, 401, 403 etc..
        res.status(500).json({ message: 'Server error', error: error.message });
    }));
});

async function getTeamOverviewTeams(token, teams) {
    const teamPromises = teams._embedded.teams.map(async (team) => {
        const teamUniformName = team.uniformName;
        const teamUsersUrl = `${DAPLA_TEAM_API_URL}/teams/${teamUniformName}/users`;
        const teamManagerUrl = `${DAPLA_TEAM_API_URL}/groups/${teamUniformName}-managers/users`;

        const [teamUsers, teamManager] = await Promise.all([
            fetchAPIData(token, teamUsersUrl, 'Failed to fetch team users'), //.catch(_ => null),
            fetchAPIData(token, teamManagerUrl, 'Failed to fetch team manager')//.catch(_ => null)
        ]);

        //if (teamUsers == null || teamManager == null) {
        //    console.log(`Failed to fetch team ${teamUniformName}`)
        //    return null;
        //}

        team["teamUserCount"] = teamUsers.count;
        team["manager"] = teamManager._embedded.users[0];
        return { ...team };
    });

    const resolvedTeams = await Promise.all(teamPromises);
    const validTeams = resolvedTeams.filter(team => team !== null);

    teams._embedded.teams = validTeams;
    teams.count = validTeams.length;
    return teams;
}


// TODO: Rework this to use the same logic as the other endpoints
app.get('/api/userProfile', async (req, res) => {
    if (!req.headers.authorization || !req.headers.authorization.startsWith("Bearer")) {
        return res.status(401).json({ message: 'No token provided' });
    }

    try {
        const token = req.headers.authorization.split("Bearer ")[1];
        const jwt = JSON.parse(atob(token.split('.')[1]));

        const cacheKey = `userProfile-${jwt.email}`;
        const cachedUserProfile = cache.get(cacheKey);
        if (cachedUserProfile) {
            return res.json(cachedUserProfile);
        }

        const [userProfile, photo, manager] = await Promise.all([
            fetchUserProfile(token, jwt.email),
            fetchPhoto(token, jwt.email),
            fetchUserManager(token, jwt.email)
        ]);
        const data = { ...userProfile, photo: photo, manager: { ...manager } };
        cache.put(cacheKey, data, 3600000);

        return res.json(data);
    } catch (error) {
        console.error(error);
        return res.status(500).json({ message: 'Server error', error: error.message });
    }
});

async function fetchAPIData(token, url, errorMessage) {
    const response = await fetch(url, getFetchOptions(token));

    if (!response.ok) {
        throw new Error(errorMessage);
    }

    return response.json();
}


async function fetchUserProfile(token, email) {
    const url = `${DAPLA_TEAM_API_URL}/users/${email}`;
    const response = await fetch(url, getFetchOptions(token));

    if (!response.ok) {
        throw new Error('Failed to fetch user profile');
    }

    return response.json();
}

async function fetchUserManager(token) {
    const jwt = JSON.parse(Buffer.from(token.split('.')[1], 'base64').toString());
    const email = jwt.email;
    const url = `${DAPLA_TEAM_API_URL}/users/${email}/manager`;
    const response = await fetch(url, getFetchOptions(token));

    if (!response.ok) {
        throw new Error('Failed to fetch user profile');
    }

    return response.json();
}

async function fetchPhoto(token, email) {
    const url = `${DAPLA_TEAM_API_URL}/users/${email}/photo`;
    const response = await fetch(url, getFetchOptions(token));

    if (!response.ok) {
        throw new Error('Failed to fetch photo');
    }

    const arrayBuffer = await response.arrayBuffer();
    const photoBuffer = Buffer.from(arrayBuffer);
    return photoBuffer.toString('base64');
}

function getFetchOptions(token) {
    return {
        method: "GET",
        headers: {
            "accept": "*/*",
            "Authorization": `Bearer ${token}`,
        }
    };
}

function getPublicKeyFromKeycloak(kid) {
    return new Promise((resolve, reject) => {
        client.getSigningKey(kid, (err, key) => {
            if (err) {
                reject(err);
                return;
            }
            if (!key) {
                reject(new Error('No key found'));
                return;
            }
            resolve(key.getPublicKey());
        });
    });
}

//const lightship = await createLightship();
// Replace above with below to get liveness and readiness probes when running locally
const lightship = await createLightship({ detectKubernetes: false });

ViteExpress.listen(app, PORT, () => {
    lightship.signalReady();
    console.log(
        `Server is listening on port ${PORT} ... ${process.env.NODE_ENV}`
    );
}).on('error', () => {
    lightship.shutdown();
});

lightship.registerShutdownHandler(async () => {
    console.log('Server is shutting down...');
});
